package main

import (
	"fmt"
	"os"

	promq "github.com/fujiwara/mackerel-plugin-prometheus-query/lib"
)

func main() {
	if err := promq.Do(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
