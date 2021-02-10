#!/bin/bash

# This folder will contain logs for all the services
mkdir logs

wait_for_credentials(){
  sleepTime=60
  while [[ ! -s $GOOGLE_APPLICATION_CREDENTIALS ]] ; do
    echo Waiting for credentials file, retrying in $sleepTime
    sleep $sleepTime
  done
  echo Credentials OK
}

prometheus(){
  sed -i '1800s' /go/src/json_exporter/examples/prometheus.yml # To see stuff on grafana every 10 min
  echo "Run Prometheus..."
  /go/src/json_exporter/prometheus-2.24.1.linux-amd64/prometheus \
               --config.file="/go/src/json_exporter/examples/prometheus.yml" \
               --web.listen-address="0.0.0.0:9090" \
               --storage.tsdb.path=/tmp &>> logs/prometheus_logs & # Runs in the background, logs sent to prometheus_logs...
}

grafana(){
  # datasource
  sed -i 's/prometheus:9090/localhost:9090/' /go/src/json_exporter/grafana/provisioning/datasources/datasource.yml
  cp /go/src/json_exporter/grafana/provisioning/datasources/datasource.yml /usr/share/grafana/conf/provisioning/datasources/
  # dashboard
  cp /go/src/json_exporter/grafana/provisioning/dashboards/dashboard.yml /usr/share/grafana/conf/provisioning/dashboards/
  cp /go/src/json_exporter/grafana/provisioning/dashboards/dashboard.json /etc/grafana/provisioning/dashboards/

  echo "Run Grafana..."
  grafana-server --homepath=/usr/share/grafana &>> logs/grafana_logs & # Runs in the background, logs sent to grafana_logs...
}

app(){

  echo "Run json_exporter"
  python -m SimpleHTTPServer 8080 &>> logs/SimpleHTTPServer_logs & # serves the files to be consumed by json_exporter (not needed in a monolithic, but needed when separating the 3 services)
  ./json_exporter http://localhost:8080/examples/output.json examples/config.yml &>> logs/json_exporter_logs &
  # Both run in the background, logs sent to SimpleHTTPServer_logs and json_exporter_logs...

  while true; do

    echo "Curl to localhost:7979"
    curl -k "http://localhost:7979/probe?target=http://localhost:8080/examples/output.json"
    sleep 900

  done
}

wait_for_credentials
prometheus
grafana
app
