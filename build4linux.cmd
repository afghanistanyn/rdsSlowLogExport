
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go mod tidy
go mod vendor
go build -v -o bin/rdsSlowLogExport

dir bin/
