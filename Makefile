all: plotly-latest.min.js go/bin/retirement

plotly-latest.min.js:
	curl https://cdn.plot.ly/plotly-latest.min.js > $@

go/bin/retirement: go/src/github.com/satori/go.uuid
	env GOPATH=`pwd`/go go install retirement

go/src/github.com/satori/go.uuid:
	env GOPATH=`pwd`/go go get github.com/satori/go.uuid

