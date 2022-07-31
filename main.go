package main

import (
	"os"

	"github.com/whalesburg/dero-stratum-miner/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
