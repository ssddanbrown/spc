#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o spc .
docker build -t ssddanbrown/spc .