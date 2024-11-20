windowsbuild:
	$env:GOOS="windows"
	go build  -o freessl.go main.go
linuxbuild:
	$env:GOOS="linux"
	go build  -o freessl_amd64.go main.go
build:
	go build  -o freessl.go main.go
