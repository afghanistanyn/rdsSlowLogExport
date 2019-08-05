package utils

import (
	"github.com/afghanistanyn/json2csv"
	"os"
)

func Save2csv(SlowRecords []map[string]interface{}, filePath string, ignoreHeader bool) (err error) {

	result, err := json2csv.JSON2CSV(SlowRecords)
	if err != nil {
		return err
	}

	// if exist , append to it (for more page)
	//os.RDWR | os.WRONLY is necessary for linux
	fs, _ := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer fs.Close()
	csvwrite := json2csv.NewCSVWriter(fs)
	csvwrite.HeaderStyle = json2csv.SlashStyle
	csvwrite.Comma = rune(';')
	err = csvwrite.WriteCSVWithoutHeader(result, ignoreHeader)
	if err != nil {
		return err
	}
	return nil
}
