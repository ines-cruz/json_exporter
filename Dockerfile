
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


EXPOSE 7979 8080 9090 3000
ADD start.sh /
RUN chmod +x /start.sh

#CMD ["/start.sh"]

RUN sleep infinity