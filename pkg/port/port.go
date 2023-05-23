package port

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/backo-go"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/remotecmd"
)

// Order of values in the TCP States enum.
// See https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/net/tcp_states.h#n12
var connectionStates = []string{
	"ESTABLISHED",
	"SYN_SENT",
	"SYN_RECV",
	"FIN_WAIT1",
	"FIN_WAIT2",
	"TIME_WAIT",
	"CLOSE",
	"CLOSE_WAIT",
	"LAST_ACK",
	"LISTEN",
	"CLOSING",
	"NEW_SYN_RECV",

	"MAX_STATES", // Leave at the end!
}

// every 2 other characters
var ipv4HexRegExp = regexp.MustCompile(".{2}")

type Connection struct {
	LocalAddress  string
	LocalPort     int
	RemoteAddress string
	RemotePort    int
	State         string
}

func (c Connection) String() string {
	return fmt.Sprintf("[%s] %s:%d -> %s:%d", c.State, c.LocalAddress, c.LocalPort, c.RemoteAddress, c.RemotePort)
}

// DetectRemotePortsBoundOnLoopback filters the given ports by returning only those that are actually bound to the loopback interface in the specified container.
func DetectRemotePortsBoundOnLoopback(ctx context.Context, execClient exec.Client, podName string, containerName string, ports []api.ForwardedPort) ([]api.ForwardedPort, error) {
	if len(ports) == 0 {
		return nil, nil
	}

	listening, err := GetListeningConnections(ctx, execClient, podName, containerName)
	if err != nil {
		return nil, err
	}
	var boundToLocalhost []api.ForwardedPort
	for _, p := range ports {
		for _, conn := range listening {
			if p.ContainerPort != conn.LocalPort {
				continue
			}
			klog.V(6).Infof("found listening connection matching container port %d: %s", p.ContainerPort, conn.String())
			ip := net.ParseIP(conn.LocalAddress)
			if ip == nil {
				klog.V(6).Infof("invalid IP address: %q", conn.LocalAddress)
				continue
			}
			if ip.IsLoopback() {
				boundToLocalhost = append(boundToLocalhost, p)
				break
			}
		}
	}
	return boundToLocalhost, nil
}

// GetListeningConnections retrieves information about ports being listened and on which local address in the specified container.
// It works by parsing information from the /proc/net/{tcp,tcp6,udp,udp6} files, and is able to parse both IPv4 and IPv6 addresses.
// See https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt for more information about the structure of these files.
func GetListeningConnections(ctx context.Context, execClient exec.Client, podName string, containerName string) ([]Connection, error) {
	return GetConnections(ctx, execClient, podName, containerName, func(state int) bool {
		return stateToString(state) == "LISTEN"
	})
}

