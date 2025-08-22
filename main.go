package main

import (
	"github.com/bqdanh/stock-trading-be/cmd"
	"os"
)

func main() {
	appCli := cmd.AppCommandLineInterface()
	if err := appCli.Run(os.Args); err != nil {
		panic(err)
	}

}
