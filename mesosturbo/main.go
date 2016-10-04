package main

import (
	goflag "flag"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/vmturbo/mesosturbo/communicator/mesoshttp"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/communicator/vmturbocommunicator"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

var dcos_uid string
var dcos_pwd string
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
var slave_port string
var username string
var password string
var targetid string
var dcos_token string

func init() {
	goflag.Set("logtostderr", "true")
	goflag.StringVar(&dcos_uid, "dcos-uid", "", "dcos username")
	goflag.StringVar(&dcos_pwd, "dcos-pwd", "", "dcos password")
	goflag.StringVar(&dcos_token, "token", "", "dcos token")
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
	goflag.StringVar(&action_api, "action-api", "", "name of actions api")
	goflag.StringVar(&action_ip, "action-ip", "127.0.0.1", "ip for taking actions")
	goflag.StringVar(&action_port, "action-port", "5000", "port for taking actions")
	goflag.StringVar(&username, "username", "defaultUsername", "username")
	goflag.StringVar(&password, "password", "defaultPassword", "password")
	goflag.StringVar(&targetid, "target-identifier", "mesostarget", "identifier for target mesos cluster")
	goflag.StringVar(&slave_port, "slave-port", "5051", "port for taking actions")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
	goflag.CommandLine.Parse([]string{})
}

func main() {
	glog.V(3).Infof(" server ip: %s, mesos master ip : %s \n", vmt_server_ip, mesos_master_ip)

	metadata := &metadata.ConnectionClient{
		DCOS_Username:      dcos_uid,
		DCOS_Password:      dcos_pwd,
		MarathonIP:         marathon_ip,
		MarathonPort:       marathon_port,
		MesosIP:            mesos_master_ip,
		MesosPort:          mesos_port,
		ActionIP:           action_ip,
		ActionPort:         action_port,
		ServerAddress:      vmt_server_ip + ":" + vmt_server_port,
		TargetType:         target_type,
		NameOrAddress:      display_name,
		Username:           username,
		TargetIdentifier:   targetid,
		Password:           password,
		LocalAddress:       "http://" + local_address,
		WebSocketUsername:  "vmtRemoteMediation",
		WebSocketPassword:  "vmtRemoteMediation",
		OpsManagerUsername: opsman_username,
		OpsManagerPassword: opsman_password,
		ActionAPI:          action_api,
		SlavePort:          slave_port,
		DCOS:               false,
	}

	if dcos_uid != "" && dcos_pwd != "" {
		metadata.DCOS = true
	}

	if metadata.DCOS {
		// check DCOS username and password work
		mesosAPIClient := &mesoshttp.MesosHTTPClient{}
		err := mesosAPIClient.DCOSLoginRequest(metadata, dcos_token)
		if err != nil {
			return
		}
	}

	// Run Mesosturbo
	mesosClient := &action.MesosClient{
		ActionIP:   metadata.ActionIP,
		ActionPort: metadata.ActionPort,
		Action:     action_api,
	}

	glog.V(3).Infof("----> metadata is %+v", metadata)

	// layerx actions
	if action_api == "layerx" {
		go api.CreateWatcher(mesosClient, metadata)
	}

	clientMap := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(clientMap, metadata)
	comm.Run()
}
