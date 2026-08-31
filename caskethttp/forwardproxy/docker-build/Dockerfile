FROM alpine:3.6

LABEL description="Docker image for casket+forwardproxy plugin."
LABEL maintainer="SergeyFrolov@colorado.edu"

RUN apk add --no-cache ca-certificates bash curl

RUN curl --fail https://getcasket.com | bash -s http.forwardproxy

COPY gen_casketfile_and_start.sh /bin/

VOLUME /root/.casket

EXPOSE 80 443 2015

ENTRYPOINT /bin/gen_casketfile_and_start.sh
