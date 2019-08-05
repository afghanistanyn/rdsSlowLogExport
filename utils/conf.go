package utils

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Conf struct {
	RegionId              string
	AccessKeyId           string
	AccessSecret          string
	DBInstanceId          string
	OutputPrefix          string
	OutputPath            string
	OutputXlsx            string
	ExcludeDBs            string
	ExportTimeOffset      string
	ExportSQLExecTimezone string
	QueryTimesThreshold   int64
}

func NewConf(confpath string, rdsSection string, outputSection string) (conf *Conf, err error) {
	//check confpath

	conf = new(Conf)
	ini_parser := IniParser{}
	if err := ini_parser.Load(confpath); err != nil {
		log.Printf("try load config file[%s] error[%s]\n", confpath, err.Error())
		return conf, err
	}

	conf.RegionId = ini_parser.GetString(rdsSection, "regionId")
	conf.AccessKeyId = ini_parser.GetString(rdsSection, "accessKeyId")
	conf.AccessSecret = ini_parser.GetString(rdsSection, "accessSecret")
	conf.DBInstanceId = ini_parser.GetString(rdsSection, "dbInstanceId")

	conf.OutputPrefix = ini_parser.GetString(outputSection, "output_prefix")
	conf.OutputPath = ini_parser.GetString(outputSection, "output_dir")
	conf.OutputXlsx = ini_parser.GetString(outputSection, "output_xlsx")
	conf.ExcludeDBs = ini_parser.GetString(outputSection, "exclude_dbs")
	conf.ExportTimeOffset = ini_parser.GetString(outputSection, "export_time_offset")
	conf.ExportSQLExecTimezone = ini_parser.GetString(outputSection, "export_sql_exec_timezone")
	queryTimesThreshold := ini_parser.GetString(outputSection, "export_querytimes_threshold")

	if queryTimesThreshold == "" {
		conf.QueryTimesThreshold = 1
	} else {
		conf.QueryTimesThreshold, err = strconv.ParseInt(queryTimesThreshold, 10, 8)
	}

	return conf, nil
}

func (conf *Conf) RemoveExcludeDBRecord() bool {
	if conf.ExcludeDBs != "" {
		return true
	} else {
		return false
	}
}

func (conf *Conf) GetExcludeDbsPattern() *regexp.Regexp {
	return regexp.MustCompile(conf.ExcludeDBs)
}

func (conf *Conf) GetExportTimeOffset() int {
	exportTimeInDays, err := strconv.Atoi(conf.ExportTimeOffset)
	if err != nil {
		return 1
	}
	if exportTimeInDays < 0 {
		exportTimeInDays = exportTimeInDays * -1
	} else if exportTimeInDays > 31 {
		// rds max allowed
		exportTimeInDays = 31
	}
	return exportTimeInDays

}

func (conf *Conf) Convert2CST() bool {
	if strings.ToLower(conf.ExportSQLExecTimezone) == "cst" {
		return true
	} else {
		return false
	}
}

func (conf *Conf) ExportXlsx() bool {
	if strings.ToLower(conf.OutputXlsx) == "true" {
		return true
	} else {
		return false
	}
}
