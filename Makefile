build:
	CGO_ENABLED=1 go build --ldflags '-extldflags "-static"' -o iftree cmd/iftree/main.go 

