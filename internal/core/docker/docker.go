package docker

import (
	"encoding/json"
	"errors"
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
