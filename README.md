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
  /conf                    - configuration file struct folder
    configuration.go       - service configuration file parse struct
  /constant                - constant value
  /daemons                 - daemon process with order
  /errors                  - error struct
    /response              - error response
  /handlers                - http handler
    /acceptances           - handler acceptance
    /endpoints             - endpoints
      kkhandlertask.go     - default handlertask with some page render func
    /minortasks            - minortask
    initializer.go         - http channel handler initializer
    route.go               - http endpoint routing procedure
    service.go             - http service procedure
  /models                  - service model
    /api                   - service api model
    /database              - service database model
  /services                - helper/services
  init.go                  - service init procedure define
/components                - common libs, for service without any git repo.
/database                  - database definition
  /seed                    - database seed file
  /table                   - database schema
/example                   - example
/resources                 - static resources
  /static
    /static
      /css                 - css files
      /img                 - static image files
      /js                  - javascript files
    favicon.ico            - favicon
    robots.txt             - robots.txt
  /template                - page template files
    _footer_claim.tmpl     - page footer claim block
    _footer_content.tmpl   - page footer content block
    _header_claim.tmpl     - page header claim block
    _header_content.tmpl   - page header content block
    _main.tmpl             - page main structure definition block
  /translation             - dictionary files
.gitignore                 - project git ignore file
.gitlab-ci.sample.yml      - ci sample file
config.sample.yaml         - configuration sample file
Dockerfile                 - docker build sample file
main.go                    - program main entrypoint
```