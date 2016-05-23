package main

import (
	"fmt"

	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/communicator/vmturbocommunicator"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

func main() {
	metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	if err != nil {
		fmt.Printf("Error!! : %s\n", err)
	}
	mesosClient := &action.MesosClient{
		MesosMasterIP:   metadata.MesosActionIP,
		MesosMasterPort: metadata.MesosActionPort,
		Action:          "",
	}
	fmt.Printf("----> metadata is %+v", metadata)
	go api.CreateWatcher(mesosClient, metadata)
	clientMap := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(clientMap, metadata)
	comm.Run()
}
