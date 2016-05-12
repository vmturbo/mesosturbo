package builder

import (
	"github.com/pamelasanchezvi/mesosturbo/pkg/action"
)

type SimulatorBuilder struct {
	master string
	port   string
	action string
}

func NewSimulatorBuilder() *SimulatorBuilder {
	return &SimulatorBuilder{
		master: "10.10.174.96",
		port:   "5555",
		action: "movetask",
	}
}
func (sb *SimulatorBuilder) Build() (*action.MesosClient, error){
	actionpath := ""
	if sb.action == "movetask" {
		actionpath = "MigrateTasks"
		// TODO set destination node , tasks etc
	}
	return &action.MesosClient{
		MesosMasterIP:   sb.master,
		MesosMasterPort: sb.port,
		Action:          actionpath,
	}, nil
}