// GetConnections retrieves information about connections in the specified container.
// It works by parsing information from the /proc/net/{tcp,tcp6,udp,udp6} files, and is able to parse both IPv4 and IPv6 addresses.
// See https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt for more information about the structure of these files.
// The specified predicate allows to filter the connections based on the state.
func GetConnections(ctx context.Context, execClient exec.Client, podName string, containerName string, statePredicate func(state int) bool) ([]Connection, error) {
	cmd := []string{
		remotecmd.ShellExecutable, "-c",
		// /proc/net/{tc,ud}p6 files might be missing if IPv6 is disabled in the host networking stack.
		// Actually /proc/net/{tc,ud}p* files might be totally missing if network stats are disabled.
		"cat /proc/net/tcp /proc/net/udp /proc/net/tcp6 /proc/net/udp6 || true",
	}
	stdout, _, err := execClient.ExecuteCommand(ctx, cmd, podName, containerName, false, nil, nil)
	if err != nil {
		return nil, err
	}

	hexToInt := func(hex string) (int, error) {
		i, parseErr := strconv.ParseInt(hex, 16, 32)
		if parseErr != nil {
			return 0, parseErr
		}
		return int(i), nil
	}

	hexRevIpV4ToString := func(hex string) (string, error) {
		parts := ipv4HexRegExp.FindAllString(hex, -1)
		result := make([]string, 0, len(parts))
		for i := len(parts) - 1; i >= 0; i-- {
			toInt, parseErr := hexToInt(parts[i])
			if parseErr != nil {
				return "", parseErr
			}
			result = append(result, fmt.Sprintf("%d", toInt))
		}
		return strings.Join(result, "."), nil
	}

	hexRevIpV6ToString := func(hex string) (string, error) {
		// In IPv6, each group of the address is 2 bytes long (4 hex characters).
		// See https://www.rfc-editor.org/rfc/rfc4291#page-4
		i := []string{
			hex[30:32],
			hex[28:30],
			hex[26:28],
			hex[24:26],
			hex[22:24],
			hex[20:22],
			hex[18:20],
			hex[16:18],
			hex[14:16],
			hex[12:14],
			hex[10:12],
			hex[8:10],
			hex[6:8],
			hex[4:6],
			hex[2:4],
			hex[0:2],
		}
		return fmt.Sprintf("%s%s:%s%s:%s%s:%s%s:%s%s:%s%s:%s%s:%s%s",
			i[12], i[13], i[14], i[15],
			i[8], i[9], i[10], i[11],
			i[4], i[5], i[7], i[7],
			i[0], i[1], i[2], i[3]), nil
	}

	parseAddrAndPort := func(s string) (addr string, port int, err error) {
		addrPortList := strings.Split(s, ":")
		if len(addrPortList) != 2 {
			return "", 0, fmt.Errorf("invalid format - must be <addr>:<port>, but was %q", s)
		}

		addrHex := addrPortList[0]
		switch len(addrHex) {
		case 8:
			addr, err = hexRevIpV4ToString(addrHex)
		case 32:
			addr, err = hexRevIpV6ToString(addrHex)
		default:
			err = fmt.Errorf("length must be 8 (IPv4) or 32 (IPv6), but was %d", len(addrHex))
		}
		if err != nil {
			return "", 0, fmt.Errorf("could not decode address info from %q: %w", s, err)
		}

		portHex := addrPortList[1]
		port, err = hexToInt(portHex)
		if err != nil {
			return "", 0, fmt.Errorf("could not decode port info from %q: %w", s, err)
		}
		return addr, port, nil
	}

	var connections []Connection
	for _, l := range stdout {
		if strings.Contains(l, "local_address") {
			// ignore header lines
			continue
		}

		/*
			We are interested only in the first 4 values, which provide information about the local address, port and the connection state.
			See https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt

					   46: 010310AC:9C4C 030310AC:1770 01
					   |      |      |      |      |   |--> connection state
					   |      |      |      |      |------> remote TCP port number
					   |      |      |      |-------------> remote IPv4 address
					   |      |      |--------------------> local TCP port number
					   |      |---------------------------> local IPv4 address
					   |----------------------------------> number of entry
		*/
		split := strings.SplitN(strings.TrimSpace(l), " ", 5)
		if len(split) < 4 {
			klog.V(5).Infof("ignored line %q because it has less than 4 space-separated elements", l)
			continue
		}
		stateHex := split[3]
		state, err := hexToInt(stateHex)
		if err != nil {
			klog.V(5).Infof("[warn] could not decode state info from line %q: %v", l, err)
			continue
		}
		if statePredicate != nil && !statePredicate(state) {
			klog.V(5).Infof("ignored line because state value does not pass predicate: %q", l)
			continue
		}

		localAddr, localPort, err := parseAddrAndPort(split[1])
		if err != nil {
			klog.V(5).Infof("ignored line because it is not possible to determine local addr and port: %q", l)
			continue
		}
		remoteAddr, remotePort, err := parseAddrAndPort(split[2])
		if err != nil {
			klog.V(5).Infof("ignored line because it is not possible to determine remote addr and port: %q", l)
			continue
		}

		connections = append(connections, Connection{
			LocalAddress:  localAddr,
			LocalPort:     localPort,
			RemoteAddress: remoteAddr,
			RemotePort:    remotePort,
			State:         stateToString(state),
		})
	}

	return connections, nil
}

