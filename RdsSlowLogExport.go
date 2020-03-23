package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/afghanistanyn/rdsSlowLogExport/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"log"
)

const (
	// must be one of {30,50,100}
	pageSize      = 50
	rdsSection    = "rds"
	outputSection = "output"
)

var (
	help     bool
	confPath string
	debug    bool
)

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&debug, "d", false, "show verbose info")
	flag.StringVar(&confPath, "c", "/usr/local/etc/rdsconf.ini", "set configuration `file`")
}

func main() {

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	conf, err := utils.NewConf(confPath, rdsSection, outputSection)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	if conf.RegionId == "" || conf.AccessKeyId == "" || conf.AccessSecret == "" || conf.DBInstanceId == "" {
		log.Println("config err , conf at rds selection can't be blank")
		os.Exit(-1)
	}

	removeExcludeDBRecord := conf.RemoveExcludeDBRecord()
	excludeDbsPattern := conf.GetExcludeDbsPattern()
	exportTimeInDays := conf.GetExportTimeOffset()
	convert2CST := conf.Convert2CST()

	client, err := rds.NewClientWithAccessKey(conf.RegionId, conf.AccessKeyId, conf.AccessSecret)
	request := rds.CreateDescribeSlowLogRecordsRequest()
	request.Scheme = "https"
	request.DBInstanceId = conf.DBInstanceId
	request.PageSize = requests.NewInteger(pageSize)

	func() {
		utcFormat := "2006-01-02T15:04"
		var exportTimeOffsetInDays int
		exportTimeOffsetInDays = exportTimeInDays - 1
		d := time.Duration(time.Duration(exportTimeOffsetInDays) * time.Hour * -24)

		loc, _ := time.LoadLocation("Asia/Shanghai")
		now := time.Now().In(loc)
		year, month, day := now.Date()

		cst2utc, _ := time.ParseDuration("-8h")
		request.StartTime = time.Date(year, month, day, 0, 0, 0, 0, now.Location()).Add(d).Add(cst2utc).Format(utcFormat) + "Z"
		request.EndTime = time.Date(year, month, day, 23, 59, 59, 0, now.Location()).Add(cst2utc).Format(utcFormat) + "Z"
		if debug {
			log.Printf("Export slowlog StartTime: %s (UTC mode)\n", request.StartTime)
			log.Printf("Export slowlog EndTime: %s (UTC mode)\n", request.EndTime)
		}
	}()

	response, err := client.DescribeSlowLogRecords(request)
	if err != nil {
		log.Print(err.Error())
	}

	if debug {
		log.Printf("Try to handle page %s \n", request.PageNumber)
		log.Printf("Request Params: %s\n", request.GetQueryParams())
	}

	var globalHaveHeader = false
	SQLSlowRecords := response.Items.SQLSlowRecord
	SlowRecords := handleRecords(SQLSlowRecords, excludeDbsPattern, removeExcludeDBRecord, convert2CST, conf.QueryTimesThreshold)

	var dateFormat = "2006-01-02"
	var outputFileFormat = "%s.csv"
	var outputFileName = conf.OutputPrefix + time.Now().Format(dateFormat)
	outputFilePath := filepath.Join(conf.OutputPath, fmt.Sprintf(outputFileFormat, outputFileName))

	//mkdir outputPath
	_ = os.Mkdir(conf.OutputPath, os.ModeDir)

	if len(SlowRecords) != 0 {
		globalHaveHeader = true
		err = utils.Save2csv(SlowRecords, outputFilePath, false)
		if err != nil {
			log.Fatalf("err occur when write csv: %s\n", err)
			os.Exit(-1)
		}
	} else {
		//the first page is blank
		//write blank file
		log.Println("the first page is blank")

	}

	var fetchCount int
	func() {
		totalRecordCount := response.TotalRecordCount
		if totalRecordCount <= pageSize {
			fetchCount = 1
		} else {
			c := totalRecordCount % pageSize
			if c == 0 {
				fetchCount = totalRecordCount / pageSize
			} else {
				fetchCount = totalRecordCount/pageSize + 1
			}
		}

		if debug {
			log.Printf("Total Records: %d, PageSize: %d , fetchCount: %d \n", totalRecordCount, pageSize, fetchCount)
		}
	}()

	//get left pages, the first page has beed processed
	for i := 1; i <= (fetchCount - 1); i++ {
		request.PageNumber = requests.NewInteger(i + 1)
		if debug {
			log.Printf("Try to handle page %s \n", request.PageNumber)
			log.Printf("Request Params: %s\n", request.GetQueryParams())
		}
		response, err = client.DescribeSlowLogRecords(request)
		if err != nil {
			log.Print(err.Error())
		}
		SQLSlowRecords := response.Items.SQLSlowRecord
		//the first page , did not ignore header
		SlowRecords = handleRecords(SQLSlowRecords, excludeDbsPattern, removeExcludeDBRecord, convert2CST, conf.QueryTimesThreshold)

		if len(SlowRecords) == 0 {
			continue
		}

		//如果前面均未添加头部, 则添加表头 , ignoreHeader=false
		if globalHaveHeader {
			err := utils.Save2csv(SlowRecords, outputFilePath, true)
			if err != nil {
				log.Fatalf("err occurd when write csv: %s\n", err)
				os.Exit(-1)
			}
		} else {
			err := utils.Save2csv(SlowRecords, outputFilePath, false)
			if err != nil {
				log.Fatalf("err occurd when write csv: %s\n", err)
				os.Exit(-1)
			}
			globalHaveHeader = true
		}
	}

	//check csv exist

	if conf.ExportXlsx() && IsExist(outputFilePath) {
		_, err = utils.Csv2excel(outputFilePath)
		if err != nil {
			log.Printf("convert csv to excel error %s \n", err)
			os.Exit(-1)
		}
	} else {
		log.Println("No file generated")
	}

	if debug {
		log.Println("Export Slow SQL Done")
	}
}

