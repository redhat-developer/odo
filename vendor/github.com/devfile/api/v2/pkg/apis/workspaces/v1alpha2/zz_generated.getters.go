package v1alpha2

// GetIsDefault returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *CommandGroup) GetIsDefault() bool {
	return getBoolOrDefault(in.IsDefault, false)
}

// GetHotReloadCapable returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *ExecCommand) GetHotReloadCapable() bool {
	return getBoolOrDefault(in.HotReloadCapable, false)
}

// GetParallel returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *CompositeCommand) GetParallel() bool {
	return getBoolOrDefault(in.Parallel, false)
}

// GetDedicatedPod returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *Container) GetDedicatedPod() bool {
	return getBoolOrDefault(in.DedicatedPod, false)
}

// GetAutoBuild returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *ImageUnion) GetAutoBuild() bool {
	return getBoolOrDefault(in.AutoBuild, false)
}

// GetRootRequired returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *Dockerfile) GetRootRequired() bool {
	return getBoolOrDefault(in.RootRequired, false)
}

// GetDeployByDefault returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *K8sLikeComponent) GetDeployByDefault() bool {
	return getBoolOrDefault(in.DeployByDefault, false)
}

// GetEphemeral returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *Volume) GetEphemeral() bool {
	return getBoolOrDefault(in.Ephemeral, false)
}

// GetSecure returns the value of the boolean property.  If unset, it's the default value specified in the devfile:default:value marker
func (in *Endpoint) GetSecure() bool {
	return getBoolOrDefault(in.Secure, false)
}

func getBoolOrDefault(input *bool, defaultVal bool) bool {
	if input != nil {
		return *input
	}
	return defaultVal
}
