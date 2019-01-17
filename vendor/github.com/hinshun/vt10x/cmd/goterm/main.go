package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gdamore/tcell"
	"github.com/hinshun/vt10x"
	"github.com/kr/pty"
)

func main() {
	err := goterm()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func goterm() error {
	cmd := exec.Command(os.Getenv("SHELL"), "-i")
	ptm, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	// f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	// if err != nil {
	// 	return err
	// }
	// state := vt10x.State{
	// 	DebugLogger: log.New(f, "", log.LstdFlags),
	// }
	var state vt10x.State
	term, err := vt10x.Create(&state, ptm)
	if err != nil {
		return err
	}
	defer term.Close()

	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	defer s.Fini()

	err = s.Init()
	if err != nil {
		return err
	}

	width, height := s.Size()
	vt10x.ResizePty(ptm, width, height)
	term.Resize(width, height)

	endc := make(chan bool)
	updatec := make(chan struct{}, 1)
	go func() {
		defer close(endc)
		for {
			err := term.Parse()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}
			select {
			case updatec <- struct{}{}:
			default:
			}
		}
	}()

	go func() {
		io.Copy(ptm, os.Stdin)
	}()

	eventc := make(chan tcell.Event, 4)
	go func() {
		for {
			eventc <- s.PollEvent()
		}
	}()

	for {
		select {
		case event := <-eventc:
			switch ev := event.(type) {
			case *tcell.EventResize:
				width, height = ev.Size()
				vt10x.ResizePty(ptm, width, height)
				term.Resize(width, height)
				s.Sync()
			}
		case <-endc:
			return nil
		case <-updatec:
			update(s, &state, width, height)
		}
	}
}

func update(s tcell.Screen, state *vt10x.State, w, h int) {
	state.Lock()
	defer state.Unlock()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c, fg, bg := state.Cell(x, y)

			style := tcell.StyleDefault
			if fg != vt10x.DefaultFG {
				style = style.Foreground(tcell.Color(fg))
			}
			if bg != vt10x.DefaultBG {
				style = style.Background(tcell.Color(bg))
			}

			s.SetContent(x, y, c, nil, style)
		}
	}
	if state.CursorVisible() {
		curx, cury := state.Cursor()
		s.ShowCursor(curx, cury)
	} else {
		s.HideCursor()
	}
	s.Show()
}