// CheckAppPortsListening checks whether all the specified ports are really opened and in LISTEN mode in each corresponding container
// of the pod specified.
// It does so by periodically looking inside the container for listening connections until it finds each of the specified ports,
// or until the specified timeout has elapsed.
func CheckAppPortsListening(
	ctx context.Context,
	execClient exec.Client,
	podName string,
	containerPortMapping map[string][]int,
	timeout time.Duration,
) error {
	if len(containerPortMapping) == 0 {
		return nil
	}

	backOffBase := 1 * time.Second
	if timeout <= backOffBase {
		return fmt.Errorf("invalid timeout: %v, must be strictly greater than %v", timeout, backOffBase)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hasPortFn := func(connections []Connection, p int) bool {
		for _, c := range connections {
			if p == c.LocalPort {
				return true
			}
		}
		return false
	}

	notListeningChan := make(chan map[string][]int)

	g := new(errgroup.Group)
	for container, ports := range containerPortMapping {
		container := container
		ports := ports

		if len(ports) == 0 {
			continue
		}

		g.Go(func() error {
			b := backo.NewBacko(backOffBase, 2, 0, 10*time.Second)
			ticker := b.NewTicker()
			portsNotListening := make(map[int]struct{})

			for {
				select {
				case <-ctxWithTimeout.Done():
					if len(portsNotListening) != 0 {
						m := make(map[string][]int)
						for p := range portsNotListening {
							m[container] = append(m[container], p)
						}
						notListeningChan <- m
					}
					return ctxWithTimeout.Err()

				case <-ticker.C:
					connections, err := GetListeningConnections(ctx, execClient, podName, container)
					if err != nil {
						klog.V(3).Infof("error getting listening connections in container %q: %v", container, err)
						for _, p := range ports {
							portsNotListening[p] = struct{}{}
						}
					} else {
						for _, p := range ports {
							if hasPortFn(connections, p) {
								delete(portsNotListening, p)
								continue
							}
							klog.V(3).Infof("port %d not listening in container %q", p, container)
							portsNotListening[p] = struct{}{}
						}
						if len(portsNotListening) == 0 {
							// no error and all ports expected to be opened are opened at this point
							return nil
						}
					}
				}
			}
		})
	}

	// Buffer of 1 because we want to close notListeningChan (because we are iterating over it).
	errChan := make(chan error, 1)
	go func() {
		errChan <- g.Wait()
		close(notListeningChan)
	}()

	notListening := make(map[string][]int)
	for e := range notListeningChan {
		for c, ports := range e {
			notListening[c] = append(notListening[c], ports...)
		}
	}

	klog.V(4).Infof("ports not listening: %v", notListening)

	if err := <-errChan; err != nil {
		msg := "error"
		if errors.Is(err, context.DeadlineExceeded) {
			msg = "timeout"
		}
		msg += " while checking for ports"
		if len(notListening) == 0 {
			klog.V(4).Infof("%s and no unreachable port detected: %v", msg, err)
			return nil
		}
		var msgList []string
		for c, ports := range notListening {
			var l []string
			for _, p := range ports {
				l = append(l, strconv.Itoa(p))
			}
			msgList = append(msgList, fmt.Sprintf("%s in container %q", strings.Join(l, ", "), c))
		}
		msg += fmt.Sprintf("; ports not listening: (%s)", strings.Join(msgList, "; "))
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func stateToString(state int) string {
	if state < 1 || state > len(connectionStates) {
		return ""
	}
	return connectionStates[state-1]
}
