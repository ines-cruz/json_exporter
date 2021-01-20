#!/bin/bash

keepgoing=1
trap '{ echo "sigint"; keepgoing=0; }' SIGINT

cd ..
cd  prometheus-2.22.2.linux-amd64
cp  home/ines/json_exporter/examples/prometheus.yml cerncontainer:/prometheus.yml
sleep 60
./prometheus --web.listen-address="0.0.0.0:9090" &

cd ..
cd json_exporter
cp /home/ines/Downloads/billingcern.json cerncontainer:/billingcern.json

make build

python -m SimpleHTTPServer 8080 &
  BACK_PID=$!
wait $BACK_PID

while (( keepgoing )); do

  ./json_exporter --config.file examples/config.yml &
  BACK_PID=$!
wait $BACK_PID

 curl -k "http://localhost:7979/probe?target=http://localhost:8080/examples/output.json"
 sleep 300 #once per day
done
