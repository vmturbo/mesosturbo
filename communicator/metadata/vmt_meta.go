package metadata

import (
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
)

const (
	// Appliance address.

	// SERVER_ADDRESS string = ""

	TARGET_TYPE string = "Mesos"

	// NAME_OR_ADDRESS string = "k8s_vmt"

	USERNAME string = "mesos_user"

	TARGET_IDENTIFIER string = "my_mesos"

	PASSWORD string = "fake_password"

	// Ops Manager related
	// OPS_MGR_USRN = "administrator"
	// OPS_MGR_PSWD = "a"

	//WebSocket related
	LOCAL_ADDRESS    = "http://172.16.201.167/"
	WS_SERVER_USRN   = "vmtRemoteMediation"
	WS_SERVER_PASSWD = "vmtRemoteMediation"
)

type VMTMeta struct {
	MarathonIP         string
	MarathonPort       string
	MesosIP            string
	MesosPort          string
	ActionIP           string
	ActionPort         string
	ServerAddress      string
	TargetType         string
	NameOrAddress      string
	Username           string
	TargetIdentifier   string
	Password           string
	LocalAddress       string
	WebSocketUsername  string
	WebSocketPassword  string
	OpsManagerUsername string
	OpsManagerPassword string
	ActionAPI          string
}

// Create a new VMTMeta from file. ServerAddress, NameOrAddress of Kubernetes target, Ops Manager Username and
// Ops Manager Password should be set by user. Other fields have default values and can be overrided.
func NewVMTMeta(metaConfigFilePath string) (*VMTMeta, error) {
	glog.V(4).Infof("in newVMTMeta\n")
	meta := &VMTMeta{
		// ServerAddress:      SERVER_ADDRESS,
		TargetType: TARGET_TYPE,
		// NameOrAddress:      NAME_OR_ADDRESS,
		Username:          USERNAME,
		TargetIdentifier:  TARGET_IDENTIFIER,
		Password:          PASSWORD,
		LocalAddress:      LOCAL_ADDRESS,
		WebSocketUsername: WS_SERVER_USRN,
		WebSocketPassword: WS_SERVER_PASSWD,
		// OpsManagerUsername: OPS_MGR_USRN,
		// OpsManagerPassword: OPS_MGR_PSWD,
	}

	glog.V(4).Infof("Now read configration from %s", metaConfigFilePath)
	metaConfig := readConfig(metaConfigFilePath)

	if metaConfig.ActionIP != "" {
		meta.ActionIP = metaConfig.ActionIP
	} else {
		glog.V(4).Infof("Error getting LayerX Master\n")
		return nil, errors.New("Error getting LayerX Master.")
	}

	if metaConfig.ActionPort != "" {
		meta.ActionPort = metaConfig.ActionPort
	} else {
		glog.V(4).Infof("Error getting LayerX Master.\n")
		return nil, errors.New("error getting LayerX Master\n")
	}

	if metaConfig.ServerAddress != "" {
		meta.ServerAddress = metaConfig.ServerAddress
		glog.V(3).Infof("VMTurbo Server Address is %s", meta.ServerAddress)

	} else {
		glog.V(4).Infof("Error getting VMTurbo server address.")
		return nil, errors.New("Error getting VMTServer address\n")
	}

	if metaConfig.TargetIdentifier != "" {
		meta.TargetIdentifier = metaConfig.TargetIdentifier
	}
	glog.V(3).Infof("TargetIdentifier is %s", meta.TargetIdentifier)

	if metaConfig.NameOrAddress != "" {
		meta.NameOrAddress = metaConfig.NameOrAddress
		glog.V(3).Infof("NameOrAddress is %s", meta.NameOrAddress)
	} else {
		glog.Infof("Error getting NameorAddress for Mesos Probe.")
		return nil, errors.New("Error getting NameorAddress from Mesos Probe")
	}

	if metaConfig.Username != "" {
		meta.Username = metaConfig.Username
	}

	if metaConfig.TargetType != "" {
		meta.TargetType = metaConfig.TargetType
	}

	if metaConfig.Password != "" {
		meta.Password = metaConfig.Password
	}

	if metaConfig.LocalAddress != "" {
		meta.LocalAddress = metaConfig.LocalAddress
	}

	if metaConfig.OpsManagerUsername != "" {
		meta.OpsManagerUsername = metaConfig.OpsManagerUsername
		glog.V(3).Infof("OpsManagerUsername is %s", meta.OpsManagerUsername)
	} else {
		glog.V(4).Infof("Error getting VMTurbo Ops Manager Username.")
		return nil, errors.New("Error getting OpsManager Username")
	}

	if metaConfig.OpsManagerPassword != "" {
		meta.OpsManagerPassword = metaConfig.OpsManagerPassword
		glog.V(3).Infof("OpsManagerPassword is %s", meta.OpsManagerPassword)
	} else {
		glog.V(4).Infof("Error getting VMTurbo Ops Manager Password.")
		return nil, errors.New("Error getting OpsManager Password")
	}

	return meta, nil
}

// Get the config from file.
func readConfig(path string) VMTMeta {
	file, e := ioutil.ReadFile(path)
	if e != nil {
		glog.Errorf("File error: %v\n", e)
		os.Exit(1)
	}
	var metaData VMTMeta
	json.Unmarshal(file, &metaData)
	glog.V(4).Infof("Results: %v\n", metaData)
	return metaData
}
