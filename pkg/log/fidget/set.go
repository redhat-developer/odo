package fidget

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cdrage/uilive"
	"github.com/spf13/pflag"
)

// Set of default constants with regards to formatting and ticker length
const tickerLength = 300
const defaultSuccessIcon = "✓"
const defaultFailIcon = "✗"
const defaultTitlePrefix = " ┌"
const defaultPrefix = " ├─ "
const defaultEndPrefix = " └─ "

// This will make sure that when using go routines, we're never going to encounter
// a race condition of reading / writing at the same time (across all spinners)
var lock sync.Mutex

// SpinnerSet is a set of spinners
type SpinnerSet struct {
	spinners []*Spinny

	// Determine if we're in debug mode or not (no spinners..)
	debug bool

	// "Protected" by Mutex, so we're not writing all over the place
	mu *sync.Mutex

	// A WaitGroup so we can wait get that extra "tick" output for the "final output"
	wg sync.WaitGroup

	// Stop will **ALL** the spinners.. Use with caution
	stop chan struct{}

	// The actual ticker time
	ticker *time.Ticker

	// Writer (we're using uilive for multiple output)
	writer      *uilive.Writer
	debugWriter io.Writer

	// Symbols, formatting and the actual "ticker"
	Frames      []string
	TitlePrefix string
	Prefix      string
	EndPrefix   string // Useful for prefixes that don't match as the "last" spinner.. ex: ├──  vs └──
	Fail        string
	Success     string
	Title       string
}

// Spinny is a tiny little spinny spinner!
type Spinny struct {
	message      string
	status       string
	start        time.Time
	timeFinished string
	debug        bool
}

// GetMessage Retrieve the message.
// We do this because we do NOT want a race condition (odo being ran multiple times at the same time..)
// This was imeplemented because there are over 40 (yes 40!) go routines being ran at the same time when using our testing suite +
// testing getting all the spinners.
func (s *Spinny) GetMessage() string {
	lock.Lock()
	message := s.message
	lock.Unlock()
	return message
}

// GetStatus retrieves the status
func (s *Spinny) GetStatus() string {
	lock.Lock()
	status := s.status
	lock.Unlock()
	return status
}

// Status updates the message output
func (s *Spinny) Status(message string) {
	lock.Lock()
	s.message = message

	// If we're in debug mode, just print it out once to log (since we're not "spinning")
	if s.debug {
		outputDebug(s.message)
	}
	lock.Unlock()
}

// Fail sets the status to fail
func (s *Spinny) Fail() {
	lock.Lock()
	s.timeFinished = s.TimeSpent()
	s.status = "fail"
	lock.Unlock()
}

// Success sets the status to success
func (s *Spinny) Success() {
	lock.Lock()
	s.timeFinished = s.TimeSpent()
	s.status = "success"
	lock.Unlock()
}

// NewSpinny creates a new spinner
func NewSpinny(message string) *Spinny {
	return &Spinny{
		message: message,
		status:  "active",
		start:   time.Now(),
		debug:   IsDebug(),
	}
}

// NewSpinnerSet initializes and returns a new spinner set with default values.
func NewSpinnerSet(spinners []*Spinny) *SpinnerSet {

	// If we're running this on Windows, don't use unicode, use the asciiSpinnerFrames
	// same goes with the fail / success icons
	frames := unicodeSpinnerFrames
	success := defaultSuccessIcon
	fail := defaultFailIcon
	if runtime.GOOS == "windows" {
		frames = asciiSpinnerFrames
		success = "V"
		fail = "X"
	}

	return &SpinnerSet{
		spinners:    spinners,
		stop:        make(chan struct{}, 1),
		ticker:      time.NewTicker(time.Millisecond * tickerLength),
		mu:          &sync.Mutex{},
		Frames:      frames,
		Success:     success,
		Fail:        fail,
		TitlePrefix: defaultTitlePrefix,
		Prefix:      defaultPrefix,
		EndPrefix:   defaultEndPrefix,
		writer:      uilive.New(),
		debugWriter: os.Stdout,
		debug:       IsDebug(),
	}
}

