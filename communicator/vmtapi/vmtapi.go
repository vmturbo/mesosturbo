package api

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/golang/glog"
	"github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	"github.com/pamelasanchezvi/mesosturbo/communicator/util"
	"github.com/pamelasanchezvi/mesosturbo/pkg/action"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var TEMPLATE_CPU_TINY = float64(1.0)
var TEMPLATE_MEM_TINY = float64(512.0)
var TEMPLATE_CPU_MICRO = float64(1.0)
var TEMPLATE_MEM_MICRO = float64(1024.0)
var TEMPLATE_CPU_SMALL = float64(1.0)
var TEMPLATE_MEM_SMALL = float64(2048.0)
var TEMPLATE_CPU_MEDIUM = float64(2.0)
var TEMPLATE_MEM_MEDIUM = float64(4096.0)
var TEMPLATE_CPU_LARGE = float64(2.0)
var TEMPLATE_MEM_LARGE = float64(8192.0)
var TEMPLATE_TINY_UUID = "DC5_1CxZMJkghjCaJOYu5"
var TEMPLATE_MICRO_UUID = "DC5_1CxZMJfgejCaJOYu5"
var TEMPLATE_SMALL_UUID = "DC5_1CxZMJkbfjCaJOYu5"
var TEMPLATE_MEDIUM_UUID = "DC5_1CxgeJkEEeCaJOYu5"
var TEMPLATE_LARGE_UUID = "DC5_1CxZMJkEEeCaJOYu5"

// api information for requests to VMT server
type VmtApi struct {
	vmtUrl    string
	extConfig map[string]string
}

// Metadata from configuration file
type Reservation struct {
	Meta *metadata.VMTMeta
}

const (
	logger = "VMTurbo API"
)

// Add a Mesos target to vmt ops manager
// example : http://localhost:8400/vmturbo/api/externaltargets?
//                     type=Mesos&nameOrAddress=10.10.150.2&username=AAA&targetIdentifier=A&password=Sysdreamworks123
func (vmtApi *VmtApi) AddMesosTarget(targetType, nameOrAddress, username, targetIdentifier, password string) error {
	glog.V(3).Infof("Calling VMTurbo REST API to added current %s target.", targetType)

	requestData := make(map[string]string)

	var requestDataBuffer bytes.Buffer

	requestData["type"] = targetType
	requestDataBuffer.WriteString("?type=")
	requestDataBuffer.WriteString(targetType)
	requestDataBuffer.WriteString("&")

	requestData["nameOrAddress"] = nameOrAddress
	requestDataBuffer.WriteString("nameOrAddress=")
	requestDataBuffer.WriteString(nameOrAddress)
	requestDataBuffer.WriteString("&")

	requestData["username"] = username
	requestDataBuffer.WriteString("username=")
	requestDataBuffer.WriteString(username)
	requestDataBuffer.WriteString("&")

	requestData["targetIdentifier"] = targetIdentifier
	requestDataBuffer.WriteString("targetIdentifier=")
	requestDataBuffer.WriteString(targetIdentifier)
	requestDataBuffer.WriteString("&")

	requestData["password"] = password
	requestDataBuffer.WriteString("password=")
	requestDataBuffer.WriteString(password)

	s := requestDataBuffer.String()

	respMsg, err := vmtApi.apiPost("/externaltargets", s)
	if err != nil {
		glog.V(4).Infof(" ERROR: %s", err)
		return err
	}
	glog.V(4).Infof("Add target response is %s", respMsg)

	return nil
}

// Discover a target using api
// http://localhost:8400/vmturbo/api/targets/mesos_vmt
func (vmtApi *VmtApi) DiscoverTarget(nameOrAddress string) error {
	glog.V(3).Info("Calling VMTurbo REST API to initiate a new discovery.")

	respMsg, err := vmtApi.apiPost("/targets/"+nameOrAddress, "")
	if err != nil {
		return err
	}
	glog.V(4).Infof("Discover target response is %s", respMsg)

	return nil
}

func (vmtApi *VmtApi) Post(postUrl, requestDataString string) (string, error) {
	return vmtApi.apiPost(postUrl, requestDataString)
}

func (vmtApi *VmtApi) Get(getUrl string) (string, error) {
	return vmtApi.apiGet(getUrl)
}

func (vmtApi *VmtApi) Delete(getUrl string) (string, error) {
	return vmtApi.apiDelete(getUrl)
}

