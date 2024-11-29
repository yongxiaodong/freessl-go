buildAll: clean buildAmd64 buildArm buildWin

clean:
	rm -f freessl_amd64 freessl_arm freessl_win.exe

buildAmd64:
	export GOOS="linux"
	export GOARCH="amd64"
	go build -o freessl_amd64 main.go

buildArm:
	export GOOS="linux"
	export GOARCH="arm"
	go build -o freessl_arm main.go

buildWin:
	export GOOS="windows"
	export GOARCH="amd64"
	go build -o freessl_win.exe main.go


# if in windows
#windowsbuild:
#	$env:GOOS="windows"
#	go build  -o freessl_win64.exe main.go
#linuxbuild:
#	$env:GOOS="linux"
#	go build  -o freessl_amd64.go main.go
#arm build:
#    $env:GOARCH="arm64"
#	$env:GOOS="linux"
#	go build  -o freessl_arm64.go main.go
#build:
#	go build  -o freessl.go main.go