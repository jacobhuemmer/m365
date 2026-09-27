package main

import (
	"fmt"
	"os"
)

func main() {
	if err := generate("features", "acceptance/generated"); err != nil {
		fmt.Fprintln(os.Stderr, "acceptance-entrypoint-generator:", err)
		os.Exit(1)
	}
}
