
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
    golang



RUN go get -u  github.com/ines-cruz/json_exporter

WORKDIR  /go/src/json_exporter/

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o json_exporter

RUN chmod 777 -R json_exporter
#Prometheus
RUN cd ~ && cd /go/src &&  wget https://github.com/prometheus/prometheus/releases/download/v2.22.2/prometheus-2.22.2.linux-amd64.tar.gz && tar -xf prometheus-*.tar.gz

RUN cd . &&  cp  /go/src/json_exporter/examples/prometheus.yml /go/src/prometheus-2.22.2.linux-amd64/prometheus.yml

#Grafana
RUN apt-get install -y adduser libfontconfig1 && wget https://dl.grafana.com/oss/release/grafana_7.3.7_amd64.deb &&  dpkg -i grafana_7.3.7_amd64.deb

EXPOSE 7979 8080 9090 3000
ADD start.sh /
RUN chmod +x /start.sh
### Containers should NOT run as root as a good practice
USER 10001
CMD ["/start.sh"]
