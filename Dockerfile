FROM debian:stable-slim

# Install runtime dependencies:
#   dumb-init — PID 1 process manager (signal forwarding + zombie reaping)
#   ca-certificates — TLS verification for outbound HTTPS
#   curl — optional health-check helper
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        dumb-init \
        ca-certificates \
        curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ADD launcher /srv/app/
ADD resources /srv/app/resources
ADD app/database/migrate /srv/app/app/database/migrate
ADD app/database/seed    /srv/app/app/database/seed

WORKDIR /srv/app

EXPOSE 8080

# dumb-init forwards signals to child processes and reaps zombies,
# preventing orphaned goroutines or stalled containers on SIGTERM.
ENTRYPOINT ["dumb-init", "--"]
CMD ["/srv/app/launcher", "-c", "config.yaml"]
