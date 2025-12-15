# GONE template project

## HOWTO

copy all files and parse to your project.

copy `config.sample.yaml` to `config.yaml`.

### Add endpoint

add go file to `/app/handlers/endpoints` and add route to `/app/handlers/route.go`

### Add configuration parameter

1. edit `/app/conf/configuration.go`
2. edit `config.yaml`

### Add more program parameter

edit `/app/init.go#FlagParse()`

### Use different config file

```bash
./<execute_file_name> -c <config_file_path>
```

### Build & Run

```bash
go build -v
./<execute_file_name>
```

### Build with target architecture

```bash
# amd64
GOOS=linux GOARCH=amd64 go build

# arm64
GOOS=linux GOARCH=arm64 go build
```

## Project Struct

```
/app
  /build_info              - build info pass in by build script (.gitlab-ci.yml)
  /components              - common libs, for service without any git repo
  /conf                    - configuration file struct folder
  /connector               - external service connectors
    /database              - database connector
    /keyspaces             - cassandra keyspaces connector
    /redis                 - redis connector
  /constant                - constant value
  /daemons                 - daemon process with order
  /database                - database definition
    /migrate               - database migration files
    /seed                  - database seed files
  /errors                  - error struct
  /handlers                - http handler
    /acceptances           - handler acceptance
    /endpoints             - endpoints
    /minortasks            - minortask
  /helpers                 - helper functions
  /models                  - service model
  /repositories            - data access layer
  /services                - business logic services
  /worker                  - background worker
    /internal              - internal worker utilities
    /tasks                 - worker task definitions
  init.go                  - service init procedure define
/docs                      - documentation
  /openapi                 - OpenAPI specification
/resources                 - static resources
  /static
    /static
      /css                 - css files
      /img                 - static image files
      /js                  - javascript files
    favicon.ico            - favicon
    robots.txt             - robots.txt
  /template                - page template files
    /default               - default language templates
    /<lang>                - language specific templates
  /translation             - dictionary files
/tests                     - test files
  /e2e                     - end-to-end tests
  /units                   - unit tests
.gitignore                 - project git ignore file
example.config.yaml        - configuration sample file
Dockerfile                 - docker build sample file
main.go                    - program main entrypoint
```