# evaluate

This directory provides a minimal local infrastructure for tests and development.

It starts only:
- MySQL
- Redis
- Cassandra

It does not start any application server or run tests.

## Usage

- Start services:

  make env-up

- Stop services:

  make env-down

- Check status:

  make env-ps

- View logs:

  make env-logs
