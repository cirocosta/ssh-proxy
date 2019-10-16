package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

var cli struct {
	Server serverCommand `command:"server"`
	Client clientCommand `command:"client"`
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
