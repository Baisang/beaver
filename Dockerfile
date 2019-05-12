FROM ubuntu:18.04

RUN sed -i 's/archive.ubuntu.com/mirrors.ocf.berkeley.edu/g' /etc/apt/sources.list \
    && sed -i 's/security.ubuntu.com/mirrors.ocf.berkeley.edu/g' /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y ca-certificates librdkafka-dev --no-install-recommends

RUN mkdir -p /opt/beaver
ADD beaver /opt/beaver/
ADD conf.yaml /opt/beaver/

WORKDIR /opt/beaver