// Call vmturbo api. return response
func (vmtApi *VmtApi) apiPost(postUrl, requestDataString string) (string, error) {
	fullUrl := "http://" + vmtApi.vmtUrl + "/vmturbo/api" + postUrl + requestDataString
	glog.V(4).Info("The full Url is ", fullUrl)
	fmt.Printf("The full Url is ", fullUrl)
	req, err := http.NewRequest("POST", fullUrl, nil)

	req.SetBasicAuth(vmtApi.extConfig["Username"], vmtApi.extConfig["Password"])
	glog.V(4).Info(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}

	respContent, err := parseAPICallResponse(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}
	glog.V(4).Infof("Post Succeed: %s", string(respContent))

	defer resp.Body.Close()
	return respContent, nil
}

// Call vmturbo api. return response
func (vmtApi *VmtApi) apiGet(getUrl string) (string, error) {
	fullUrl := "http://" + vmtApi.vmtUrl + "/vmturbo/api" + getUrl
	glog.V(4).Info("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)

	req.SetBasicAuth(vmtApi.extConfig["Username"], vmtApi.extConfig["Password"])
	glog.V(4).Info(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}
	respContent, err := parseAPICallResponse(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}
	glog.V(4).Infof("Get Succeed: %s", string(respContent))
	defer resp.Body.Close()
	return respContent, nil
}

// Delete API call
func (vmtApi *VmtApi) apiDelete(getUrl string) (string, error) {
	fullUrl := "http://" + vmtApi.vmtUrl + "/vmturbo/api" + getUrl
	glog.V(4).Info("The full Url is ", fullUrl)
	req, err := http.NewRequest("DELETE", fullUrl, nil)

	req.SetBasicAuth(vmtApi.extConfig["Username"], vmtApi.extConfig["Password"])
	glog.V(4).Info(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}
	respContent, err := parseAPICallResponse(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return "", err
	}
	glog.V(4).Infof("DELETE call Succeed: %s", string(respContent))
	defer resp.Body.Close()
	return respContent, nil
}

// this method takes in a reservation response and should return the reservation uuid, if there is any
func parseAPICallResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("response sent in is nil")
	}
	glog.V(4).Infof("response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Error after ioutil.ReadAll: %s", err)
		return "", err
	}
	glog.V(4).Infof("response content is %s", string(content))

	return string(content), nil
}

func (r *Reservation) GetVMTReservation(task *action.PendingTask) (string, error) {
	taskSpec := GetRequestTaskSpec(task)
	reservationResult, err := r.RequestPlacement(task.Name, taskSpec, nil)
	fmt.Printf(" -----> Requesting for task %s and spec %+v \n", task.Name, taskSpec)
	if err != nil {
		fmt.Printf("error in get Reservation line 214  %s\n", err)
		return "", err
	}
	return reservationResult, nil
}

// TODO for SDK
func getTemplateSize(pendingtask *action.PendingTask) string {
	template := ""
	taskCPU := pendingtask.CPUs
	taskMem := pendingtask.Mem
	if taskCPU < TEMPLATE_CPU_TINY || taskMem < TEMPLATE_MEM_TINY {
		template = TEMPLATE_TINY_UUID
	} else if taskCPU < TEMPLATE_CPU_MICRO || taskMem < TEMPLATE_MEM_MICRO {
		template = TEMPLATE_MICRO_UUID
	} else if taskCPU < TEMPLATE_CPU_SMALL || taskMem < TEMPLATE_MEM_SMALL {
		template = TEMPLATE_SMALL_UUID
	} else if taskCPU < TEMPLATE_CPU_MEDIUM || taskMem < TEMPLATE_MEM_MEDIUM {
		template = TEMPLATE_MEDIUM_UUID
	} else if taskCPU < TEMPLATE_CPU_LARGE || taskMem < TEMPLATE_MEM_LARGE {
		template = TEMPLATE_LARGE_UUID
	}
	return template
}

func GetRequestTaskSpec(task *action.PendingTask) map[string]string {
	// TODO for sdk how to get size for pod vs task
	templateUUIDs := getTemplateSize(task)
	requestMap := make(map[string]string)
	// TODO this name is not supposed to be hardcoded , same for Kubeturbo
	requestMap["reservation_name"] = "MesosReservationTest"
	requestMap["num_instances"] = "1"
	requestMap["template_name"] = templateUUIDs
	requestMap["templateUuids[]"] = templateUUIDs
	return requestMap
}

