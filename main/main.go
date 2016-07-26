package main

import (
	goflag "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/communicator/vmturbocommunicator"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

var vmt_server_ip string
var mesos_master_ip string
var target_type string

func init() {
	/*
		serverIpFlag := &goflag.Flag{
			Name:  "server-ip",
			Usage: "the vmt server IP address",
		}
		mesosMasterFlag := &goflag.Flag{
			Name:  "mesosMaster-ip",
			Usage: "the IP address of the mesos cluster master",
		}*/
	goflag.Set("logtostderr", "true")
	goflag.StringVar(&vmt_server_ip, "server-ip", "", "vmt server ip")
	goflag.StringVar(&mesos_master_ip, "mesos-master-ip", "", "mesos master ip")
	goflag.StringVar(&target_type, "target-type", "", "target type")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
}

func main() {

	var actions string
	fmt.Printf("In main :)")
	fmt.Printf(" server: %s, master : %s \n", vmt_server_ip, mesos_master_ip)
	fmt.Println("flags: %+v", goflag.Args())

	metadata := &metadata.VMTMeta{
		MesosActionIP:      mesos_master_ip,
		MesosActionPort:    "0",
		ServerAddress:      vmt_server_ip,
		TargetType:         target_type,
		NameOrAddress:      "mesos1",
		Username:           "mesos_user",
		TargetIdentifier:   "mesos1",
		Password:           "password",
		LocalAddress:       "localhost",
		WebSocketUsername:  "vmtRemoteMediation",
		WebSocketPassword:  "vmtRemoteMediation",
		OpsManagerUsername: "administrator",
		OpsManagerPassword: "a",
	}

	//metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")

	//if err != nil {
	//	fmt.Printf("Error!! : %s\n", err)
	//}

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
