package watch

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/term"
)

func getKeyWatcher(ctx context.Context) chan byte {

	keyInput := make(chan byte)

	go func() {
		stdinfd := int(os.Stdin.Fd())
		oldState, err := term.GetState(stdinfd)
		if err != nil {
			fmt.Println(fmt.Errorf("getstate: %w", err))
			return
		}
		err = enableCharInput(stdinfd)
		if err != nil {
			fmt.Println(fmt.Errorf("enableCharInput: %w", err))
			return
		}
		defer func() {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case b := <-getKey():
				keyInput <- b
			}
		}

	}()

	return keyInput
}

func getKey() chan byte {

	ch := make(chan byte)

	go func() {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(fmt.Errorf("read: %w", err))
			return
		}
		ch <- b[0]
	}()

	return ch
}
