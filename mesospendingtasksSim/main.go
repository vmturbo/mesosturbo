package main

import (
	"fmt"
	"github.com/pamelasanchezvi/mesosturbo/cmd/simulation/builder"
	"github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	api "github.com/pamelasanchezvi/mesosturbo/communicator/vmtapi"
	"github.com/pamelasanchezvi/mesosturbo/pkg/action"
)

func main() {
	simbuilder := builder.NewSimulatorBuilder()
	// TODO flags
	// TODO init with flags
	simulator, err := simbuilder.Build()
	if err != nil {
		fmt.Printf("error %s \n", err)
	}
	//	actor := action.RequestMesosAction(simulator.MesosClient())
	pending, err := action.RequestPendingTasks(simulator)
	if err != nil {
		fmt.Printf("error %s \n", err)
	}
	metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	if err != nil {
		fmt.Println("error from metadata")
	}
	fmt.Printf("----> metadata is %+v", metadata)
	var taskDestinationMap = make(map[string]string)
	var newreservation *api.Reservation
	for i := range pending {
		fmt.Printf("pendingtasks are name:  %s and Id : %s \n", pending[i].Name, pending[i].Id)
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
			fmt.Printf("error %s \n", err)
		}
		fmt.Printf("result is : %s \n", res)
	}
}
