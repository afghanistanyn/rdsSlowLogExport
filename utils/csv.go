package utils

import (
	"github.com/afghanistanyn/json2csv"
	"log"
	"os"
)

func Save2csv(SlowRecords []map[string]interface{}, filePath string, ignoreHeader bool) (ok bool, err error) {

	result, err := json2csv.JSON2CSV(SlowRecords)
	if err != nil {
		log.Println(err)
	}

	// if exist , append to it (for more page)
	fs, _ := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0644)
	defer fs.Close()
	csvwrite := json2csv.NewCSVWriter(fs)
	csvwrite.HeaderStyle = json2csv.SlashStyle
	csvwrite.Comma = rune(';')
	err = csvwrite.WriteCSVWithoutHeader(result, ignoreHeader)
	if err != nil {
		return false, err
	}
	return true, nil
}
