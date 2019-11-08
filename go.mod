module github.com/ItsJimi/casa

go 1.12

require (
	github.com/anvie/port-scanner v0.0.0-20180225151059-8159197d3770 // indirect
	github.com/cespare/reflex v0.2.0 // indirect
	github.com/getcasa/sdk v0.0.0-20191107193439-b1803b625dc9
	github.com/gorilla/websocket v1.4.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/echo/v4 v4.1.6
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/ogier/pflag v0.0.1 // indirect
	github.com/oklog/ulid/v2 v2.0.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/valyala/fasttemplate v1.1.0 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.12.0
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6
	golang.org/x/net v0.0.0-20191108063844-7e6e90b9ea88 // indirect
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/tools v0.0.0-20191107235519-f7ea15e60b12 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/getcasa/sdk v0.0.0-20191107193439-b1803b625dc9 => ../casa-sdk
