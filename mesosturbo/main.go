package main

import (
	"bytes"
	"encoding/json"
	goflag "flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/communicator/vmturbocommunicator"
	"github.com/vmturbo/mesosturbo/pkg/action"
	"io/ioutil"
	"net/http"
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

	fmt.Printf("In main :)")
	fmt.Printf(" server ip: %s, mesos master ip : %s \n", vmt_server_ip, mesos_master_ip)

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
	}

	//metadata, err := metadata.NewVMTMeta("../communicator/metadata/config.json")
	/*
		if err != nil {
			fmt.Printf("Error!! : %s\n", err)
		}
	*/

	// check DCOS username and password work
	temporaryToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik9UQkVOakZFTWtWQ09VRTRPRVpGTlRNMFJrWXlRa015Tnprd1JrSkVRemRCTWpBM1FqYzVOZyJ9.eyJlbWFpbCI6Imphbm5sZW5vMUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiaXNzIjoiaHR0cHM6Ly9kY29zLmF1dGgwLmNvbS8iLCJzdWIiOiJnb29nbGUtb2F1dGgyfDEwMTU1OTUwNDc0NDg4OTU5OTQxNCIsImF1ZCI6IjN5RjVUT1N6ZGxJNDVRMXhzcHh6ZW9HQmU5Zk54bTltIiwiZXhwIjoxNDc0NzQ0NDk3LCJpYXQiOjE0NzQzMTI0OTd9.0_Mv7OycD9ZqXIif-bWDXfxewBLbpqnWNrfdo6rO1IDm39CH8D3jwffonkiwKKnoVo-GYXeayBfgasxUFSO9q2LtC-9c7Gr5RxBeYXaP3t9MHnIyFbO_kFTaCHSNU65atNaWV4bL0XRrAYFxxz3RMoA2z6hvh9cmjOlf_7X8YioPbJnLp-3mksNHIXf91yxavGgvgUvO_QUcMkVJ1FxBDYAAmOvzqKcHfmOECYoMxIOkxXH763W3ezP8_e0NHzAQypQAOuRDWRim761iz2fiIhiTfQnCBq2QvV1qOEjQYgj1vTHL9IA-Vxef17DaJ0zprJss2h9lPwWK7a0htDLpPw"

	dcos_token = temporaryToken
	var jsonStr []byte
	url := "http://dcos-turb-elasticl-1errcrm5owzjl-1572006965.us-east-1.elb.amazonaws.com/acs/api/v1/auth/login"

	if dcos_token == "" {
		fmt.Println(`{"uid":"` + dcos_uid + `","password":"` + dcos_pwd + `"}`)
		jsonStr = []byte(`{"uid":"` + dcos_uid + `","password":"` + dcos_pwd + `"}`)
	} else {
		fmt.Println(`{"uid":"` + dcos_uid + `","password":"` + dcos_pwd + `","token":"` + dcos_token + `"}`)
		jsonStr = []byte(`{"uid":"` + dcos_uid + `","password":"` + dcos_pwd + `","token":"` + temporaryToken + `"}`)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		// TODO check for Authorization error
		fmt.Printf("Error in POST request: %s", err)
		return
	}

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return
	}

	fmt.Printf("response content is %s", string(body))
	byteContent := []byte(body)
	var tokenResp = new(util.TokenResponse)
	err = json.Unmarshal(byteContent, &tokenResp)
	if err != nil {
		fmt.Printf("error in json unmarshal : %s", err)
	}

	metadata.Token = tokenResp.Token
	fmt.Println("--> token we got is ")
	fmt.Println(tokenResp.Token)
	/*url := "http://dcos-turb-elasticl-1errcrm5owzjl-1572006965.us-east-1.elb.amazonaws.com/mesos/master/state.json"

	payload := strings.NewReader("{\r\n   \"uid\":\"" + dcos_uid + "+\",\r\n   \"token\": \"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ik9UQkVOakZFTWtWQ09VRTRPRVpGTlRNMFJrWXlRa015Tnprd1JrSkVRemRCTWpBM1FqYzVOZyJ9.eyJlbWFpbCI6Imphbm5sZW5vMUBnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiaXNzIjoiaHR0cHM6Ly9kY29zLmF1dGgwLmNvbS8iLCJzdWIiOiJnb29nbGUtb2F1dGgyfDEwMTU1OTUwNDc0NDg4OTU5OTQxNCIsImF1ZCI6IjN5RjVUT1N6ZGxJNDVRMXhzcHh6ZW9HQmU5Zk54bTltIiwiZXhwIjoxNDc0NzQ0NDk3LCJpYXQiOjE0NzQzMTI0OTd9.0_Mv7OycD9ZqXIif-bWDXfxewBLbpqnWNrfdo6rO1IDm39CH8D3jwffonkiwKKnoVo-GYXeayBfgasxUFSO9q2LtC-9c7Gr5RxBeYXaP3t9MHnIyFbO_kFTaCHSNU65atNaWV4bL0XRrAYFxxz3RMoA2z6hvh9cmjOlf_7X8YioPbJnLp-3mksNHIXf91yxavGgvgUvO_QUcMkVJ1FxBDYAAmOvzqKcHfmOECYoMxIOkxXH763W3ezP8_e0NHzAQypQAOuRDWRim761iz2fiIhiTfQnCBq2QvV1qOEjQYgj1vTHL9IA-Vxef17DaJ0zprJss2h9lPwWK7a0htDLpPw\",\r\n   \"password\":\"Sysdreamworks123\"\r\n}")

	req, _ := http.NewRequest("GET", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "token=eyJhbGciOiJIUzI1NiIsImtpZCI6InNlY3JldCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIzeUY1VE9TemRsSTQ1UTF4c3B4emVvR0JlOWZOeG05bSIsImVtYWlsIjoiamFubmxlbm8xQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJleHAiOjEuNDc0NzQ0NDk3ZSswOSwiaWF0IjoxLjQ3NDMxMjQ5N2UrMDksImlzcyI6Imh0dHBzOi8vZGNvcy5hdXRoMC5jb20vIiwic3ViIjoiZ29vZ2xlLW9hdXRoMnwxMDE1NTk1MDQ3NDQ4ODk1OTk0MTQiLCJ1aWQiOiJqYW5ubGVubzFAZ21haWwuY29tIn0.SEZ96q2LQoXK6MoeJ7VN_XRUzV5fD_hWZghNXTasSbQ")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}

	glog.V(4).Infof("response content is %s", string(content))
	byteContent := []byte(content)
	var jsonMesosMaster = new(util.MesosAPIResponse)
	err = json.Unmarshal(byteContent, &jsonMesosMaster)
	if err != nil {
		glog.Errorf("error in json unmarshal : %s", err)
	}
	fmt.Println(res)
	fmt.Println(string(body))

	metadata.Token = ""
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
