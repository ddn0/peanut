package dnet

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Return a non-loopback IP address for this machine, preferring IPv4 over IPv6
func MyIpAddr() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var six net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.IsLoopback() {
				continue
			} else if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			} else if six != nil && ipnet.IP.To16() != nil {
				six = ipnet.IP
			}
		}
	}
	return six, nil
}

func DockerIp() (net.IP, error) {
	if e := os.Getenv("DOCKER_HOST"); len(e) == 0 {
	} else if u, err := url.Parse(e); err == nil {
		if addr, err := net.ResolveTCPAddr("tcp", u.Host); err != nil {
			return nil, err
		} else {
			return addr.IP, nil
		}
	}
	return MyIpAddr()
}

// Returns the docker-compose environment bindings that would have been added
// by docker compose/docker run --link
//
// https://docs.docker.com/compose/env/
func ComposeEnv(proto, port, hip, hport string) []string {
	var env []string
	uproto := strings.ToUpper(proto)
	env = append(env, fmt.Sprintf("PORT=%s://%s:%s", proto, hip, hport))
	env = append(env, fmt.Sprintf("PORT_%s_%s=%s://%s:%s", port, uproto, proto, hip, hport))
	env = append(env, fmt.Sprintf("PORT_%s_%s_ADDR=%s", port, uproto, hip))
	env = append(env, fmt.Sprintf("PORT_%s_%s_PORT=%s", port, uproto, hport))
	env = append(env, fmt.Sprintf("PORT_%s_%s_PROTO=%s", port, uproto, proto))
	return env
}
