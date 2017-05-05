package cluster

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/vektorlab/gaffer/cluster/service"
	"io/ioutil"
	"os"
)

// HostIDPath is the path where Gaffer stores
// a unique identifier for this host. If the
// identifier is deleted or changes the host
// will not be permitted to re-join the cluster.
const HostIDPath string = "/tmp/gaffer.id"

// ErrHostNotRegistered indicates this host is
// not registered with the host ID
type ErrHostNotRegistered struct {
	id string
}

func (e ErrHostNotRegistered) Error() string {
	return fmt.Sprintf("Host not registered (%s)", e.id)
}

// Host is unique server with one
// or more running processes
type Host struct {
	ID         string                      `json:"id"`
	Hostname   string                      `json:"hostname"`
	Registered bool                        `json:"registered"`
	Services   map[string]*service.Service `json:"services"`
}

func (h Host) me() bool {
	if h.ID == "" {
		return false
	}
	raw, err := ioutil.ReadFile(HostIDPath)
	if err != nil {
		return false
	}
	if string(raw) == h.ID {
		return true
	}
	return false
}

func (h *Host) Register() error {
	if !h.me() {
		return ErrHostNotRegistered{h.ID}
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	h.Hostname = hostname
	h.Registered = true
	return nil
}

func NewHost() *Host {
	return &Host{ID: uuid.NewV4().String()}
}