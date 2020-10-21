FROM ubuntu:18.04

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    nano \
    vim \
    zip \
    git \
    jq


EXPOSE      7979
