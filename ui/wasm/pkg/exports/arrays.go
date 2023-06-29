package exports

import "syscall/js"

type userCommand struct {
	Name             string
	Group            string
	Default          bool
	CommandLine      string
	HotReloadCapable bool
	WorkingDir       string
}

func getStringArray(value js.Value) []string {
	l := value.Length()
	result := make([]string, 0, l)
	for i := 0; i < l; i++ {
		s := value.Index(i).String()
		if len(s) > 0 {
			result = append(result, s)
		}
	}
	return result
}

func getUserCommandArray(value js.Value) []userCommand {
	l := value.Length()
	result := make([]userCommand, 0, l)
	for i := 0; i < l; i++ {
		v := value.Index(i)
		result = append(result, userCommand{
			Name:             v.Get("name").String(),
			Group:            v.Get("group").String(),
			Default:          v.Get("default").Bool(),
			CommandLine:      v.Get("commandLine").String(),
			HotReloadCapable: v.Get("hotReloadCapable").Bool(),
			WorkingDir:       v.Get("workingDir").String(),
		})
	}
	return result
}
