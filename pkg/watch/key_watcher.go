package watch

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// getKeyWatcher returns a channel which will emit
// characters when keys are pressed on the keyboard
func getKeyWatcher(ctx context.Context, out io.Writer) <-chan byte {

	keyInput := make(chan byte)

	go func() {
		stdinfd := int(os.Stdin.Fd())
		if !term.IsTerminal(stdinfd) {
			return
		}

		// First set the terminal in character mode
		// to be able to read characters as soon as
		// they are emitted, instead of waiting
		// for newline characters
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

		// Wait for the context to be cancelled
		// or a character to be emitted
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

// getKey returns a channel which will emit a character
// when a key is pressed on the keyboard
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
