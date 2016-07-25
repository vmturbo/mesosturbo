package main

import (
	goflag "flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/communicator/vmturbocommunicator"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

func init() {
	goflag.Set("logtostderr", "true")

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
}

func main() {
	var actions string
	glog.Infof("In main :)")
	metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	if err != nil {
		glog.Infof("Error!! : %s\n", err)
	}
	mesosClient := &action.MesosClient{
		MesosMasterIP:   metadata.MesosActionIP,
		MesosMasterPort: metadata.MesosActionPort,
		Action:          "",
	}
	fmt.Printf("----> metadata is %+v", metadata)
	// Pending Task watcher
	// TODO flag
	actions = "disabled"
	if actions == "enabled" {
		go api.CreateWatcher(mesosClient, metadata)
	}
	clientMap := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(clientMap, metadata)
	comm.Run()
}
