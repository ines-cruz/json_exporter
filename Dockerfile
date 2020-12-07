
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

USER root


RUN go get -u  github.com/ines-cruz/json_exporter

WORKDIR  /go/src/json_exporter/

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o json_exporter

RUN chmod 777 -R json_exporter
RUN cd ~ && cd /go/src &&  wget https://github.com/prometheus/prometheus/releases/download/v2.22.2/prometheus-2.22.2.linux-amd64.tar.gz && tar -xf prometheus-*.tar.gz


RUN cd . &&  cp  /go/src/json_exporter/examples/prometheus.yml /go/src/prometheus-2.22.2.linux-amd64/prometheus.yml


EXPOSE 9090 8080 7979

ADD start.sh /
RUN chmod +x /start.sh

CMD ["/start.sh"]
