package main

import (
	"fmt"

	"github.com/redhat-developer/ocdev/cmd"
)

func main() {
	fmt.Print(cmd.GenerateCLIDocs())
}
