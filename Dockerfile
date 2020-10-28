FROM ubuntu:18.04

# ------------------ General Stuff
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    awscli \
    python3-pip \
    nano \
    vim \
    zip \
    git \
    snapd \
    net-tools \
    jq \
    golang \
    python
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
    #google-cloud-api \
    #google-api-client \


ENV GOPATH /root/go

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"



    # ------------------ Clone TS repo and get bash
RUN cd ~; cd go/src ; echo "git clone -q https://github.com/ines-cruz/json_exporter ; cd json_exporter" > ~/.bashrc



RUN go get -u \
    cloud.google.com/go/ \
    google.golang.org/api/option \
    google.golang.org/api/googleapi \
    google.golang.org/api/bigquery/v2 \
    go.opencensus.io/trace






RUN pip3 install --upgrade requests

EXPOSE  7979
