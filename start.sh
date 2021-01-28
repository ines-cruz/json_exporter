#!/bin/bash

keepgoing=1
trap '{ echo "sigint"; keepgoing=0; }' SIGINT



cd ..
cd json_exporter

#export GOOGLE_APPLICATION_CREDENTIALS="/go/src/json_exporter/billingcern.json"
#export GOOGLE_APPLICATION_CREDENTIALS="/home/ines/Downloads/billingcern.json"

python -m SimpleHTTPServer 8080 &
python -m SimpleHTTPServer 7979 &

#while (( keepgoing )); do

 #./json_exporter http://localhost:8080/examples/output.json examples/config.yml &


 #curl -k "http://localhost:7979/probe?target=http://localhost:8080/examples/output.json"

 #sleep 3600s
#done