func buildReservationParameterString(requestSpec map[string]string) (string, error) {
	requestData := make(map[string]string)

	var requestDataBuffer bytes.Buffer

	if reservation_name, ok := requestSpec["reservation_name"]; !ok {
		glog.Errorf("reservation name is not registered")
		return "", fmt.Errorf("reservation_name has not been registered.")
	} else {
		requestData["reservationName"] = reservation_name
		requestDataBuffer.WriteString("?reservationName=")
		requestDataBuffer.WriteString(reservation_name)
		requestDataBuffer.WriteString("&")
	}

	if num_instances, ok := requestSpec["num_instances"]; !ok {
		glog.Errorf("num_instances not registered.")
		return "", fmt.Errorf("num_instances has not been registered.")
	} else {
		requestData["count"] = num_instances
		requestDataBuffer.WriteString("count=")
		requestDataBuffer.WriteString(num_instances)
		requestDataBuffer.WriteString("&")
	}

	if template_name, ok := requestSpec["template_name"]; !ok {
		glog.Errorf("template name is not registered")
		return "", fmt.Errorf("template_name has not been registered.")
	} else {
		requestData["templateName"] = template_name
		requestDataBuffer.WriteString("templateName=")
		requestDataBuffer.WriteString(template_name)
		requestDataBuffer.WriteString("&")
	}

	if templateUuids, ok := requestSpec["templateUuids[]"]; !ok {
		glog.Errorf("templateUuids is not specified.")
		return "", fmt.Errorf("templateUuids[] has not been registered.")
	} else {
		requestData["templateUuids[]"] = templateUuids
		requestDataBuffer.WriteString("templateUuids[]=")
		requestDataBuffer.WriteString(templateUuids)
	}

	s := requestDataBuffer.String()
	glog.V(4).Infof("parameters are %s", s)
	return s, nil
}

