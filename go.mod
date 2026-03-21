module github.com/panyam/sdl

go 1.25.0

require (
	connectrpc.com/connect v1.19.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
	github.com/panyam/goapplib v0.0.34
	github.com/panyam/goutils v0.1.13
	github.com/panyam/protoc-gen-go-wasmjs v0.0.33
	github.com/panyam/servicekit v0.0.6
	github.com/panyam/templar v0.0.31
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	google.golang.org/genproto/googleapis/api v0.0.0-20260319201613-d00831a3d3e7
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/panyam/gocurrent v0.0.11 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260319201613-d00831a3d3e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/panyam/goapplib v0.0.34 => ./locallinks/newstack/goapplib/main

replace github.com/panyam/templar v0.0.31 => ./locallinks/newstack/templar/main
