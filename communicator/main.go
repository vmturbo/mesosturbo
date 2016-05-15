package main

import (
	"fmt"
	"github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	"github.com/pamelasanchezvi/mesosturbo/communicator/vmturbocommunicator"
)

func main() {
	fmt.Println("in main now")
	metadata, err := metadata.NewVMTMeta("./metadata/config.json")
	if err != nil {
		fmt.Println("error from metadata")
	}
	fmt.Printf("----> metadata is %+v", metadata )
	mesosClient := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(mesosClient, metadata)
	comm.Run()
}
