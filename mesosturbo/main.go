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
var mesos_port string
var target_type string
var local_address string
var vmt_server_port string
var display_name string
var marathon_ip string
var marathon_port string
var opsman_username string
var opsman_password string
var action_api string
var action_ip string
var action_port string

func init() {
	goflag.Set("logtostderr", "true")
	goflag.StringVar(&display_name, "display-name", "mesos_default", "display name")
	goflag.StringVar(&vmt_server_ip, "server-ip", "", "vmt server ip")
	goflag.StringVar(&vmt_server_port, "vmt-port", "80", "vmt port, default 8080")
	goflag.StringVar(&mesos_master_ip, "mesos-master-ip", "10.10.174.96", "mesos master ip")
	goflag.StringVar(&mesos_port, "mesos-port", "5050", "mesos master port")
	goflag.StringVar(&target_type, "target-type", "", "target type")
	goflag.StringVar(&local_address, "local-address", "", "local address")
	goflag.StringVar(&marathon_ip, "marathon-ip", "127.0.0.1", "ip of master that runs marathon ")
	goflag.StringVar(&marathon_port, "marathon-port", "8080", "port of master that runs marathon ")
	goflag.StringVar(&opsman_username, "ops-username", "", "ops manager username")
	goflag.StringVar(&opsman_password, "ops-password", "", "ops manager password")
	goflag.StringVar(&action_api, "action-api", "layerx", "name of actions api")
	goflag.StringVar(&action_ip, "action-ip", "127.0.0.1", "ip for taking actions")
	goflag.StringVar(&action_port, "action-port", "5000", "port for taking actions")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
}

func main() {

	fmt.Printf("In main :)")
	fmt.Printf(" server ip: %s, mesos master ip : %s \n", vmt_server_ip, mesos_master_ip)

	metadata := &metadata.VMTMeta{
		MarathonIP:         marathon_ip,
		MarathonPort:       marathon_port,
		MesosIP:            mesos_master_ip,
		MesosPort:          mesos_port,
		ActionIP:           action_ip,
		ActionPort:         action_port,
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
		ActionAPI:          "layerx",
	}

	//metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	/*
		if err != nil {
			fmt.Printf("Error!! : %s\n", err)
		}
	*/
	mesosClient := &action.MesosClient{
		ActionIP:   metadata.ActionIP,
		ActionPort: metadata.ActionPort,
		Action:     "layerx",
	}

	fmt.Printf("----> metadata is %+v", metadata)

	// layerx actions
	if action_api == "layerx" {
		go api.CreateWatcher(mesosClient, metadata)
	}

	clientMap := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(clientMap, metadata)
	comm.Run()
}
