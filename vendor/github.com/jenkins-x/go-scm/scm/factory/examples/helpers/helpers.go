package helpers

import (
	"fmt"
	"os"
)

// Fail fails the program due to an error
func Fail(err error) {
	fmt.Printf("ERROR: %s\n", err.Error())
	os.Exit(1)
}
