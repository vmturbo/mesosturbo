package main

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/cmd/simulation/builder"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	api "github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

func main() {
	simbuilder := builder.NewSimulatorBuilder()
	// TODO flags
	// TODO init with flags
	simulator, err := simbuilder.Build()
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
	//	actor := action.RequestMesosAction(simulator.MesosClient())
	pending, err := action.RequestPendingTasks(simulator)
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
	metadata, err := metadata.NewConnectionClient("../communicator/metadata/config.json")
	if err != nil {
		glog.Errorf("error from metadata")
	}
	glog.V(4).Infof("----> metadata is %+v \n", metadata)
	var taskDestinationMap = make(map[string]string)
	var newreservation *api.Reservation
	for i := range pending {
		glog.V(3).Infof("Pendingtasks are name:  %s and Id : %s \n", pending[i].Name, pending[i].Id)
		newreservation = &api.Reservation{
			Meta: metadata,
		}
		name := pending[i].Name
		taskDestinationMap[name] = newreservation.GetVMTReservation(pending[i])
		// assign Tasks
		client := action.MesosClient{
			MesosMasterIP:   "10.10.174.96",
			MesosMasterPort: "5555",
			Action:          "AssignTasks",
			DestinationId:   taskDestinationMap[name],
			TaskId:          pending[i].Id,
		}
		res, err := action.RequestMesosAction(&client)
		if err != nil {
			glog.Errorf("error %s \n", err)
		}
		glog.V(4).Infof("result is : %s \n", res)
	}
}
