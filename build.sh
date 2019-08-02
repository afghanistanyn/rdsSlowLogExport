#!/usr/bin/env bash

set -e

build() {
    go mod tidy
    go mod vendor
    go build -o bin/rdsSlowLogExport
}

package () {
    if [ -f "bin/rdsSlowLogExport" ];then
        tar vczf rdsSlowLogExport.tar.gz zbin/ conf/ doc/ export/
    else
        echo "build faild"
        exit -1
    fi
}

install() {
    tar vxzf rdsSlowLogExport.tar.gz -C /usr/local
}

build
package
install