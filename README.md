# Casa (WIP)

Casa is a home automation system to control your home. Link with a [gateway](https://github.com/ItsJimi/casa-gateway) which control devices, you can get datas, send actions or create automations.

## Build

### arm64 (nas synology)

```
sudo env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o casa-server *.go
```

### amd64

```
go build -o casa-server *.go
```

## Launch

- Start docker (depends of your OS)
- Start database on docker

```
docker-compose up -d
```

- Init database

```
./casa-server init
```

- Start server

```
./casa-server start
```
