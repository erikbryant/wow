package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/application"
	"github.com/erikbryant/wow/internal/shopping"
)

func main() {
	app, err := application.New("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = shopping.Shop(app)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
