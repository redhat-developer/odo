package portForward

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/fatih/color"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"

	"k8s.io/klog"
)

type PortWriter struct {
	buffer io.Writer
	end    chan bool
	len    int
	// mapping indicates the list of endpoints open by containers
	mapping map[string][]v1alpha2.Endpoint
	fwPorts []api.ForwardedPort
}

// NewPortWriter creates a writer that will write the content in buffer,
// and Wait will return after strings "Forwarding from 127.0.0.1:" has been written "len" times
func NewPortWriter(buffer io.Writer, len int, mapping map[string][]v1alpha2.Endpoint) *PortWriter {
	return &PortWriter{
		buffer:  buffer,
		len:     len,
		end:     make(chan bool),
		mapping: mapping,
	}
}

func (o *PortWriter) Write(buf []byte) (n int, err error) {

	s := string(buf)
	if strings.HasPrefix(s, "Forwarding from 127.0.0.1") {

		fwPort, err := getForwardedPort(o.mapping, s)
		if err == nil {
			o.fwPorts = append(o.fwPorts, fwPort)
		} else {
			klog.V(4).Infof("unable to get forwarded port: %v", err)
		}

		// Also set the colour to bolded green for easier readability
		fmt.Fprintf(o.buffer, " -  %s", log.SboldColor(color.FgGreen, s))
		o.len--
		if o.len == 0 {
			o.end <- true
		}
	}
	return len(buf), nil
}

func (o *PortWriter) Wait() {
	<-o.end
}

func (o *PortWriter) GetForwardedPorts() []api.ForwardedPort {
	return o.fwPorts
}

func getForwardedPort(mapping map[string][]v1alpha2.Endpoint, s string) (api.ForwardedPort, error) {
	regex := regexp.MustCompile(`Forwarding from 127.0.0.1:([0-9]+) -> ([0-9]+)`)
	matches := regex.FindStringSubmatch(s)
	if len(matches) < 3 {
		return api.ForwardedPort{}, errors.New("unable to analyze port forwarding string")
	}
	localPort, err := strconv.Atoi(matches[1])
	if err != nil {
		return api.ForwardedPort{}, err
	}
	remotePort, err := strconv.Atoi(matches[2])
	if err != nil {
		return api.ForwardedPort{}, err
	}
	fp := api.ForwardedPort{
		LocalAddress:  "127.0.0.1",
		LocalPort:     localPort,
		ContainerPort: remotePort,
	}
containerLoop:
	for container, endpoints := range mapping {
		for _, ep := range endpoints {
			if ep.TargetPort == remotePort {
				fp.ContainerName = container
				fp.PortName = ep.Name
				fp.Exposure = string(ep.Exposure)
				break containerLoop
			}
		}
	}
	return fp, nil
}
