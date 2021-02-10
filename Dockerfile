FROM ubuntu:18.04 AS build

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apt-get update && apt-get -y install \
    git \
    curl \
    python3-pip \
    wget \
    nano \
    python \
    golang \
    jq


########## Fetch project, build it and set permissions
RUN go get -u github.com/ines-cruz/json_exporter
WORKDIR  /go/src/
RUN git clone https://github.com/ines-cruz/json_exporter # Using online repo instead of locally cloned repo
COPY start.sh /go/src/json_exporter/
COPY billingcern.json /go/src/json_exporter/
WORKDIR  /go/src/json_exporter/
RUN go build -o json_exporter
RUN chmod 777 -R json_exporter
ENV GOOGLE_APPLICATION_CREDENTIALS /go/src/json_exporter/billingcern.json

########## Prometheus
RUN wget https://github.com/prometheus/prometheus/releases/download/v2.24.1/prometheus-2.24.1.linux-amd64.tar.gz && \
    tar -xf prometheus-*.tar.gz



########## Install Grafana
RUN apt-get install -y adduser libfontconfig1 && \
    wget https://dl.grafana.com/oss/release/grafana_7.3.7_amd64.deb && \
    dpkg -i grafana_7.3.7_amd64.deb

USER root
EXPOSE 7979 8080 9090 3000
########## Dealing with openshift permissions constrains
RUN sed -i 's/grafana:x:102:103::/grafana:x:1008110000:1008110000::/' /etc/passwd
RUN chown -R grafana:grafana /var/lib/grafana /var/log/grafana /usr/share/grafana /etc/grafana /go/src/json_exporter/
RUN chmod -R 777 /var/lib/grafana /var/log/grafana /usr/share/grafana /etc/grafana /go/src/json_exporter/


ENTRYPOINT bash start.sh

