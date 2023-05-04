FROM ubuntu:focal

ARG VERSION
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y wget apt-transport-https gnupg2 && \
    wget -qO /etc/apt/trusted.gpg.d/SDF.asc https://apt.diamnet.org/SDF.asc && \
    echo "deb https://apt.diamnet.org focal stable" | tee -a /etc/apt/sources.list.d/SDF.list && \
    apt-get update && apt-cache madison diamnet-aurora diamnet-core && apt-get install -y diamnet-aurora=${VERSION} diamnet-core && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /var/log/*.log /var/log/*/*.log

EXPOSE 8000
ENTRYPOINT ["/usr/bin/diamnet-aurora"]
