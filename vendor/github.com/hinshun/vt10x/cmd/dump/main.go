package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ptm, pts, err := pty.Open()
	if err != nil {
		return err
	}
	defer pts.Close()
	defer ptm.Close()

	c := exec.Command(os.Getenv("SHELL"))
	c.Stdout = pts
	c.Stdin = pts
	c.Stderr = pts

	var state vt10x.State
	term, err := vt10x.Create(&state, ptm)
	if err != nil {
		return err
	}
	defer term.Close()

	rows, cols := state.Size()
	vt10x.ResizePty(ptm, cols, rows)

	go func() {
		for {
			err := term.Parse()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}
		}
	}()

	err = c.Start()
	if err != nil {
		return err
	}

	time.Sleep(time.Second)
	fmt.Println(state.String())
	return nil
}
