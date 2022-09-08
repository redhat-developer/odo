package watch

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func getKeyWatcher(ctx context.Context, out io.Writer) <-chan byte {

	keyInput := make(chan byte)

	go func() {
		stdinfd := int(os.Stdin.Fd())
		if !term.IsTerminal(stdinfd) {
			return
		}

		oldState, err := term.GetState(stdinfd)
		if err != nil {
			fmt.Fprintln(out, fmt.Errorf("getstate: %w", err))
			return
		}
		err = enableCharInput(stdinfd)
		if err != nil {
			fmt.Fprintln(out, fmt.Errorf("enableCharInput: %w", err))
			return
		}
		defer func() {
			_ = term.Restore(stdinfd, oldState)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case b := <-getKey(out):
				keyInput <- b
			}
		}
	}()

	return keyInput
}

func getKey(out io.Writer) <-chan byte {

	ch := make(chan byte)

	go func() {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Fprintln(out, fmt.Errorf("read: %w", err))
			return
		}
		ch <- b[0]
	}()

	return ch
}
