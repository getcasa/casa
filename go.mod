module github.com/ItsJimi/casa

go 1.12

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/getcasa/sdk v0.0.0-20191111095811-ef20c3aec8c1
	github.com/gorilla/websocket v1.4.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/oklog/ulid/v2 v2.0.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/valyala/fasttemplate v1.1.0 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.12.0
	golang.org/x/crypto v0.0.0-20191108234033-bd318be0434a
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5 // indirect
	golang.org/x/sys v0.0.0-20191110163157-d32e6e3b99c4 // indirect
	golang.org/x/tools v0.0.0-20191111134443-9f4e7946603e // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/getcasa/sdk v0.0.0-20191107193439-b1803b625dc9 => ../casa-sdk