// put this in SDK TODO Pam Thursday
// Create the reservation specification and
// return map which has container name as key and slave name as value
func (this *Reservation) RequestPlacement(containerName string, requestSpec, filterProperties map[string]string) (string, error) {
	extCongfix := make(map[string]string)
	extCongfix["Username"] = this.Meta.OpsManagerUsername
	extCongfix["Password"] = this.Meta.OpsManagerPassword
	vmturboApi := NewVmtApi(this.Meta.ServerAddress, extCongfix)

	glog.V(4).Info("Inside RequestPlacement")
	fmt.Println("inside RequestPlacement")
	parameterString, err := buildReservationParameterString(requestSpec)
	if err != nil {
		return "", err
	}
	fmt.Printf("parameterString %s", parameterString)
	reservationUUID, err := vmturboApi.Post("/reservations", parameterString)
	if err != nil {
		return "", fmt.Errorf("Error posting reservations: %s", err)
	}
	reservationUUID = strings.Replace(reservationUUID, "\n", "", -1)
	glog.V(3).Infof("Reservation UUID is %s", string(reservationUUID))

	var getResponse string
	var getRevErr error
	// TODO, do we want to wait for a predefined time or send send API requests multiple times.
	for counter := 0; counter < 10; counter++ {
		time.Sleep(2 * time.Second)
		fmt.Printf("reserve UUID  %s", reservationUUID)
		getResponse, getRevErr = vmturboApi.Get("/reservations/" + reservationUUID)
		if getRevErr == nil {
			fmt.Printf("Got a reservation destination! \n")
			break
		} else {
			// handle the fact this reservation didn't get placed
		}
	}
	// After trying to get or getting the destination, delete the reservation.
	deleteResponse, err := vmturboApi.Delete("/reservations/" + reservationUUID)
	if err != nil {
		// TODO, Should we return without placement?
		return "", fmt.Errorf("Error deleting reservations destinations: %s response is : %s \n", err, deleteResponse)
	}
	//	glog.V(4).Infof("delete response of reservation %s is %s", reservationUUID, deleteResponse)
	if getRevErr != nil {
		return "", fmt.Errorf("Error getting reservations destinations: %s", err)
	}
	/*
		containerToSlaveMap, err := parseGetReservationResponse(containerName, getResponse)
		if err != nil {
			return nil, fmt.Errorf("Error parsing reservation destination returned from VMTurbo server: %s", err)
		}*/
	dest, err := GetTaskReservationDestination(getResponse)
	fullUrl := "http://" + this.Meta.MesosActionIP + ":5050" + "/state"
	fmt.Println("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)
	fmt.Println(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	respMap, err := util.CreateSlaveIpIdMap(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	fmt.Println("Get Succeed: %v", respMap)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("error line 342 vmtapi")
	}
	fmt.Printf("destination is: %s", dest)
	return respMap[dest], nil
}

// TODO change
type ServiceEntities struct {
	XMLName     xml.Name     `xml:"ServiceEntities"`
	ActionItems []ActionItem `xml:"ActionItem"`
}
type ActionItem struct {
	Datastore      string `xml:"datastore,attr"`
	DataStoreState string `xml:"datastoreState,attr"`
	Host           string `xml:"host,attr"`
	HostState      string `xml:"hostState,attr"`
	Name           string `xml:"name,attr"`
	Status         string `xml:"status,attr"`
	User           string `xml:"user,attr"`
	Vdc            string `xml:"vdc,attr"`
	VdcState       string `xml:"vdcState,attr"`
	VM             string `xml:"vm,attr"`
	VMState        string `xml:"vmState,attr"`
}

func decodeReservationResponse(content string) (*ServiceEntities, error) {
	// This is a temp solution. delete the encoding header.
	validStartIndex := strings.Index(content, ">")
	validContent := content[validStartIndex:]

	se := &ServiceEntities{}
	err := xml.Unmarshal([]byte(validContent), se)
	if err != nil {
		return nil, fmt.Errorf("Error decoding content: %s", err)
	}
	if se == nil {
		return nil, fmt.Errorf("Error decoding content. Result is null.")
	} else if len(se.ActionItems) < 1 {
		return nil, fmt.Errorf("Error decoding content. No ActionItem.")
	}
	return se, nil
}

func GetTaskReservationDestination(content string) (string, error) {
	se, err := decodeReservationResponse(content)
	if err != nil {
		return "", err
	}
	if se.ActionItems[0].VM == "" {
		return "", fmt.Errorf("Reservation destination get from VMT server is null.")
	}

	// Now only support a single reservation each time.
	return se.ActionItems[0].VM, nil
}

// this method takes in a http get response for reservation and should return the reservation uuid, if there is any
func parseGetReservationResponse(podName, content string) (map[string]string, error) {
	if content == "" {
		return nil, fmt.Errorf("No valid reservation result.")
	}
	// Decode reservation content.
	dest, err := GetTaskReservationDestination(content)
	if err != nil {
		return nil, err
	}
	glog.V(3).Infof("Deploy destination for Pod %s is %s", podName, dest)
	// TODO should parse the content. Currently don't know the correct get response content.
	pod2NodeMap := make(map[string]string)
	pod2NodeMap[podName] = dest
	return pod2NodeMap, nil
}

func NewVmtApi(url string, externalConfiguration map[string]string) *VmtApi {
	return &VmtApi{
		vmtUrl:    url,
		extConfig: externalConfiguration,
	}
}

// Watches for pending tasks through layerx requests
// Creates VMT reservation and placement request if pending tasks are found
func CreateWatcher(client *action.MesosClient, mesosmetadata *metadata.VMTMeta) {
	for {
		time.Sleep(time.Second * 20)
		pending, err := action.RequestPendingTasks(client)
		if err != nil {
			fmt.Printf("error %s \n", err)
		}

		if len(pending) > 0 {
			var taskDestinationMap = make(map[string]string)
			var newreservation *Reservation
			for i := range pending {
				fmt.Printf("pendingtasks are name:  %s and Id : %s \n", pending[i].Name, pending[i].Id)
				newreservation = &Reservation{
					Meta: mesosmetadata,
				}
				name := pending[i].Name
				taskDestinationMap[name], err = newreservation.GetVMTReservation(pending[i])
				if err != nil {
					fmt.Printf("Pending task %s is not getting placement, still pending.\n", pending[i].Name)
					continue
				}
				// assign Tasks if we got a placement
				client.Action = "AssignTasks"
				client.DestinationId = taskDestinationMap[name]
				client.TaskId = pending[i].Id
				res, err := action.RequestMesosAction(client)
				if err != nil {
					fmt.Printf("error %s \n", err)
				}
				fmt.Printf("result is : %s \n", res)
			}
		}
	}
}
