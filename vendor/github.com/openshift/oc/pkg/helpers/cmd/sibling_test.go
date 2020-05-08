package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandInSiblingCommandPath(t *testing.T) {
	testCases := []struct {
		name     string
		command  *cobra.Command
		expected string
	}{
		{
			name:     "SiblingsGenerationOne",
			command:  newCommand("b", withParent("a", withCommand("test"))),
			expected: "a test",
		},
		{
			name: "SiblingsGenerationTwo",
			command: newCommand("c",
				withParent("b",
					withCommand("test"),
					withParent("a"),
				),
			),
			expected: "a b test",
		},
		{
			name: "Nibling",
			command: newCommand("c",
				withParent("b",
					withParent("a",
						withCommand("test"),
					),
				),
			),
			expected: "a test",
		},
		{
			name: "SiblingAndNibling",
			command: newCommand("c",
				withParent("b",
					withCommand("test"),
					withParent("a",
						withCommand("test"),
					),
				),
			),
			expected: "a b test",
		},
		{
			name: "GreatNibling",
			command: newCommand("d",
				withParent("c",
					withParent("b",
						withCommand("test"),
						withParent("a",
							withCommand("test"),
						),
					),
				),
			),
			expected: "a b test",
		},
		{
			name: "Cousin",
			command: newCommand("c",
				withParent("b",
					withParent("a",
						withCommand("d",
							withCommand("test")),
					),
				),
			),
			expected: "test",
		},
		{
			name:     "Root",
			command:  newCommand("a"),
			expected: "test",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args0Save := os.Args[0]
			defer func() { os.Args[0] = args0Save }()
			os.Args[0] = tc.command.Root().Name()
			result := strings.Join(SiblingOrNiblingCommand(tc.command, "test"), " ")
			if tc.expected != result {
				t.Fatalf("Expected: '%v', Actual: '%v", tc.expected, result)
			}
		})
	}
}

func newCommand(name string, options ...func(*cobra.Command)) *cobra.Command {
	cmd := &cobra.Command{Use: name}
	for _, f := range options {
		f(cmd)
	}
	return cmd
}

func withCommand(name string, options ...func(*cobra.Command)) func(*cobra.Command) {
	return func(parent *cobra.Command) {
		child := &cobra.Command{Use: name}
		parent.AddCommand(child)
		for _, f := range options {
			f(child)
		}
	}
}

func withParent(name string, options ...func(*cobra.Command)) func(*cobra.Command) {
	return func(child *cobra.Command) {
		parent := &cobra.Command{Use: name}
		parent.AddCommand(child)
		for _, f := range options {
			f(parent)
		}
	}
}
