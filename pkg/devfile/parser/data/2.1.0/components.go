package version210

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// GetComponents returns the slice of DevfileComponent objects parsed from the Devfile
func (d *Devfile210) GetComponents() []common.DevfileComponent {
	return d.Components
}

// AddComponents adds the slice of DevfileComponent objects to the devfile's components
// if a component is already defined, error out
func (d *Devfile210) AddComponents(components []common.DevfileComponent) error {
	componentMap := make(map[string]bool)
	for _, component := range d.Components {
		componentMap[component.Container.Name] = true
	}

	for _, component := range components {
		if _, ok := componentMap[component.Container.Name]; !ok {
			d.Components = append(d.Components, component)
		} else {
			return fmt.Errorf("component %v is already present in the devfile", component.Container.Name)
		}
	}
	return nil
}

// UpdateComponent updates the component with the given name
func (d *Devfile210) UpdateComponent(component common.DevfileComponent) {
	for i := range d.Components {
		if d.Components[i].Container.Name == strings.ToLower(component.Container.Name) {
			d.Components[i] = component
		}
	}
}

// GetCommands returns the slice of DevfileCommand objects parsed from the Devfile
func (d *Devfile210) GetCommands() []common.DevfileCommand {
	var commands []common.DevfileCommand

	for _, command := range d.Commands {
		// we convert devfile command id to lowercase so that we can handle
		// cases efficiently without being error prone
		// we also convert the odo push commands from build-command and run-command flags
		if command.Exec != nil {
			command.Exec.Id = strings.ToLower(command.Exec.Id)
		} else if command.Composite != nil {
			command.Composite.Id = strings.ToLower(command.Composite.Id)
		}

		commands = append(commands, command)
	}

	return commands
}

// AddCommands adds the slice of DevfileCommand objects to the Devfile's commands
// if a command is already defined, error out
func (d *Devfile210) AddCommands(commands []common.DevfileCommand) error {
	commandsMap := make(map[string]bool)
	for _, command := range d.Commands {
		commandsMap[command.Exec.Id] = true
	}

	for _, command := range commands {
		if _, ok := commandsMap[command.Exec.Id]; !ok {
			d.Commands = append(d.Commands, command)
		} else {
			return fmt.Errorf("command %v is already present in the devfile", command.Exec.Id)
		}
	}
	return nil
}

// UpdateCommand updates the command with the given id
func (d *Devfile210) UpdateCommand(command common.DevfileCommand) {
	for i := range d.Commands {
		if d.Commands[i].Exec.Id == strings.ToLower(command.Exec.Id) {
			d.Commands[i] = command
		}
	}
}

// GetParent returns the DevfileParent object parsed from devfile
func (d *Devfile210) GetParent() common.DevfileParent {
	return d.Parent
}

// SetParent sets the parent for the devfile
func (d *Devfile210) SetParent(parent common.DevfileParent) {
	d.Parent = parent
}

// GetProjects returns the DevfileProject Object parsed from devfile
func (d *Devfile210) GetProjects() []common.DevfileProject {
	return d.Projects
}

// AddProjects adss the slice of Devfile projects to the Devfile's project list
// if a project is already defined, error out
func (d *Devfile210) AddProjects(projects []common.DevfileProject) error {
	projectsMap := make(map[string]bool)
	for _, project := range d.Projects {
		projectsMap[project.Name] = true
	}

	for _, project := range projects {
		if _, ok := projectsMap[project.Name]; !ok {
			d.Projects = append(d.Projects, project)
		} else {
			return fmt.Errorf("project %v is already present in the devfile", project.Name)
		}
	}
	return nil
}

// UpdateProject updates the slice of DevfileCommand projects parsed from the Devfile
func (d *Devfile210) UpdateProject(project common.DevfileProject) {
	for i := range d.Projects {
		if d.Projects[i].Name == strings.ToLower(project.Name) {
			d.Projects[i] = project
		}
	}
}

//SetSchemaVersion sets devfile schema version
func (d *Devfile210) SetSchemaVersion(version string) {
	d.SchemaVersion = version
}

// GetMetadata returns the DevfileMetadata Object parsed from devfile
func (d *Devfile210) GetMetadata() common.DevfileMetadata {
	return d.Metadata
}

// SetMetadata sets the metadata for devfile
func (d *Devfile210) SetMetadata(name, version string) {
	d.Metadata = common.DevfileMetadata{
		Name:    name,
		Version: version,
	}
}

// GetEvents returns the Events Object parsed from devfile
func (d *Devfile210) GetEvents() common.DevfileEvents {
	return d.Events
}

// AddEvents adds the Events Object to the devfile's events
// if the event is already defined in the devfile, error out
func (d *Devfile210) AddEvents(events common.DevfileEvents) error {
	if len(events.PreStop) > 0 {
		if len(d.Events.PreStop) > 0 {
			return fmt.Errorf("pre stop event is already defined in the devfile")
		} else {
			d.Events.PreStop = events.PreStop
		}
	}

	if len(events.PreStart) > 0 {
		if len(d.Events.PreStart) > 0 {
			return fmt.Errorf("pre start event is already defined in the devfile")
		} else {
			d.Events.PreStart = events.PreStart
		}
	}

	if len(events.PostStop) > 0 {
		if len(d.Events.PostStop) > 0 {
			return fmt.Errorf("post stop event is already defined in the devfile")
		} else {
			d.Events.PostStop = events.PostStop
		}
	}

	if len(events.PostStart) > 0 {
		if len(d.Events.PostStart) > 0 {
			return fmt.Errorf("post start event is already defined in the devfile")
		} else {
			d.Events.PostStart = events.PostStart
		}
	}
	return nil
}

// UpdateEvents updates the devfile's events
// it only updates the events passed to it
func (d *Devfile210) UpdateEvents(postStart, postStop, preStart, preStop []string) {
	if len(postStart) != 0 {
		d.Events.PostStart = postStart
	}
	if len(postStop) != 0 {
		d.Events.PostStop = postStop
	}
	if len(preStart) != 0 {
		d.Events.PreStart = preStart
	}
	if len(preStop) != 0 {
		d.Events.PreStop = preStop
	}
}

// GetAliasedComponents returns the slice of DevfileComponent objects that each have an alias
func (d *Devfile210) GetAliasedComponents() []common.DevfileComponent {
	// V2 has name required in jsonSchema
	return d.Components
}
