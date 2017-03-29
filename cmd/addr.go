package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/ddn0/go-dockerclient"
	"github.com/ddn0/peanut/dnet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addrCmd = &cobra.Command{
	Use:   "addr <port>",
	Short: "show address information for a docker service",
	RunE:  runAddr,
}

type Port struct {
	Port  int
	Proto string
}

func (a Port) String() string {
	return fmt.Sprintf("%d/%s", a.Port, a.Proto)
}

type Address struct {
	net.Addr           // Network address
	ServiceName string // Service name
	ServicePort Port   // Nominal service port
	GuessedHost bool   // Failed to find an exact match for service query
}

func (a Address) Convert() MarshalAddress {
	var ip string
	var port int
	var proto string
	if i, ok := a.Addr.(*net.TCPAddr); ok {
		ip = i.IP.String()
		port = i.Port
		proto = "tcp"
	} else if i, ok := a.Addr.(*net.UDPAddr); ok {
		ip = i.IP.String()
		port = i.Port
		proto = "udp"
	}
	return MarshalAddress{
		IP:          ip,
		Port:        port,
		Proto:       proto,
		ServiceName: a.ServiceName,
		ServicePort: a.ServicePort,
		GuessedHost: a.GuessedHost,
	}
}

// Stable version of Address suitable for marhsalling/output
type MarshalAddress struct {
	IP          string
	Port        int
	Proto       string
	ServiceName string
	ServicePort Port
	GuessedHost bool
}

func (a MarshalAddress) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func envAddrs() (addrs []Address) {
	r := regexp.MustCompile(`(.*)_PORT_(\d+)_(UDP|TCP)`)
	for _, v := range os.Environ() {
		splits := strings.SplitN(v, "=", 2)
		key := splits[0]
		value := splits[1]

		matches := r.FindStringSubmatch(key)
		if len(matches) == 0 {
			continue
		}

		u, err := url.Parse(value)
		if err != nil {
			continue
		}

		port, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		var addr net.Addr
		switch matches[3] {
		case "UDP":
			addr, err = net.ResolveUDPAddr(u.Scheme, u.Host)
		case "TCP":
			addr, err = net.ResolveTCPAddr(u.Scheme, u.Host)
		}
		if err != nil || addr == nil {
			continue
		}
		addrs = append(addrs, Address{
			Addr:        addr,
			ServiceName: matches[1],
			ServicePort: Port{
				Port:  port,
				Proto: strings.ToLower(matches[3]),
			},
		})
	}
	return
}

func dockerAddrs() (addrs []Address) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return
	}
	cs, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return
	}

	myIp, err := dnet.DockerIp()
	if err != nil {
		myIp, err = dnet.MyIpAddr()
	}
	if err != nil {
		return
	}

	for _, c := range cs {
		cc, err := client.InspectContainer(c.ID)
		if err != nil {
			return
		}
		for port, bindings := range cc.NetworkSettings.Ports {
			sproto := strings.ToLower(port.Proto())
			sport, err := strconv.Atoi(port.Port())
			if err != nil {
				continue
			}

			for _, b := range bindings {
				hp, err := strconv.Atoi(b.HostPort)
				if err != nil {
					continue
				}
				var addr net.Addr
				switch strings.ToUpper(port.Proto()) {
				case "UDP":
					addr = &net.UDPAddr{
						IP:   myIp,
						Port: hp,
					}
				case "TCP":
					addr = &net.TCPAddr{
						IP:   myIp,
						Port: hp,
					}
				}
				if addr == nil {
					continue
				}

				addrs = append(addrs, Address{
					Addr:        addr,
					ServiceName: path.Base(cc.Name),
					ServicePort: Port{
						Port:  sport,
						Proto: sproto,
					},
				})
			}
		}
	}

	return
}

func getAddrs() (addrs []Address) {
	addrs = append(addrs, envAddrs()...)
	addrs = append(addrs, dockerAddrs()...)
	return
}

func matchingAddrs(port Port, service string) (Address, error) {
	addrs := getAddrs()

	for _, addr := range addrs {
		switch {
		case addr.ServicePort == port && addr.ServiceName == service:
			return addr, nil
		case addr.ServicePort == port && service == "":
			return addr, nil
		}
	}
	return Address{}, fmt.Errorf("no matching service found")
}

func parsePort(s string) (Port, error) {
	splits := strings.SplitN(s, "/", 2)

	pstr := s
	proto := "tcp"
	if len(splits) == 2 {
		pstr = splits[0]
		proto = strings.ToLower(splits[1])
	}

	if p, err := strconv.Atoi(pstr); err != nil {
		return Port{}, err
	} else {
		return Port{
			Port:  p,
			Proto: proto,
		}, nil
	}
}

func runAddr(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("need a port")
	}

	port, err := parsePort(args[0])
	if err != nil {
		return err
	}

	addr, err := matchingAddrs(port, viper.GetString("container"))
	if err != nil {
		return err
	}

	return print(addr.Convert(), viper.GetString("format"), viper.GetString("filter"))
}

func init() {
	c := addrCmd
	flags := c.Flags()

	RootCmd.AddCommand(c)
	flags.String("container", "", "If multiple ports match, return the one the named container")
	flags.String("format", "text", "Output format {json,text,yaml}")
	flags.String("filter", "", "Filter in the syntax of the go package template")
}
