package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Exec func(args ...string) ([]byte, error)

type Client struct {
	run Exec
}

func New(run Exec) *Client {
	return &Client{run: run}
}

var ErrDaemonDown = errors.New("Docker Desktop não está rodando")

type Container struct {
	ID     string
	Name   string
	Image  string
	State  string
	Status string
	Ports  string
}

type rawContainer struct {
	ID     string `json:"ID"`
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	State  string `json:"State"`
	Status string `json:"Status"`
	Ports  string `json:"Ports"`
}

type Image struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	Size       string `json:"Size"`
}

func (c *Client) Info() (string, error) {
	out, err := c.run("info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return "", ErrDaemonDown
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) ListContainers() ([]Container, error) {
	out, err := c.run("ps", "-a", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	raws := parseLines[rawContainer](out)
	list := make([]Container, 0, len(raws))
	for _, r := range raws {
		list = append(list, Container{
			ID:     r.ID,
			Name:   firstName(r.Names),
			Image:  r.Image,
			State:  r.State,
			Status: r.Status,
			Ports:  r.Ports,
		})
	}
	return list, nil
}

func (c *Client) ListImages() ([]Image, error) {
	out, err := c.run("images", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	return parseLines[Image](out), nil
}

type ComposeService struct {
	Name     string
	Service  string
	State    string
	Status   string
	Health   string
	ExitCode int
	Ports    string
}

type publisher struct {
	URL           string `json:"URL"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
	Protocol      string `json:"Protocol"`
}

type rawComposeService struct {
	Name       string      `json:"Name"`
	Service    string      `json:"Service"`
	State      string      `json:"State"`
	Status     string      `json:"Status"`
	Health     string      `json:"Health"`
	ExitCode   int         `json:"ExitCode"`
	Publishers []publisher `json:"Publishers"`
}

func (c *Client) ComposePS() ([]ComposeService, error) {
	out, err := c.run("compose", "ps", "-a", "--format", "json")
	if err != nil {
		return nil, err
	}
	raws := parseComposeOutput(out)
	list := make([]ComposeService, 0, len(raws))
	for _, r := range raws {
		list = append(list, ComposeService{
			Name:     r.Name,
			Service:  r.Service,
			State:    r.State,
			Status:   r.Status,
			Health:   r.Health,
			ExitCode: r.ExitCode,
			Ports:    formatPublishers(r.Publishers),
		})
	}
	return list, nil
}

func parseComposeOutput(out []byte) []rawComposeService {
	if i := strings.IndexFunc(string(out), func(r rune) bool {
		return r != ' ' && r != '\t' && r != '\n' && r != '\r'
	}); i >= 0 && out[i] == '[' {
		var arr []rawComposeService
		if err := json.Unmarshal(out, &arr); err != nil {
			return nil
		}
		return arr
	}
	return parseLines[rawComposeService](out)
}

func formatPublishers(pubs []publisher) string {
	var parts []string
	for _, p := range pubs {
		if p.PublishedPort != 0 {
			parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", p.URL, p.PublishedPort, p.TargetPort, p.Protocol))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.TargetPort, p.Protocol))
		}
	}
	return strings.Join(parts, ", ")
}

func parseLines[T any](out []byte) []T {
	var list []T
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var v T
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			continue
		}
		list = append(list, v)
	}
	return list
}

func firstName(names string) string {
	if i := strings.IndexByte(names, ','); i >= 0 {
		return names[:i]
	}
	return names
}
