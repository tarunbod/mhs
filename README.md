# mhs - Mock HTTP Server

Simple command-line tool to spin up an HTTP server that serves status codes
with empty responses and/or files on the given paths.

## Installation

```
$ go install github.com/tarunbod/mhs@latest
```

## Docs

```
$ mhs --help
USAGE:
  mhs [options] [/request-path response-template]...

OPTIONS:
  -p int
        port to serve on (default 8080)

RESPONSE TEMPLATES
A response template can either be a status code, a path to an existing directory, or a file. It is assumed to be a file path if it is not a valid status code and does not exist as a directory.

EXAMPLES
Serve current directory on port 8080:
  mhs -p 8081
Serve current directory on port 8081:
  mhs -p 8081
Serve 200s from /ok and 500s from /error:
  mhs /ok 200 /error 500
Serve 200s from /status and the "/tmp" directory from /files:
  mhs /status 200 /files /tmp
```
