module github.com/panyam/leetcoach

go 1.24

require (
	github.com/alexedwards/scs/v2 v2.8.0
	github.com/fatih/color v1.18.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/joho/godotenv v1.5.1
	github.com/panyam/oneauth v0.0.3
	github.com/panyam/s3gen v0.0.18
	github.com/panyam/templar v0.0.4
	golang.org/x/oauth2 v0.24.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/adrg/frontmatter v0.2.0 // indirect
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/morrisxyang/xreflect v0.0.0-20231001053442-6df0df9858ba // indirect
	github.com/panyam/goutils v0.1.2 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/yuin/goldmark v1.7.8 // indirect
	github.com/yuin/goldmark-highlighting v0.0.0-20220208100518-594be1970594 // indirect
	go.abhg.dev/goldmark/anchor v0.2.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/panyam/s3gen v0.0.18 => ./locallinks/s3gen

replace github.com/panyam/templar v0.0.4 => ./locallinks/templar/
