package probe

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"strconv"
)

type TaskBuilder struct {
	ConstraintsMap map[string][][]string
	PortsMap       map[string][]string
}

func getPorts(portMappings []util.PortMapping) []string {
	ports := []string{}
	for _, portMapping := range portMappings {
		if &portMapping.HostPort != nil && portMapping.HostPort != 0 {
			ports = append(ports, strconv.Itoa(portMapping.HostPort))
		}
	}
	return ports
}

// Utility functions for building task probe objects
func (t *TaskBuilder) BuildConstraintMap(apps []util.App) map[string][][]string {
	constrMap := make(map[string][][]string)
	portsMap := make(map[string][]string)
	for _, app := range apps {
		var ports []string
		constrMap[app.Name] = app.Constraints
		if app.RequirePorts && app.Container.Docker.PortMappings != nil {
			ports = getPorts(app.Container.Docker.PortMappings)
		}
		portsMap[app.Name] = ports
		glog.V(2).Infof("---> app: %s ports are %+v", app.Name, app.Container.Docker.PortMappings)
	}
	t.ConstraintsMap = constrMap
	t.PortsMap = portsMap
	return t.ConstraintsMap
}

func (t *TaskBuilder) SetTaskConstraints(probe *TaskProbe) {
	probe.Constraints = t.ConstraintsMap[probe.Task.Name]
}
