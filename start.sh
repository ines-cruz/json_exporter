#!/bin/bash

keepgoing=1
trap '{ echo "sigint"; keepgoing=0; }' SIGINT

cd ..
cd  prometheus-2.22.2.linux-amd64

chown -R 9090:9090 /.prometheus
./prometheus --web.listen-address="test-cloudtracking.web.cern.ch:9090" &


cd ..
cd json_exporter
python -m SimpleHTTPServer 8080 &
python -m SimpleHTTPServer 7979 &

while (( keepgoing )); do
  ./json_exporter test-cloudtracking.web.cern.ch/examples/output.json examples/config.yml &

 curl test-cloudtracking.web.cern.ch/probe?target=test-cloudtracking.web.cern.ch/examples/output.json
 sleep 3600
done
