module github.com/SergeyDavidenko/logger/benchmarks

go 1.25

require (
	github.com/SergeyDavidenko/logger v0.0.0
	github.com/apex/log v1.9.0
	github.com/go-kit/log v0.2.1
	github.com/inconshreveable/log15 v2.16.0+incompatible
	github.com/rs/zerolog v1.34.0
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.0
)

require (
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
)

replace github.com/SergeyDavidenko/logger => ../
