package v1alpha2

func (container DevWorkspaceTemplateSpecContent) GetToplevelLists() TopLevelLists {
	return TopLevelLists{
		"Components":      extractKeys(container.Components),
		"Projects":        extractKeys(container.Projects),
		"StarterProjects": extractKeys(container.StarterProjects),
		"Commands":        extractKeys(container.Commands),
	}
}

func (container ParentOverrides) GetToplevelLists() TopLevelLists {
	return TopLevelLists{
		"Components":      extractKeys(container.Components),
		"Projects":        extractKeys(container.Projects),
		"StarterProjects": extractKeys(container.StarterProjects),
		"Commands":        extractKeys(container.Commands),
	}
}

func (container PluginOverridesParentOverride) GetToplevelLists() TopLevelLists {
	return TopLevelLists{
		"Components": extractKeys(container.Components),
		"Commands":   extractKeys(container.Commands),
	}
}

func (container PluginOverrides) GetToplevelLists() TopLevelLists {
	return TopLevelLists{
		"Components": extractKeys(container.Components),
		"Commands":   extractKeys(container.Commands),
	}
}
