FROM debian:stable-slim

# Install CA certificate for Cassandra SSL connection
# debian already has ca-certificates
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ADD launcher /srv/scaffold/
ADD exec /srv/scaffold/exec
ADD resources /srv/scaffold/resources
ADD app/database/migrate /srv/scaffold/app/database/migrate
ADD app/database/seed /srv/scaffold/resources/seed

WORKDIR /srv/scaffold

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]
CMD ["/srv/scaffold/launcher", "-c", "config.yaml"]
