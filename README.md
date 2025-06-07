# gecho

A simple HTTP and TCP server to spin up if you want to debug your network connectivity.

## RUN

```sh
go run .
```

```sh
docker run --rm --name gecho -p 8080:8080 -p 8081:8081 allaman/gecho
```

## Configuration

```
go run . <http-addr:http-port> <tcp-addr:tcp-port>
```

## API

http hello world

```sh
curl localhost:8080
```

HTTP echo

```sh
curl -X POST -d "hello world" localhost:8080/echo
```

TCP echo

```sh
echo "hello world" | nc localhost 8081
```
