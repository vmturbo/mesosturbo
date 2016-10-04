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

var dcosUID string
var dcosPwd string
var vmtServerIP string
var mesosMasterIP string
var mesosPort string
var targetType string
var localAddress string
var vmtServerPort string
var displayName string
var marathonIP string
var marathonPort string
var opsmanUsername string
var opsmanPassword string
var actionAPI string
var actionIP string
var actionPort string
var slavePort string
var username string
var password string
var targetid string
var dcosToken string

func init() {
	goflag.Set("logtostderr", "true")
	goflag.StringVar(&dcosUID, "dcos-uid", "", "dcos username")
	goflag.StringVar(&dcosPwd, "dcos-pwd", "", "dcos password")
	goflag.StringVar(&dcosToken, "token", "", "dcos token")
	goflag.StringVar(&displayName, "display-name", "mesos_default", "display name")
	goflag.StringVar(&vmtServerIP, "server-ip", "", "vmt server ip")
	goflag.StringVar(&vmtServerPort, "vmt-port", "80", "vmt port, default 8080")
	goflag.StringVar(&mesosMasterIP, "mesos-master-ip", "10.10.174.96", "mesos master ip")
	goflag.StringVar(&mesosPort, "mesos-port", "5050", "mesos master port")
	goflag.StringVar(&targetType, "target-type", "", "target type")
	goflag.StringVar(&localAddress, "local-address", "", "local address")
	goflag.StringVar(&marathonIP, "marathon-ip", "127.0.0.1", "ip of master that runs marathon ")
	goflag.StringVar(&marathonPort, "marathon-port", "8080", "port of master that runs marathon ")
	goflag.StringVar(&opsmanUsername, "ops-username", "", "ops manager username")
	goflag.StringVar(&opsmanPassword, "ops-password", "", "ops manager password")
	goflag.StringVar(&actionAPI, "action-api", "", "name of actions api")
	goflag.StringVar(&actionIP, "action-ip", "127.0.0.1", "ip for taking actions")
	goflag.StringVar(&actionPort, "action-port", "5000", "port for taking actions")
	goflag.StringVar(&username, "username", "defaultUsername", "username")
	goflag.StringVar(&password, "password", "defaultPassword", "password")
	goflag.StringVar(&targetid, "target-identifier", "mesostarget", "identifier for target mesos cluster")
	goflag.StringVar(&slavePort, "slave-port", "5051", "port for taking actions")
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()
	goflag.CommandLine.Parse([]string{})
}

func main() {
	glog.V(3).Infof(" server ip: %s, mesos master ip : %s \n", vmtServerIP, mesosMasterIP)

	metadata := &metadata.ConnectionClient{
		DCOS_Username:      dcosUID,
		DCOS_Password:      dcosPwd,
		MarathonIP:         marathonIP,
		MarathonPort:       marathonPort,
		MesosIP:            mesosMasterIP,
		MesosPort:          mesosPort,
		ActionIP:           actionIP,
		ActionPort:         actionPort,
		ServerAddress:      vmtServerIP + ":" + vmtServerPort,
		TargetType:         targetType,
		NameOrAddress:      displayName,
		Username:           username,
		TargetIdentifier:   targetid,
		Password:           password,
		LocalAddress:       "http://" + localAddress,
		WebSocketUsername:  "vmtRemoteMediation",
		WebSocketPassword:  "vmtRemoteMediation",
		OpsManagerUsername: opsmanUsername,
		OpsManagerPassword: opsmanPassword,
		ActionAPI:          actionAPI,
		SlavePort:          slavePort,
		DCOS:               false,
	}

	if dcosUID != "" && dcosPwd != "" {
		metadata.DCOS = true
	}

	if metadata.DCOS {
		// check DCOS username and password work
		mesosAPIClient := &mesoshttp.MesosHTTPClient{}
		err := mesosAPIClient.DCOSLoginRequest(metadata, dcosToken)
		if err != nil {
			return
		}
	}

	// Run Mesosturbo
	mesosClient := &action.MesosClient{
		ActionIP:   metadata.ActionIP,
		ActionPort: metadata.ActionPort,
		Action:     actionAPI,
	}

	glog.V(3).Infof("----> metadata is %+v", metadata)

	// layerx actions
	if actionAPI == "layerx" {
		go api.CreateWatcher(mesosClient, metadata)
	}

	clientMap := make(map[string]string)
	comm := vmturbocommunicator.NewVMTCommunicator(clientMap, metadata)
	comm.Run()
}
