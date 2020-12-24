FROM golang:1.12-stretch

LABEL maintainer="Alex Cordeiro <alexc@diamnet.org>"

EXPOSE 5432
EXPOSE 8000

COPY docker/dependencies /
RUN ["chmod", "+x", "/dependencies"]
RUN /dependencies


COPY docker/setup /
RUN ["chmod", "+x", "/setup"]
RUN /setup

COPY docker/conf /opt/diamnet/conf
RUN crontab -u diamnet /opt/diamnet/conf/crontab.txt

COPY docker/start /
RUN ["chmod", "+x", "/start"]
ENTRYPOINT ["/start"]
