json_exporter
========================

A [prometheus](https://prometheus.io/) exporter which scrapes remote JSON by JSONPath.
Forked from the [Prometheus Community](https://github.com/prometheus-community/json_exporter/)

# Build

```sh
make build
```

# Example Usage

```sh
$ cat example/output.json
{
"values": {
"sumCores": 44223,
"sumCost": 1972.7199999998902,
"sumGPU": 579,
"sumMem": 14959926,
"sumVM": 3764
}
}

$ cat example/config.yml
metrics:
- name: testBQ
  type: object
  path: $.values
  values:
    spent: $.sumCost     # static value
    vmnum: $.sumVM # dynamic value
    sumMemory: $.sumMem
    sumCores: $.sumCores
    sumGPU: $.sumGPU

For more detailed examples refer Prometheus community [documentation](https://github.com/prometheus-community/json_exporter/blob/master/README.md)
$ python -m SimpleHTTPServer 8080 &
Serving HTTP on 0.0.0.0 port 8080 ...

$ ./json_exporter http://localhost:8080/example/output.json example/config.yml &
INFO[2016-02-08T22:44:38+09:00] metric registered;name:<example_global_value>
INFO[2016-02-08T22:44:38+09:00] metric registered;name:<example_value_boolean>
INFO[2016-02-08T22:44:38+09:00] metric registered;name:<example_value_active>
INFO[2016-02-08T22:44:38+09:00] metric registered;name:<example_value_count>
127.0.0.1 - - [08/Feb/2016 22:44:38] "GET /example/data.json HTTP/1.1" 200 -


$ curl "http://localhost:7979/probe?target=http://localhost:8080/example/output.json"
127.0.0.1 - - [05/Nov/2020 12:45:43] "GET /example/output.json HTTP/1.1" 200 -
# HELP testBQ_spent testBQ
# TYPE testBQ_spent untyped
testBQ_spent 1972.7199999998902
# HELP testBQ_sumCores testBQ
# TYPE testBQ_sumCores untyped
testBQ_sumCores 44223
# HELP testBQ_sumGPU testBQ
# TYPE testBQ_sumGPU untyped
testBQ_sumGPU 579
# HELP testBQ_sumMemory testBQ
# TYPE testBQ_sumMemory untyped
testBQ_sumMemory 1.4959926e+07
# HELP testBQ_vmnum testBQ
# TYPE testBQ_vmnum untyped
testBQ_vmnum 3764
```

# Docker

```console
$ docker run --rm -it -p 9090:9090 -v $PWD/example/prometheus.yml:/etc/prometheus/prometheus.yml --network host prom/prometheus
```


#Prometheus
```console

$ cd prometheus
$ ./prometheus --web.listen-address="0.0.0.0:9090"

$ cat prometheus.yml

global:
  scrape_interval: 15s
  scrape_timeout: 10s
  evaluation_interval: 15s
alerting:
  alertmanagers:
  - static_configs:
    - targets: []
    scheme: http
    timeout: 10s
    api_version: v1
rule_files:
- /etc/prometheus/alerts.yml
scrape_configs:
- job_name: prometheus
  honor_timestamps: true
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets:
    - localhost:9090
    - localhost:7979
- job_name: json_exporter
  honor_timestamps: true
  scrape_interval: 5s
  metrics_path: /probe
  scheme: http
  static_configs:
  - targets:
    - localhost:7979
  params:
    target: ["http://localhost:8080/example/output.json"]
  relabel_configs:
    - source_labels: [__param_target]
      target_label: endpoint
      action: replace

```

# See Also
- [kawamuray/jsonpath](https://github.com/kawamuray/jsonpath#path-syntax) : For syntax reference of JSONPath.
  Originally forked from nicksardo/jsonpath(now is https://github.com/NodePrime/jsonpath).
