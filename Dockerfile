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
RUN go get -u  github.com/ines-cruz/json_exporter
WORKDIR  /go/src/json_exporter/
# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o json_exporter
RUN chmod 777 -R json_exporter

########## Prometheus
RUN cd /go/src && \
    wget https://github.com/prometheus/prometheus/releases/download/v2.24.1/prometheus-2.24.1.linux-amd64.tar.gz && \
    tar -xf prometheus-*.tar.gz

RUN cp /go/src/json_exporter/examples/prometheus.yml /go/src/prometheus-2.24.1.linux-amd64/prometheus.yml


########## Install Grafana
RUN apt-get install -y adduser libfontconfig1 && \
    wget https://dl.grafana.com/oss/release/grafana_7.3.7_amd64.deb && \
    dpkg -i grafana_7.3.7_amd64.deb

USER root
EXPOSE 7979 8080 9090 3000
ADD start.sh /
########## Dealing with openshift permissions constrains
RUN sed -i 's/grafana:x:102:103::/grafana:x:1000920000:1000920000::/' /etc/passwd
RUN chown -R grafana:grafana /var/lib/grafana /var/log/grafana /usr/share/grafana /etc/grafana /go/src/json_exporter/
RUN chmod -R 777 /var/lib/grafana /var/log/grafana /usr/share/grafana /etc/grafana /go/src/json_exporter/

#ENTRYPOINT sleep infinity
ENTRYPOINT echo "About to run Prometheus..." && \
          cd /go/src/prometheus-2.24.1.linux-amd64 && \
          ./prometheus --web.listen-address="0.0.0.0:9090" --storage.tsdb.path=/tmp & \
          echo "About to run Grafana..." && \
          grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/grafana.ini & \
          echo "About to run start.sh..." && \
          /start.sh
