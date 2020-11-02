FROM ubuntu:18.04

# ------------------ General Stuff
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
    #python-pip \
    #python2.7 \

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


ADD . /go/src/github.com/ines-cruz/json_exporter

RUN go get -u -d github.com/ines-cruz/json_exporter/


    # ------------------ Clone TS repo and get bash
RUN  echo " cd go/src/github.com/ines-cruz/json_exporter" > ~/.bashrc 




RUN go get -u \
    cloud.google.com/go/ \
    google.golang.org/api/option \
    google.golang.org/api/googleapi \
    google.golang.org/api/bigquery/v2 \
    go.opencensus.io/trace



RUN pip3 install --upgrade requests

EXPOSE  7979
