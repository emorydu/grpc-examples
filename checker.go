// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var (
		path string
	)

	flag.StringVar(&path, "path", "", "vpn configuration src filepath")

	flag.Parse()

	iptables, err := withdraw(path)
	if err != nil {
		panic(err)
	}
	c := NewVpnConf(iptables)

	c.Check()
}

type vpnConf struct {
	iptables []string
}

func (c vpnConf) Generate(cmd string) error {

	return nil
}

func (c vpnConf) Check() {
	for _, v := range c.iptables {
		if strings.Contains(v, "iptables") || strings.Contains(v, "ip6tables") {
			src := v
			checkCmd := strings.ReplaceAll(src, "-A", "-C")
			_, err := c.execute(checkCmd)
			if err == nil {
				continue
			} else {
				_, err = c.execute(src)
				if err != nil {
					log.Printf("generator iptables rule: %s: %v", src, err)
					continue
				}
			}
		} else {
			// route or rule
			err := c.RuleOrRoute(v)
			if err != nil {
				log.Printf("%s handle failed", v)
			}
		}
	}

}

const (
	Ipv6Rule  = "ip -6 rule"
	Ipv4Rule  = "ip -4 rule"
	Ipv6Route = "ip -6 route"
	Ipv4Route = "ip -4 route"

	Rule = "rule"
)

func (c vpnConf) RuleOrRoute(s string) error {
	identifier, tm := split(s)
	switch {
	case strings.Contains(s, Ipv6Route) || strings.Contains(s, Ipv4Route):
		prefix := Ipv6Route
		if strings.Contains(s, "-4") {
			prefix = Ipv4Route
		}
		exist, err := c.RouteOrRuleExist(identifier, prefix, tm)
		if err != nil {
			return fmt.Errorf("route exists execute error: %v", err)
		}
		if !exist {
			_, err := c.execute(s)
			if err != nil {
				return fmt.Errorf("generator ip route %s failed: %v", s, err)
			}
		}
	case strings.Contains(s, Ipv4Rule) || strings.Contains(s, Ipv6Rule):
		prefix := Ipv6Rule
		if strings.Contains(s, "-4") {
			prefix = Ipv4Rule
		}
		exist, err := c.RouteOrRuleExist("", prefix, tm)
		if err != nil {
			return fmt.Errorf("rule exists execute error: %v", err)
		}
		if !exist {
			_, err := c.execute(s)
			if err != nil {
				return fmt.Errorf("generator ip rule %s failed: %v", s, err)
			}
		}
	default:
		return fmt.Errorf("invalid rule or route")
	}

	return nil
}

func (c vpnConf) RouteOrRuleExist(identifier, prefix, name string) (bool, error) {
	var (
		out string
		err error
	)
	if strings.Contains(prefix, Rule) {
		s := fmt.Sprintf("%s show table %s", prefix, name)
		out, err = executeWithPipe(s, "grep 200")
	} else {
		out, err = executeWithPipe(fmt.Sprintf("%s show table %s", prefix, name), fmt.Sprintf("grep %s", identifier))
		//out, err = c.execute(fmt.Sprintf("%s show table %s | grep '%s'", prefix, name, identifier))
	}

	if err != nil {
		return false, err
	}
	if out == "" {
		return false, nil
	}
	return true, nil
}

func executeWithPipe(important, extra string) (string, error) {
	res := strings.Split(extra, "")

	mainCommand := exec.Command("/bin/bash", "-c", important)
	extraCommand := exec.Command(res[0], res[1])
	mainCommand.Stdin, _ = extraCommand.StdoutPipe()
	out, err := mainCommand.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execute command error: %w", err)
	}

	return string(out), nil
}

func split(s string) (identifier, last string) {
	vals := strings.Split(s, " ")
	if strings.Contains(vals[4], "/") {
		identifier = strings.Split(vals[4], "/")[0]
	} else {
		identifier = vals[4]
	}

	return identifier, vals[len(vals)-1]
}

func (c vpnConf) execute(cmd string) (string, error) {
	command := exec.Command("/bin/bash", "-c", cmd)
	out, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execute command error: %w", err)
	}

	return string(out), nil
}

func NewVpnConf(ips []string) vpnConf {
	return vpnConf{iptables: ips}
}

type Checker interface {
	Check()
}

type Generator interface {
	Generate() error
}

type CheckGenerator interface {
	Checker
	Generator
}

var ErrFormatPostUp = errors.New("postup format error")

func withdraw(path string) ([]string, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	lines := make([]string, 0, 1)

	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')

		if err == io.EOF || strings.Contains(line, "PostUp = ") {
			lines = append(lines, line)
			break
		}
		if err != nil {
			return nil, err
		}
	}

	if len(lines) != 1 {
		return nil, ErrFormatPostUp
	}

	iptables := make([]string, 0)
	s := strings.TrimPrefix(lines[0], "PostUp = ")
	for _, v := range strings.Split(s, ";") {
		if strings.HasPrefix(v, "ip -4 route del ") || strings.HasPrefix(v, "ip -6 route del") || v == "\n" {
			continue
		}
		iptables = append(iptables, v)
	}

	return iptables, nil
}