// Stop stops the spinner (internal function)
func (s *SpinnerSet) Stop() {
	// Wait for the goroutine to finish and for it do to its last tick output
	// if we're in debug, who cares
	if !s.debug {
		s.stop <- struct{}{}
		s.wg.Wait()
	}
}

// End stops the WHOLE spinner set, if set to "False" it will assume everything that's
// still active has failed. If it's set to True, it will assume everything has succeeded
func (s *SpinnerSet) End(end bool) {

	for _, spinny := range s.spinners {
		if spinny.GetStatus() == "active" && end {
			spinny.Success()
		} else if spinny.GetStatus() == "active" && !end {
			spinny.Fail()
		}
	}

	s.Stop()

}

// Start starts the entire spinner (in a goroutine as well!)
func (s *SpinnerSet) Start() {

	// We are deploying only *one* go routine, so let's add it to the wait group.
	// the goroutine will forever loop until it has received a stop on the channel.
	s.wg.Add(1)

	// Start the writer

	if s.debug {
		// Tick once (and only once), there will be no spinners here.
		s.Tick(" •  ")
	} else {
		s.writer.Start()

		// Create the goroutine
		go func() {

			// Endless loop
			for {

				// Use the same "tick" for all of the spinners..
				for _, frame := range s.Frames {
					select {

					// Stop if s.Stop() has been passed in. This stops ALL of the spinners, regardless of their status.. to
					// make sure that it actually outputs correctly!
					case <-s.stop:
						// Do *one* more "tick"* then stop in order to output the final status updates
						s.Tick(frame)
						s.writer.Stop()
						s.wg.Done()
						return

					case <-s.ticker.C:
						// Print one "tick" / output.
						s.Tick(frame)
					}
				}

			}

		}()

	}

}

// Tick is the "output" functionality.
func (s *SpinnerSet) Tick(frame string) {

	// Lock when printing (just in case) so we aren't overriding anything
	s.mu.Lock()
	defer s.mu.Unlock()

	// Show a title if passed in
	if s.Title != "" {
		_, _ = fmt.Fprintf(s.writer.Newline(), "%s %s\n", s.TitlePrefix, s.Title)
	}

	for k, v := range s.spinners {

		// Use the "end" prefix if it's the last spinner..
		prefix := s.Prefix
		if k == len(s.spinners)-1 {
			prefix = s.EndPrefix
		}

		// Output the current status
		switch v.GetStatus() {
		case "fail":
			_, _ = fmt.Fprintf(s.writer.Newline(), "%s%s %s [%s]\n", prefix, s.Fail, v.GetMessage(), v.timeFinished)

		case "success":
			_, _ = fmt.Fprintf(s.writer.Newline(), "%s%s %s [%s]\n", prefix, s.Success, v.GetMessage(), v.timeFinished)

		default:
			_, _ = fmt.Fprintf(s.writer.Newline(), "%s%s %s [%s]\n", prefix, frame, v.GetMessage(), v.TimeSpent())
		}

	}

}

// TimeSpent returns the seconds spent since the spinner first started
func (s *Spinny) TimeSpent() string {
	currentTime := time.Now()
	timeElapsed := currentTime.Sub(s.start)

	// Print ms if less than a second
	// else print out minutes if more than 1 minute
	// else print the default (seconds)
	if timeElapsed > time.Minute {
		return fmt.Sprintf("%.0fm", timeElapsed.Minutes())
	} else if timeElapsed < time.Minute && timeElapsed > time.Second {
		return fmt.Sprintf("%.0fs", timeElapsed.Seconds())
	} else if timeElapsed < time.Second && timeElapsed > time.Millisecond {
		return fmt.Sprintf("%dms", timeElapsed.Nanoseconds()/int64(time.Millisecond))
	}

	return fmt.Sprintf("%dns", timeElapsed.Nanoseconds())
}

func outputDebug(out string) {
	fmt.Println(" •  " + out)
}

// IsDebug has been copied over from log/status.go, this checks to see if we're on debug or not.
func IsDebug() bool {

	flag := pflag.Lookup("v")

	if flag != nil {
		return !strings.Contains(pflag.Lookup("v").Value.String(), "0")
	}

	return false
}
