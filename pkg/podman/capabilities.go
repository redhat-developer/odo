package podman

type Capabilities struct {
	Cgroupv2 bool
}

func (o *PodmanCli) GetCapabilities() (Capabilities, error) {
	var result Capabilities
	info, err := o.getInfo()
	if err != nil {
		return Capabilities{}, err
	}
	if info.Host.CgroupsVersion == "v2" {
		result.Cgroupv2 = true
	}
	return result, nil
}
