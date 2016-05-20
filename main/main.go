package main

import (
	"fmt"
	"github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	"github.com/pamelasanchezvi/mesosturbo/communicator/vmtapi"
	"github.com/pamelasanchezvi/mesosturbo/communicator/vmturbocommunicator"
	"github.com/pamelasanchezvi/mesosturbo/pkg/action"
)

func main() {
	fmt.Println("in main now")

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
