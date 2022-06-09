package remotecmd

import (
	"io"
	"reflect"
	"testing"
)

func Test_createConsoleOutputWriterAndChannel(t *testing.T) {

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Close channel with no text sent (nil input)",
			input:    nil,
			expected: nil,
		},
		{
			name:     "Close channel with no text sent (zero-length input slice)",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "Close channel with a single line of text sent",
			input:    []string{"one"},
			expected: []string{"one"},
		},
		{
			name:     "Close channel with multiple lines of text sent",
			input:    []string{"one", "two", "three", "four", "five"},
			expected: []string{"one", "two", "three", "four", "five"},
		},
	}
	writeInputData := func(in []string, w *io.PipeWriter) {
		defer w.Close()
		for _, s := range in {
			_, err := w.Write([]byte(s + "\n"))
			if err != nil {
				return
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			inputWriter, outputChan := createConsoleOutputWriterAndChannel()

			// Write input text, in a separate goroutine to be able to iterate over outputChan as long as data is being written to inputWriter
			go writeInputData(tt.input, inputWriter)

			var out []string
			for s := range outputChan {
				out = append(out, s)
			}

			if len(out) != len(tt.expected) {
				t.Fatalf("length of output response %v did not match length of expected output %v", out, tt.expected)
			}
			if !reflect.DeepEqual(tt.expected, out) {
				t.Fatalf("output response %v did not match expected output %v", out, tt.expected)
			}

		})
	}

}
