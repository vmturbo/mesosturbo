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
var local_address string
var vmt_server_port string
var display_name string
var marathon_master string

func init() {
	goflag.Set("logtostderr", "true")
	goflag.StringVar(&display_name, "display-name", "mesos_default", "display name")
	goflag.StringVar(&vmt_server_ip, "server-ip", "", "vmt server ip")
	goflag.StringVar(&vmt_server_port, "vmt-port", "8080", "vmt port, default 8080")
	goflag.StringVar(&mesos_master_ip, "mesos-master-ip", "", "mesos master ip")
	goflag.StringVar(&target_type, "target-type", "", "target type")
	goflag.StringVar(&local_address, "local-address", "", "local address")
	goflag.StringVar(&marathon_master, "marathon-master", "", "ip of master that runs marathon ")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
}

func main() {

	var actions string

	fmt.Printf("In main :)")
	fmt.Printf(" server ip: %s, mesos master ip : %s \n", vmt_server_ip, mesos_master_ip)

	metadata := &metadata.VMTMeta{
		MesosMarathonIP:    marathon_master,
		MesosActionIP:      mesos_master_ip,
		MesosActionPort:    "5555",
		ServerAddress:      vmt_server_ip + ":" + vmt_server_port,
		TargetType:         target_type,
		NameOrAddress:      display_name,
		Username:           "mesos_user",
		TargetIdentifier:   "mesos1",
		Password:           "password",
		LocalAddress:       "http://" + local_address,
		WebSocketUsername:  "vmtRemoteMediation",
		WebSocketPassword:  "vmtRemoteMediation",
		OpsManagerUsername: "administrator",
		OpsManagerPassword: "a",
	}

	//metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	/*
		if err != nil {
			fmt.Printf("Error!! : %s\n", err)
		}
	*/
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
