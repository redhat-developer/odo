package state

import (
	"fmt"
	"os"
)

func pidExists(pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	_, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}
	return true, nil
}
