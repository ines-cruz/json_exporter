FROM ubuntu:18.04

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    python3-pip \
    nano \
    git \
    jq \
    golang \
    python \
    make


RUN pip3 install \
    pyyaml \
    jsonschema \
    google-cloud \
    google-cloud-resource-manager \
    google-cloud-bigquery \
    google-api \
    google-api-python-client \
    google-auth-oauthlib \
    zeep \
    cs \
    boto3


ENV GOPATH /root/go
ENV PATH=$PATH:$GOPATH/bin


RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN go get -u github.com/ines-cruz/json_exporter


RUN git clone https://github.com/ines-cruz/json_exporter.git
RUN cd json_exporter/  && make build

RUN wget https://github.com/prometheus/prometheus/releases/download/v2.22.0/prometheus-2.22.0.linux-386.tar.gz
RUN tar -xf prometheus-*.tar.gz

RUN cp json_exporter/example/prometheus.yml prometheus-*/prometheus.yml

EXPOSE  7979
ENTRYPOINT ["json_exporter"]
