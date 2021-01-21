#!/bin/bash

keepgoing=1
trap '{ echo "sigint"; keepgoing=0; }' SIGINT

cd ..
cd  prometheus-2.22.2.linux-amd64
cp /home/ines/json_exporter/examples/prometheus.yml cerncontainer:/prometheus.yml 

./prometheus --web.listen-address="0.0.0.0:9090" &

cd ..
cd json_exporter
cp /home/ines/Downloads/billingcern.json cerncontainer:/billingcern.json

make build
while kill -0 $BACK_PID ; do
python -m SimpleHTTPServer 8080 &
python -m SimpleHTTPServer 7979 &
 echo "Process is still active..."
    sleep 1
done
while (( keepgoing )); do
while kill -0 $BACK_PID ; do

 ./json_exporter http://localhost:8080/examples/output.json examples/config.yml &
 echo "Process is still active222..."
    sleep 30
done

 curl -k "http://localhost:7979/probe?target=http://localhost:8080/examples/output.json"
 sleep 300
done
