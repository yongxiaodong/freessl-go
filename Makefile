windowsbuild:
	$env:GOOS="windows"
	go build main.go
linuxbuild:
	$env:GOOS="linux"
	go  build  main.go
build:
	go build main.go -o free_ssl.go