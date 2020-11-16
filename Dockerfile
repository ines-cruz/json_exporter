

FROM golang:1.15 AS build

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apt install git
WORKDIR /go/src/json_exporter/

RUN go get -u github.com/ines-cruz/json_exporter


EXPOSE 7979 8080 9090
# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o json_exporter

# This results in a single layer image
FROM ubuntu:18.04

RUN apt-get update && apt-get -y install \
    curl \
    python3-pip \
    wget \
    python

COPY --from=build /go/src/json_exporter/ /json_exporter


RUN chmod 777 -R /json_exporter
RUN wget https://github.com/prometheus/prometheus/releases/download/v2.22.0/prometheus-2.22.0.linux-386.tar.gz


RUN tar -xf prometheus-*.tar.gz



RUN cp  json_exporter/examples/prometheus.yml prometheus-*/prometheus.yml

USER 1001

CMD ["./prometheus" , "--web.listen-address="0.0.0.0:9090"" ,"&"]

CMD ["python" , "-m", "SimpleHTTPServer", "8080", "&"]

CMD [ "/json_exporter", "http://localhost:8080/example/output.json example/config.yml", "&  "]

ENTRYPOINT ["curl"]
#CMD curl http://cloud-tracking.web.cern.ch/probe?target=http://localhost:8080/examples/output.json
