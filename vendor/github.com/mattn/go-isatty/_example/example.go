package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

func main() {
	if isatty.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println("Is Terminal")
	} else {
		fmt.Println("Is Not Terminal")
	}
}