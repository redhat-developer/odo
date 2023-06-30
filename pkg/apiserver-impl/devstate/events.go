package devstate

func (o *DevfileState) UpdateEvents(event string, commands []string) (DevfileContent, error) {
	switch event {
	case "postStart":
		o.Devfile.Data.UpdateEvents(commands, nil, nil, nil)
	case "postStop":
		o.Devfile.Data.UpdateEvents(nil, commands, nil, nil)
	case "preStart":
		o.Devfile.Data.UpdateEvents(nil, nil, commands, nil)
	case "preStop":
		o.Devfile.Data.UpdateEvents(nil, nil, nil, commands)
	}
	return o.GetContent()
}
