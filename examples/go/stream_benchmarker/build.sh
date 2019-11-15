#!/bin/bash

set -o errexit

SOURCE=${1:-"benchmark/*.go"}
OUT_NAME=${2:-"benchmark"}

ARCH=('386' 'amd64')

OS=('linux' 'windows' 'darwin')

for GOARCH in "${ARCH[@]}"; do
    for GOOS in "${OS[@]}"; do
        BUILD_DIR="bin/${GOOS}/${GOARCH}"
        FILENAME=${OUT_NAME}
        if [[ "${GOOS}" == "windows" ]]; then
            FILENAME="${OUT_NAME}.exe"
        fi
        mkdir -p ${BUILD_DIR}
        echo "Building ${BUILD_DIR}/${FILENAME}"
        GOARCH=${GOARCH} GOOS=${GOOS} go build -o ${BUILD_DIR}/${FILENAME} ${SOURCE}
    done
done