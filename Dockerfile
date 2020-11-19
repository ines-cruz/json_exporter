
FROM golang:1.15 AS build

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apt-get update && apt-get -y install \
    git \
    curl \
    python3-pip \
    wget \
    nano \
    python

WORKDIR /go/src/json_exporter/

RUN go get -u github.com/ines-cruz/json_exporter


# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o json_exporter


RUN chmod 777 -R json_exporter
WORKDIR /go/src/
RUN wget https://github.com/prometheus/prometheus/releases/download/v2.22.0/prometheus-2.22.0.linux-386.tar.gz && tar -xf prometheus-*.tar.gz


RUN cp  json_exporter/examples/prometheus.yml prometheus-*/prometheus.yml
WORKDIR json_exporter
RUN make build

WORKDIR prometheus-2.22.0.linux-386

EXPOSE 9090 8080 7979
USER 1001
CMD ./prometheus --web.listen-address="test-cloudtracking.web.cern.ch:9090" &

ENTRYPOINT ["json_exporter"]

CMD python -m SimpleHTTPServer 8080 &

CMD  ./json_exporter test-cloudtracking.web.cern.ch:8080/examples/output.json examples/config.yml &

CMD curl test-cloudtracking.web.cern.ch/probe?target=test-cloudtracking.web.cern.ch:8080/examples/output.json
