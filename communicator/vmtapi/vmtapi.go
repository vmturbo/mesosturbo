package api

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	"io/ioutil"
	"net/http"
)

type VmtApi struct {
	vmtUrl    string
	extConfig map[string]string
}

type Reservation struct {
	Meta *metadata.VMTMeta
}

const (
	logger = "VMTurbo API"
)

// Add a Kuberenets target to vmt ops manager
// example : http://localhost:8400/vmturbo/api/externaltargets?
//                     type=Kubernetes&nameOrAddress=10.10.150.2&username=AAA&targetIdentifier=A&password=Sysdreamworks123
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

func (r *Reservation) GetVMTReservation(task *util.Task) map[string]string {
	taskSpec := GetRequestTaskSpec(task)
	reservationResult, err := r.RequestPlacement(task.Name, taskSpec, nil)

}

// TODO for SDK ?
func getTemplateSize(task *util.Task) string {
	template := ""
	taskCPU := "2"
	taskMem := "2000"
	if taskCPU < TEMPLATE_CPU_TINY || taskMem < TEMPLATE_MEM_TINY {
		template = TEMPLATE_TINY_UUID
	}
	return template
}

func GetRequestTaskSpec(task *util.Task) map[string]strinig {
	// TODO for sdk how to get size for pod vs task
	templateUUIDs := getTemplateSize(task)
	requestMap := make(map[string]string)
	requestMap["reservation_name"] = "MesosReservationTest"
	requestMap["num_instances"] = "1"
	requestMap["templateName"] = templateUUIDs
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
func (this *Reservation) RequestPlacement(containerName string, requestSpec, filterProperties map[string]string) (map[string]string, error) {
	extCongfix := make(map[string]string)
	extCongfix["Username"] = this.Meta.OpsManagerUsername
	extCongfix["Password"] = this.Meta.OpsManagerPassword
	vmturboApi := vmtapi.NewVmtApi(this.Meta.ServerAddress, extCongfix)

	glog.V(4).Info("Inside RequestPlacement")
	fmt.Println("inside RequestPlacement")
	parameterString, err := buildReservationParameterString(requestSpec)
	if err != nil {
		return nil, err
	}

	reservationUUID, err := vmturboApi.Post("/reservations", parameterString)
	if err != nil {
		return nil, fmt.Errorf("Error posting reservations: %s", err)
	}
	reservationUUID = strings.Replace(reservationUUID, "\n", "", -1)
	glog.V(3).Infof("Reservation UUID is %s", string(reservationUUID))

	// TODO, do we want to wait for a predefined time or send send API requests multiple times.
	time.Sleep(2 * time.Second)
	getResponse, getRevErr := vmturboApi.Get("/reservations/" + reservationUUID)
	// After getting the destination, delete the reservation.
	deleteResponse, err := vmturboApi.Delete("/reservations/" + reservationUUID)
	if err != nil {
		// TODO, Should we return without placement?
		return nil, fmt.Errorf("Error deleting reservations destinations: %s", err)
	}
	glog.V(4).Infof("delete response of reservation %s is %s", reservationUUID, deleteResponse)
	if getRevErr != nil {
		return nil, fmt.Errorf("Error getting reservations destinations: %s", err)
	}
	pod2nodeMap, err := parseGetReservationResponse(podName, getResponse)
	if err != nil {
		return nil, fmt.Errorf("Error parsing reservation destination returned from VMTurbo server: %s", err)
	}

	return pod2nodeMap, nil
}

func NewVmtApi(url string, externalConfiguration map[string]string) *VmtApi {
	return &VmtApi{
		vmtUrl:    url,
		extConfig: externalConfiguration,
	}
}
