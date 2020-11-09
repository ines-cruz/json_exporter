FROM ubuntu:18.04
EXPOSE  7979 8080
# Set working directory as "/"
WORKDIR /
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



ENV GOPATH /root/go
ENV PATH=$PATH:$GOPATH/bin


RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && go get -u github.com/ines-cruz/json_exporter


RUN  git clone https://github.com/ines-cruz/json_exporter.git
WORKDIR json_exporter
RUN make build


CMD ["/bin/bash"]
