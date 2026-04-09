.PHONY: clean test

mackerel-plugin-prometheus-query: *.go lib/*.go
	go build -o mackerel-plugin-prometheus-query .

clean:
	rm -f mackerel-plugin-prometheus-query dist/*

test:
	go test -race ./...
