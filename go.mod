module github.com/indaco/tempo

go 1.23.5

toolchain go1.23.9

require (
	github.com/evanw/esbuild v0.25.5
	github.com/fatih/color v1.18.0
	github.com/indaco/tempo-api v0.0.0-20250217085709-fd62d35b4d54
	github.com/urfave/cli/v3 v3.3.8
	golang.org/x/mod v0.25.0
	golang.org/x/sync v0.15.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

// replace (
// 	github.com/indaco/tempo-api => ../tempo-api
// )
