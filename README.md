json_exporter
========================

A [prometheus](https://prometheus.io/) exporter which scrapes remote JSON by JSONPath.
Forked from the [Prometheus Community](https://github.com/prometheus-community/json_exporter/)


This exporter collects billing data from Google's BigQuery database.
Dashboards are available in Grafana.

After cloning the project you can either run it locally or with Docker.

Documentation can be found [here](https://codimd.web.cern.ch/s/mKlCCS_vs#)

# Cloning and running the project 
```sh
git clone https://github.com/ines-cruz/json_exporter.git
docker-compose build
docker-compose up
```
If you go to http://localhost:3000/dashboards all the dashboards saved will be displayed
In http://localhost:9090/ you can see the data in Prometheus

The advantage of this run over the local one explained below is that you don't need to install Grafana or Prometheus.
Simply clone the project, go into the project's directory and run the above commands.

Grafana will get the data from Prometheus (datasource is already configured within the project in grafana/provisioning/datasources) and display it in a dashboard (dashboard is already configures within the project in grafana/provisioning/dashboards)

Prometheus will get the data from the application.

The application runs a query and obtains data from Google's BigQuery

# Local Run: 
## Build

```sh
make build
```

## Example Usage

```sh
$ cat examples/output.json
{
"values": [
{
"CPUh": 313,
"GPUh": 70,
"amountSpent": 48.65,
"memoryMB": 225360,
"month": "01-2021",
"numberVM": 313,
"projectid": "x-cern"
}
]
}

$ cat examples/config.yml
metrics:
- name: testBQ
  type: object
  path: $.values[*]
  labels:
    projectid: $.projectid
    month: $.month
  values:
    amountSpent: $.amountSpent
    numberVM: $.numberVM
    memoryMB: $.memoryMB
    CPUh: $.CPUh
    GPUh: $.GPUh


For more detailed examples refer Prometheus community [documentation](https://github.com/prometheus-community/json_exporter/blob/master/README.md)

$ python -m SimpleHTTPServer 8080 &
Serving HTTP on 0.0.0.0 port 8080 ...

$ ./json_exporter http://localhost:8080/examples/output.json examples/config.yml &
127.0.0.1 - - [08/Feb/2016 22:44:38] "GET /example/data.json HTTP/1.1" 200 -

$ curl "http://localhost:7979/probe?target=http://localhost:8080/examples/output.json"
127.0.0.1 - - [05/Nov/2020 12:45:43] "GET /example/output.json HTTP/1.1" 200 -
 # HELP testBQ_CPUh testBQ
 # TYPE testBQ_CPUh untypedapp_1         
 testBQ_CPUh{month="01-2021",projectid="x-cern"} 375
 testBQ_CPUh{month="01-2021",projectid="y-cern"} 36972  
# HELP testBQ_GPUh testBQ
# TYPE testBQ_GPUh untyped
 testBQ_GPUh{month="01-2021",projectid="x-cern"} 84
 testBQ_GPUh{month="01-2021",projectid="y-cern"} 103

```

# Docker

```console
$ docker run --rm -it -p 9090:9090 -v $PWD/example/prometheus.yml:/etc/prometheus/prometheus.yml --network host prom/prometheus
```

