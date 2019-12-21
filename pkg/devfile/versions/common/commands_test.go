package common

import "testing"

func TestValidateCommands(t *testing.T) {

	t.Run("valid commands", func(t *testing.T) {

		commandTable := []struct {
			commands []DevfileCommand
		}{
			{
				commands: []DevfileCommand{
					{Name: "build app"},
					{Name: "run app"},
				},
			},
			{
				commands: []DevfileCommand{
					{Name: "buildApp"},
					{Name: "runApp"},
				},
			},
			{
				commands: []DevfileCommand{
					{Name: "AppBuildSomeServer"},
					{Name: "AppRunSomeServer"},
				},
			},
		}

		for _, cmd := range commandTable {
			err := ValidateCommands(cmd.commands)
			if err != nil {
				t.Errorf("Unexpected error: '%v'", err)
			}
		}
	})

	t.Run("build command not present", func(t *testing.T) {

		commands := []DevfileCommand{
			{Name: "run app"},
		}

		err := ValidateCommands(commands)
		if err == nil {
			t.Errorf("Expected error, didn't get one")
		}
	})

	t.Run("run command not present", func(t *testing.T) {

		commands := []DevfileCommand{
			{Name: "run app"},
		}

		err := ValidateCommands(commands)
		if err == nil {
			t.Errorf("Expected error, didn't get one")
		}
	})
}
