package docker

import (
	"encoding/json"
	"strings"
)

type Exec func(args ...string) ([]byte, error)

type Client struct {
	run Exec
}

func New(run Exec) *Client {
	return &Client{run: run}
}

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

func (c *Client) ListContainers() ([]Container, error) {
	out, err := c.run("ps", "-a", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	return parseContainers(out), nil
}

func parseContainers(out []byte) []Container {
	var list []Container
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var r rawContainer
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		list = append(list, Container{
			ID:     r.ID,
			Name:   firstName(r.Names),
			Image:  r.Image,
			State:  r.State,
			Status: r.Status,
			Ports:  r.Ports,
		})
	}
	return list
}

func firstName(names string) string {
	if i := strings.IndexByte(names, ','); i >= 0 {
		return names[:i]
	}
	return names
}
