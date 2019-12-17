GIT_VER := $(shell git describe --tags)
export GO111MODULE := on

.PHONY: clean test dist release

mackerel-plugin-prometheus-query: *.go lib/*.go
	go build -o mackerel-plugin-prometheus-query .

clean:
	rm -f mackerel-plugin-prometheus-query dist/*

test:
	go test -race ./...

dist:
	CGO_ENABLED=0 \
		goxz \
		-build-ldflags="-s -w -X main.Version=${GIT_VER}" \
		-os=darwin,linux -arch=amd64 -d=dist .

release:
	ghr -u fujiwara -r mackerel-plugin-prometheus-query -n "$(GIT_VER)" $(GIT_VER) dist/
