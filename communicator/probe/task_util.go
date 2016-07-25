package probe

import (
	"github.com/vmturbo/mesosturbo/communicator/util"
)

type TaskBuilder struct {
	ConstraintsMap map[string][][]string
}

// Utility functions for building task probe objects
func (t *TaskBuilder) BuildConstraintMap(apps []util.App) map[string][][]string {
	constrMap := make(map[string][][]string)
	for _, app := range apps {
		constrMap[app.Name] = app.Constraints
	}
	t.ConstraintsMap = constrMap
	return t.ConstraintsMap
}

func (t *TaskBuilder) SetTaskConstraints(probe *TaskProbe) {
	probe.Constraints = t.ConstraintsMap[probe.Task.Name]
}
