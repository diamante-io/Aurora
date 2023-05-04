FROM ubuntu:20.04

MAINTAINER Bartek Nowotarski <bartek@diamnet.org>

ARG DIAMNET_CORE_VERSION 
ENV DIAMNET_CORE_VERSION=${DIAMNET_CORE_VERSION:-*}
# to remove tzdata interactive flow
ENV DEBIAN_FRONTEND=noninteractive

ADD dependencies /
RUN ["chmod", "+x", "dependencies"]
RUN /dependencies

ADD diamnet-core.cfg /

ADD start /
RUN ["chmod", "+x", "start"]

ENTRYPOINT ["/start"]
