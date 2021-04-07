FROM debian:stable-slim

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ADD go-scaffold /
ADD config.yaml /
ADD resources /resources
ENV GOGC=100

EXPOSE 8080

CMD ["/go-scaffold"]