FROM alpine:3

COPY casket /usr/bin

RUN mkdir /config \
    && mkdir /data \
    && mkdir /www

ENV CASKETPATH=/data

WORKDIR /config

ENTRYPOINT ["casket", "-agree", "-root", "/www"]