func handleRecords(records []rds.SQLSlowRecord, excludeDbsPattern *regexp.Regexp, removeExcludeDBRecord bool, convert2CST bool, queryTimesThreshold int64) []map[string]interface{} {
	//remove record for exclude dbs
	if removeExcludeDBRecord {
		if debug {
			log.Println("Try to remove the Records match exclude_dbs")
		}

		for i := len(records) - 1; i >= 0; i-- {
			if excludeDbsPattern.MatchString(records[i].DBName) {
				records = append(records[:i], records[i+1:]...)
			}
		}

		if debug {
			log.Printf("%d records left after remove \n", len(records))
		}
	}

	if debug {
		log.Printf("Try to remove the Records that QueryTimes less than %d sec\n", queryTimesThreshold)
	}
	//remove records that QueryTimes less than `queryTimesThreshold` sec
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].QueryTimes < queryTimesThreshold {
			records = append(records[:i], records[i+1:]...)
		}
	}

	if debug {
		log.Printf("%d records left after remove \n", len(records))
	}

	if convert2CST {
		if debug {
			log.Println("Try to Convert Record ExecutionStartTime to CST")
		}
		cstFormat := "2006-01-02T15:04:05 CST"
		for i, Record := range records {
			//cstLoc , _ := time.LoadLocation("Asia/Shanghai")
			executionStartTimeInUTC, _ := time.Parse(time.RFC3339, Record.ExecutionStartTime)
			records[i].ExecutionStartTime = executionStartTimeInUTC.Add(8 * time.Hour).Format(cstFormat)
		}
	}

	sqlrecordsBytes, _ := json.Marshal(records)
	//delete \n from sqltext
	var slowrecords = string(sqlrecordsBytes)
	slowrecords = strings.Replace(slowrecords, "\\n", "", -1)

	//var SlowRecord map[string]interface{}
	var SlowRecords []map[string]interface{}
	if err := json.Unmarshal([]byte(slowrecords), &SlowRecords); err != nil {
		log.Println(err)
	}
	// the json data to convert must be a map or slice , can't be stuct
	// so convert SQLSlowRecords(struct) to SlowRecords ([]map)
	return SlowRecords
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
