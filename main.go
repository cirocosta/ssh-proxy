package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var cli struct {
	Serve   serveCommand   `command:"serve"`
	Connect connectCommand `command:"connect"`
}

func main() {
	parser := flags.NewParser(&cli, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
