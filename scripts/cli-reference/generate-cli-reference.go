package main

import (
	"fmt"

	"github.com/redhat-developer/odo/cmd"
)

func main() {
	fmt.Print(cmd.GenerateCLIReference())
}
