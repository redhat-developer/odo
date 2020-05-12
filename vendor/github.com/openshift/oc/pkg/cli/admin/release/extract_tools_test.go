package release

import (
	"bytes"
	"testing"
)

func Test_copyAndReplace(t *testing.T) {
	buffer := 4
	tests := []struct {
		name         string
		input        string
		replacements []replacement
		expected     string
		error        string
	}{
		{
			name:  "buffer too small",
			input: "1234",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aaaaa"),
					value:  "A",
				},
			},
			error: "the buffer size must be greater than 5 bytes to find rep-A",
		},
		{
			name:     "buffer too large",
			input:    "123",
			expected: "123",
		},
		{
			name:  "value too large",
			input: "1234",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "AA",
				},
			},
			error: "the rep-A value has 2 bytes, but the maximum replacement length is 1",
		},
		{
			name:     "A beginning of file",
			input:    "aa345678",
			expected: "A\x00345678",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "A end of buffer",
			input:    "12aa5678",
			expected: "12A\x005678",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "A cross buffer",
			input:    "123aa678",
			expected: "123A\x00678",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "A beginning of buffer",
			input:    "1234aa78",
			expected: "1234A\x0078",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "A end of file",
			input:    "123456aa",
			expected: "123456A\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "A buffer too large",
			input:    "12345aa",
			expected: "12345A\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
			},
		},
		{
			name:     "AB beginning of file",
			input:    "aabb5678",
			expected: "A\x00B\x005678",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
		{
			name:     "BA beginning of file",
			input:    "bbaa5678",
			expected: "B\x00A\x005678",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
		{
			name:     "AB end of buffer",
			input:    "1234aabb",
			expected: "1234A\x00B\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
		{
			name:     "AB cross buffer",
			input:    "123aa6bb",
			expected: "123A\x006B\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
		{
			name:     "AB end of file",
			input:    "1234aabb",
			expected: "1234A\x00B\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
		{
			name:     "BA end of file",
			input:    "1234bbaa",
			expected: "1234B\x00A\x00",
			replacements: []replacement{
				{
					name:   "rep-A",
					marker: []byte("aa"),
					value:  "A",
				},
				{
					name:   "rep-B",
					marker: []byte("bb"),
					value:  "B",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tt.input))
			w := &bytes.Buffer{}
			err := copyAndReplace(nil, w, r, buffer, tt.replacements, "test")
			if (err == nil && tt.error != "") || (err != nil && err.Error() != tt.error) {
				t.Fatalf("unexpected error: %v != %v", err, tt.error)
			}
			actual := w.String()
			if actual != tt.expected {
				t.Fatalf("unexpected response body: %q != %q", actual, tt.expected)
			}
		})
	}
}
