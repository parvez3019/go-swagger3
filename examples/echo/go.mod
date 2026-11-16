module github.com/parvez3019/go-swagger3/examples/echo

go 1.26.0

require (
	github.com/labstack/echo/v4 v4.16.4
	github.com/parvez3019/go-swagger3 v1.2.0
	github.com/parvez3019/go-swagger3/swagger/echo v0.0.0
)

require (
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace (
	github.com/parvez3019/go-swagger3 => ../..
	github.com/parvez3019/go-swagger3/swagger/echo => ../../swagger/echo
)
