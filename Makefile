windowsbuild:
	$env:GOOS="windows"
	go build  -o freessl_win64.exe main.go
linuxbuild:
	$env:GOOS="linux"
	go build  -o freessl_amd64.go main.go
arm build:
    $env:GOARCH="arm64"
	$env:GOOS="linux"
	go build  -o freessl_arm64.go main.go
build:
	go build  -o freessl.go main.go
