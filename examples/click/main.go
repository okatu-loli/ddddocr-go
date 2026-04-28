package main

import (
	"encoding/json"
	"fmt"
	"os"

	ddddocr "github.com/okatu-loli/ddddocr-go"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/click <image>")
		os.Exit(2)
	}

	client := ddddocr.NewClient(ddddocr.ClientConfig{})
	result, err := client.ClickFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
