# Casa (WIP)

Casa is a home automation system to control your home. Link with a [gateway](https://github.com/ItsJimi/casa-gateway) which control devices, you can get datas, send actions or create automations.

## Build

### arm64 (nas synology)
```
sudo env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o casa-gateway *.go
```

### amd64
```
go build -o casa-gateway *.go
```
