# gecho

A simple HTTP and TCP server to spin up if you want to debug your network connectivity.

## RUN

```sh
go run .
```

```sh
docker run --rm --name gecho -p 8080:8080 -p 8081:8081 allaman/gecho:main # or ghcr.io/allaman/gecho:main
```

## Configuration

```
go run . <http-addr:http-port> <tcp-addr:tcp-port>
```

## API

HTTP Hello World - returns hostname and request headers

```sh
curl localhost:8080
```

HTTP echo - returns payload

```sh
curl -X POST -d "hello world" localhost:8080/echo
```

TCP echo - returns payload

```sh
echo "hello world" | nc localhost 8081
```
