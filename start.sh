#!/bin/bash

keepgoing=1
trap '{ echo "sigint"; keepgoing=0; }' SIGINT

cd ..
cd  prometheus-2.22.2.linux-amd64
./prometheus --web.listen-address="0.0.0.0:9090" &


cd ..
cd json_exporter
python -m SimpleHTTPServer 8080 &
python -m SimpleHTTPServer 7979 &
./json_exporter http://localhost:8080/examples/output.json examples/config.yml &

while (( keepgoing )); do
 curl http://localhost:7979/probe?target=http://localhost:8080/examples/output.json
 sleep 3600
done
