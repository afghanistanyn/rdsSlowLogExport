package utils

import (
	"bufio"
	"fmt"
	"github.com/tealeg/xlsx"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type MyReg struct {
	Reg *regexp.Regexp
}

func Csv2excel(file2c string) (ok bool, err error) {
	csv, err := os.Open(file2c)
	if err != nil {
		panic(err)
	}
	defer csv.Close()

	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("sheet1")
	if err != nil {
		return false, err
	}

	buf := bufio.NewReader(csv)
	var myReg = new(MyReg)
	myReg.Reg = regexp.MustCompile(`[;]`)

	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		//file read finished.
		if err != nil && err == io.EOF {
			break
		}
		// read going on.
		line = strings.TrimSpace(line)
		myReg.handleLine(sheet, line)
	}

	var TargetNaming = "%s.xlsx"
	baseName := filepath.Base(csv.Name())
	basePath := filepath.Dir(file2c)
	basenameAndExtension := strings.Split(baseName, ".")
	xlsxFileName := fmt.Sprintf(TargetNaming, basenameAndExtension[0])
	file2Save := filepath.Join(basePath, xlsxFileName)
	err = xlsxFile.Save(file2Save)
	if err != nil {
		log.Printf(err.Error())
		return false, err
	}

	return true, nil
}

func (m *MyReg) handleLine(sheet *xlsx.Sheet, line string) {
	items := m.Reg.Split(line, -1)
	row := sheet.AddRow()
	for _, item := range items {
		cell := row.AddCell()
		cell.Value = item
	}
}
