package vmturbocommunicator

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	vmtmeta "github.com/pamelasanchezvi/mesosturbo/communicator/metadata"
	"github.com/pamelasanchezvi/mesosturbo/communicator/probe"
	"github.com/pamelasanchezvi/mesosturbo/communicator/util"
	vmtapi "github.com/pamelasanchezvi/mesosturbo/communicator/vmtapi"
	"github.com/pamelasanchezvi/mesosturbo/pkg/action"
	comm "github.com/vmturbo/vmturbo-go-sdk/communicator"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// impletements sdk.ServerMessageHandler
type MesosServerMessageHandler struct {
	//	mesosClient *client.Client
	meta   *vmtmeta.VMTMeta
	wsComm *comm.WebSocketCommunicator
	//etcdStorage storage.Storage
	lastDiscoveryTime *time.Time
	slaveUseMap       map[string]*util.CalculatedUse
}

// Use the vmt restAPI to add a Kubernetes target.
func (handler *MesosServerMessageHandler) AddTarget() {
	glog.V(4).Infof("------> in AddTarget()")
	vmtUrl := handler.wsComm.VmtServerAddress

	extCongfix := make(map[string]string)
	extCongfix["Username"] = handler.meta.OpsManagerUsername
	extCongfix["Password"] = handler.meta.OpsManagerPassword
	vmturboApi := vmtapi.NewVmtApi(vmtUrl, extCongfix)

	// Add Kubernetes target.
	// targetType, nameOrAddress, targetIdentifier, password
	vmturboApi.AddMesosTarget(handler.meta.TargetType, handler.meta.NameOrAddress, handler.meta.Username, handler.meta.TargetIdentifier, handler.meta.Password)
}

// Send an API request to make server start a discovery process on current Mesos
func (handler *MesosServerMessageHandler) DiscoverTarget() {
	vmtUrl := handler.wsComm.VmtServerAddress

	extCongfix := make(map[string]string)
	extCongfix["Username"] = handler.meta.OpsManagerUsername
	extCongfix["Password"] = handler.meta.OpsManagerPassword
	vmturboApi := vmtapi.NewVmtApi(vmtUrl, extCongfix)

	// Discover Kubernetes target.
	vmturboApi.DiscoverTarget(handler.meta.NameOrAddress)
}

// If server sends a validation request, validate the request.
// TODO, for now k8s validate all the request. aka, no matter what usr/passwd is provided, always pass validation.
// The correct bahavior is to set ErrorDTO when validation fails.
func (handler *MesosServerMessageHandler) Validate(serverMsg *comm.MediationServerMessage) {
	//Always send Validated for now
	glog.V(3).Infof("Kubernetes validation request from Server")

	// 1. Get message ID.
	messageID := serverMsg.GetMessageID()
	// 2. Build validationResponse.
	validationResponse := new(comm.ValidationResponse)
	// 3. Create client message with ClientMessageBuilder.
	clientMsg := comm.NewClientMessageBuilder(messageID).SetValidationResponse(validationResponse).Create()
	handler.wsComm.SendClientMessage(clientMsg)

	// TODO: Need to sleep some time, waiting validated. Or we should add reponse msg from server.
	time.Sleep(100 * time.Millisecond)
	glog.V(3).Infof("Discovery Target after validation")

	handler.DiscoverTarget()
}

func (handler *MesosServerMessageHandler) keepDiscoverAlive(messageID int32) {
	//
	glog.V(3).Infof("Keep Alive")

	keepAliveMsg := &comm.KeepAlive{}
	clientMsg := comm.NewClientMessageBuilder(messageID).SetKeepAlive(keepAliveMsg).Create()

	handler.wsComm.SendClientMessage(clientMsg)
}

// DiscoverTopology receives a discovery request from server and start probing the Mesos.
func (handler *MesosServerMessageHandler) DiscoverTopology(serverMsg *comm.MediationServerMessage) {
	//lastTime := *handler.lastDiscoveryTime
	//currentTime := time.Now()
	//duration := time.Since(lastTime)
	//secondsSinceLastDiscovery := duration.Seconds()
	//Discover the Mesos topology
	glog.V(3).Infof("Discover topology request from server.")

	// 1. Get message ID
	messageID := serverMsg.GetMessageID()
	var stopCh chan struct{} = make(chan struct{})
	go util.Until(func() { handler.keepDiscoverAlive(messageID) }, time.Second*10, stopCh)
	defer close(stopCh)

	// 2. Build discoverResponse
	mesosProbe, err := handler.NewMesosProbe(handler.slaveUseMap)
	//	mesosProbe.TimeSinceLastDisc = secondsSinceLastDiscovery
	res := mesosProbe.Slaves[0].Resources
	fmt.Printf("at Discover topology: disk %f, mem %f , cpu %f  \n", res.Disk, res.Mem, res.CPUs)
	nodeEntityDtos, err := ParseNode(mesosProbe, mesosProbe.SlaveUseMap)
	fmt.Printf(" slave name is %s \n", nodeEntityDtos[0].DisplayName)
	if err != nil {
		// TODO, should here still send out msg to server?
		glog.Errorf("Error parsing nodes: %s. Will return.", err)
		return
	}
	containerEntityDtos, err := ParseTask(mesosProbe)
	if err != nil {
		// TODO, should here still send out msg to server? Or set errorDTO?
		glog.Errorf("Error parsing pods: %s. Will return.", err)
		return
	}
	/*	appEntityDtos, err := kubeProbe.ParseApplication(api.NamespaceAll)
		if err != nil {
			glog.Errorf("Error parsing applications: %s. Will return.", err)
			return
		}

		serviceEntityDtos, err := kubeProbe.ParseService(api.NamespaceAll, labels.Everything())
		if err != nil {
			// TODO, should here still send out msg to server? Or set errorDTO?
			glog.Errorf("Error parsing services: %s. Will return.", err)
			return
		}
	*/
	entityDtos := nodeEntityDtos
	entityDtos = append(entityDtos, containerEntityDtos...)
	//	entityDtos = append(entityDtos, appEntityDtos...)
	//	entityDtos = append(entityDtos, serviceEntityDtos...)
	discoveryResponse := &comm.DiscoveryResponse{
		EntityDTO: entityDtos,
	}

	// 3. Build Client message
	clientMsg := comm.NewClientMessageBuilder(messageID).SetDiscoveryResponse(discoveryResponse).Create()
	fmt.Println("----------->about to call sdk send %+v \n", clientMsg)
	handler.wsComm.SendClientMessage(clientMsg)
}

func (handler *MesosServerMessageHandler) ActionBuilder(actionItem *sdk.ActionItemDTO, slavemap map[string]string) (*action.MesosClient, error) {
	actionpath := ""
	if actionItem.GetActionType() == sdk.ActionItemDTO_MOVE {
		glog.V(3).Infof("moving pod now")
		actionpath = "MigrateTasks"
		// create map
		if actionItem.GetTargetSE().GetEntityType() == sdk.EntityDTO_CONTAINER {
			newSEType := actionItem.GetNewSE().GetEntityType()
			if newSEType == sdk.EntityDTO_VIRTUAL_MACHINE || newSEType == sdk.EntityDTO_PHYSICAL_MACHINE {
				targetNode := actionItem.GetNewSE()
				var machineIPs []string
				switch newSEType {
				case sdk.EntityDTO_VIRTUAL_MACHINE:
					vmData := targetNode.GetVirtualMachineData()
					if vmData == nil {
						return nil, fmt.Errorf("Missing VM data")
					}
					machineIPs = vmData.GetIpAddress()
					break
				}
				targetContainer := actionItem.GetTargetSE()
				containerId := targetContainer.GetId()
				slaveId := ""
				//	slaveId, err := getNodeIdFromIP(machineIPs)
				for j := range machineIPs {
					if _, ok := slavemap[machineIPs[j]]; ok {
						slaveId = slavemap[machineIPs[j]]
						break
					}
				}
				return &action.MesosClient{
					MesosMasterIP:   handler.meta.MesosActionIP,
					MesosMasterPort: handler.meta.MesosActionPort,
					Action:          actionpath,
					DestinationId:   slaveId,
					TaskId:          containerId,
				}, nil
			}
		} else {
			// TODO
			return nil, fmt.Errorf("Missing data for move")
		}
		return nil, fmt.Errorf("Missing data for move")
	} // else if provision TODO
	return nil, fmt.Errorf("Missing data for move")
}

func createSlaveIdIpMap(resp *http.Response) (map[string]string, error) {
	fmt.Println("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	glog.V(3).Infof(" from glog response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}

	fmt.Println(" response was: %s", string(content))

	glog.V(4).Infof("response content is %s", string(content))
	byteContent := []byte(content)
	var jsonMesosMaster = new(util.MesosAPIResponse)
	err = json.Unmarshal(byteContent, &jsonMesosMaster)
	res := jsonMesosMaster.Slaves[0].Resources
	fmt.Printf("the MesosAPIResponse disk %f , mem %f , cpus %f  \n", res.Disk, res.Mem, res.CPUs)
	if err != nil {
		fmt.Printf("error in json unmarshal : %s", err)
	}
	SlaveIpIdMap := make(map[string]string)
	slaves := jsonMesosMaster.Slaves
	for i := range slaves {
		s := slaves[i]
		slaveIP := getSlaveIP(slaves[i])
		SlaveIpIdMap[slaveIP] = s.Id
	}
	return SlaveIpIdMap, nil
}

// Receives an action request from server and call ActionExecutor to execute action.
func (handler *MesosServerMessageHandler) HandleAction(serverMsg *comm.MediationServerMessage) {
	fmt.Println("--------> handleAction called")
	//	messageID := serverMsg.GetMessageID()
	actionRequest := serverMsg.GetActionRequest()
	actionItemDTO := actionRequest.GetActionItemDTO()
	glog.V(3).Infof("The received ActionItemDTO is %v", actionItemDTO)

	fullUrl := "http://" + handler.meta.MesosActionIP + ":5050" + "/state"
	fmt.Println("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)
	fmt.Println(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	respMap, err := createSlaveIdIpMap(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	fmt.Println("Get Succeed: %v", respMap)
	defer resp.Body.Close()

	simulator, err := handler.ActionBuilder(actionItemDTO, respMap)
	if err != nil {
		fmt.Printf("error %s \n", err)
	}
	_, err = action.RequestMesosAction(simulator)
	if err != nil {
		fmt.Printf("error %s \n", err)
	}

	/*
		err := actionExecutor.ExcuteAction(actionItemDTO, messageID)
		if err != nil {
			glog.Errorf("Error execute action: %s", err)
		}
	*/
}

func (handler *MesosServerMessageHandler) NewMesosProbe(previousUseMap map[string]*util.CalculatedUse) (*util.MesosAPIResponse, error) {
	fullUrl := "http://" + "10.10.174.91" + ":5050" + "/state"
	fmt.Println("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)

	//req.SetBasicAuth(vmtApi.extConfig["Username"], vmtApi.extConfig["Password"])
	fmt.Println(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return nil, err
	}
	respContent, err := parseAPIStateResponse(resp)
	if respContent.SlaveIdIpMap == nil {
		respContent.SlaveIdIpMap = make(map[string]string)
	}
	// UPDATE RESOURCE UNITS AFTER HTTP REQUEST
	for idx := range respContent.Slaves {
		s := respContent.Slaves[idx]
		s.Resources.Mem = s.Resources.Mem * float64(1024)
		s.UsedResources.Mem = s.UsedResources.Mem * float64(1024)
		s.OfferedResources.Mem = s.OfferedResources.Mem * float64(1024)
	}
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return nil, err
	}
	fmt.Println("Get Succeed: %v", respContent)
	defer resp.Body.Close()
	//	resp.Body.Close()

	fullUrl = "http://" + "10.10.174.91" + ":5050" + "/tasks"
	fmt.Println("The tasks full Url is ", fullUrl)
	req, err = http.NewRequest("GET", fullUrl, nil)

	//req.SetBasicAuth(vmtApi.extConfig["Username"], vmtApi.extConfig["Password"])
	fmt.Println(req)
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		glog.Errorf("Error getting tasks response: %s", err)
		return nil, err
	}
	/*
		configFile, err := os.Open("task.json")
		if err != nil {
			fmt.Println("opening config file", err.Error())
		}
		var jsonTasks = new(util.MasterTasks)

		jsonParser := json.NewDecoder(configFile)
		if err = jsonParser.Decode(jsonTasks); err != nil {
			fmt.Println("parsing config file", err.Error())
		}
		taskContent := jsonTasks
	*/
	taskContent, err := parseAPITasksResponse(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return nil, err
	}
	for j := range taskContent.Tasks {
		t := taskContent.Tasks[j]
		t.Resources.Mem = t.Resources.Mem * float64(1024)
	}
	respContent.TaskMasterAPI = *taskContent

	defer resp.Body.Close()

	// STATS
	var mapTaskRes map[string]util.Resources
	mapTaskRes = make(map[string]util.Resources)
	var mapSlaveUse map[string]*util.CalculatedUse
	mapSlaveUse = make(map[string]*util.CalculatedUse)
	arrM := respContent.TaskMasterAPI.Tasks
	for j := range arrM {
		if arrM[j].State != "TASK_RUNNING" {
			continue
		} else {
			for i := range respContent.Slaves {
				fmt.Println("=============> next slave")
				s := respContent.Slaves[i]
				fullUrl := "http://" + getSlaveIP(s) + ":5051" + "/monitor/statistics.json"
				req, err := http.NewRequest("GET", fullUrl, nil)
				req.Close = true
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println("Error getting response: %s", err)
					return nil, err
				}
				defer resp.Body.Close()
				stringResp, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("error %s", err)
				}
				//	fmt.Println("response is %+v", string(stringResp))
				byteContent := []byte(stringResp)
				var usedRes = new([]util.Executor)
				err = json.Unmarshal(byteContent, &usedRes)
				if err != nil {
					fmt.Printf("JSON erroriiiiiiii %s", err)
				}
				var res *util.Resources
				var arrOfExec []util.Executor
				arrOfExec = *usedRes
				for i := range arrOfExec {
					e := arrOfExec[i]
					if e.Id == arrM[j].Id {
						res = &util.Resources{
							Disk: float64(0),
							Mem:  e.Statistics.MemLimitBytes / float64(1024.00),
							CPUs: e.Statistics.CPUsLimit,
						}
						taskId := arrM[j].Id
						mapTaskRes[taskId] = *res
						break
					}
				}
				// SLAVE MONITOR
				if _, ok := mapSlaveUse[s.Id]; !ok {
					i := 0 // TODO sunday
					executor := arrOfExec[i]
					//if first time ??
					var prevSecs float64
					fmt.Println("          CALCULATION START :::")
					fmt.Printf("executor.Statistics.CPUsystemTimeSecs : %f \n", executor.Statistics.CPUsystemTimeSecs)
					fmt.Printf("executor.Statistics.CPUuserTimeSecs :%f \n", executor.Statistics.CPUuserTimeSecs)

					curSecs := executor.Statistics.CPUsystemTimeSecs + executor.Statistics.CPUuserTimeSecs
					if handler.lastDiscoveryTime == nil {
						fmt.Println("last time from handler is nil")
					}
					if previousUseMap == nil { //CPUsumSystemUserSecs == float64(0) {
						fmt.Println(" map was nil !!")
						prevSecs = curSecs

					} else {
						if _, ok := previousUseMap[s.Id]; !ok {
							fmt.Println("****** slave not found")
							// TODO NEW SLAVES
							continue
						}
						prevSecs = previousUseMap[s.Id].CPUsumSystemUserSecs
						fmt.Printf("previous system + user : %f and time %+v\n", prevSecs, respContent.TimeSinceLastDisc)
					}
					diffSecs := curSecs - prevSecs
					fmt.Println(" t1 - t0 : %f \n", diffSecs)
					var lastTime time.Time
					if handler.lastDiscoveryTime == nil {
						lastTime = time.Now()
					} else {
						lastTime = *handler.lastDiscoveryTime
					}
					diffTime := time.Since(lastTime)
					fmt.Printf(" last time on record : %+v \n", lastTime)
					diffT := diffTime.Seconds()
					fmt.Printf("time since last discovery in sec : %f \n", diffT)
					usedCPUfraction := diffSecs / diffT
					// ratio * cores * 1000kHz
					fmt.Printf("-------------> utilization CPU %f", usedCPUfraction)

					usedCPU := usedCPUfraction * s.Resources.CPUs * float64(1000)
					mapSlaveUse[s.Id] = &util.CalculatedUse{
						CPUs:                 usedCPU,
						CPUsumSystemUserSecs: curSecs,
					}
					fmt.Printf("-------------> used CPU %f", usedCPU)
					s.Calculated.CPUs = usedCPU
					// UPDATE stats
					timeNow := time.Now()
					handler.lastDiscoveryTime = &timeNow
				}
			}
		}
	}
	// map task to resources
	respContent.MapTaskResources = mapTaskRes
	handler.slaveUseMap = mapSlaveUse
	respContent.SlaveUseMap = mapSlaveUse
	return respContent, nil
}

func ParseNode(m *util.MesosAPIResponse, slaveUseMap map[string]*util.CalculatedUse) ([]*sdk.EntityDTO, error) {
	fmt.Printf("in ParseNode")
	result := []*sdk.EntityDTO{}
	for i := range m.Slaves {
		s := m.Slaves[i]
		// build sold commodities
		slaveProbe := &probe.NodeProbe{
			MasterState: m,
		}
		commoditiesSold, err := slaveProbe.CreateCommoditySold(&s, slaveUseMap)
		if err != nil {
			fmt.Printf("error is : %s", err)
			return result, err
		}
		slaveIP := getSlaveIP(s)
		m.SlaveIdIpMap[s.Id] = slaveIP
		entityDTO := buildVMEntityDTO(slaveIP, s.Id, s.Name, commoditiesSold)
		result = append(result, entityDTO)
	}
	fmt.Printf(" entity DTOs : %d", len(result))
	return result, nil
}

func ParseTask(m *util.MesosAPIResponse) ([]*sdk.EntityDTO, error) {
	result := []*sdk.EntityDTO{}
	taskList := m.TaskMasterAPI.Tasks
	for i := range taskList {
		taskProbe := &probe.TaskProbe{
			Task: &taskList[i],
		}
		if taskProbe.Task.State == "TASK_KILLED" {
			continue
		}
		//ipAddress := slaveIdIpMap[taskProbe.Task.SlaveId]
		//usedResources := taskProbe.GetUsedResourcesForTask(ipAddress)
		taskResource, err := taskProbe.GetTaskResourceStat(m.MapTaskResources, taskProbe.Task)
		if err != nil {
			fmt.Printf("error is : %s", err)
		}
		commoditiesSoldContainer := taskProbe.GetCommoditiesSoldByContainer(taskProbe.Task, taskResource)
		commoditiesBoughtContainer := taskProbe.GetCommoditiesBoughtByContainer(taskProbe.Task, taskResource)

		entityDTO, _ := buildTaskContainerEntityDTO(m.SlaveIdIpMap, taskProbe.Task, commoditiesSoldContainer, commoditiesBoughtContainer)

		result = append(result, entityDTO)

		commoditiesSoldApp := taskProbe.GetCommoditiesSoldByApp(taskProbe.Task, taskResource)
		commoditiesBoughtApp := taskProbe.GetCommoditiesBoughtByApp(taskProbe.Task, taskResource)

		entityDTO = buildTaskAppEntityDTO(m.SlaveIdIpMap, taskProbe.Task, commoditiesSoldApp, commoditiesBoughtApp)

		result = append(result, entityDTO)
	}
	fmt.Printf(" entity DTOs : %d", len(result))
	return result, nil
}

func parseAPITasksResponse(resp *http.Response) (*util.MasterTasks, error) {
	fmt.Println("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	glog.V(3).Infof(" from glog response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}

	fmt.Println(" response was: %s", string(content))

	glog.V(4).Infof("response content is %s", string(content))

	byteContent := []byte(content)
	var jsonTasks = new(util.MasterTasks)
	err = json.Unmarshal(byteContent, &jsonTasks)

	//	fmt.Printf("the MesosAPIResponse disk %f , mem %f , cpus %f  \n", res.Disk, res.Mem, res.CPUs)
	if err != nil {
		fmt.Printf("error in json unmarshal : %s", err)
	}
	return jsonTasks, nil
}

func parseAPIStateResponse(resp *http.Response) (*util.MesosAPIResponse, error) {
	fmt.Println("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	glog.V(3).Infof(" from glog response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}

	fmt.Println(" response was: %s", string(content))

	glog.V(4).Infof("response content is %s", string(content))
	byteContent := []byte(content)
	var jsonMesosMaster = new(util.MesosAPIResponse)
	err = json.Unmarshal(byteContent, &jsonMesosMaster)
	res := jsonMesosMaster.Slaves[0].Resources
	fmt.Printf("the MesosAPIResponse disk %f , mem %f , cpus %f  \n", res.Disk, res.Mem, res.CPUs)
	if err != nil {
		fmt.Printf("error in json unmarshal : %s", err)
	}
	return jsonMesosMaster, nil
}

func buildTaskAppEntityDTO(slaveIdIp map[string]string, task *util.Task, commoditiesSold []*sdk.CommodityDTO, commoditiesBoughtMap map[*sdk.ProviderDTO][]*sdk.CommodityDTO) *sdk.EntityDTO {
	appEntityType := sdk.EntityDTO_APPLICATION
	id := task.Name + "::" + "APP:" + task.Id
	dispName := "APP:" + task.Name
	entityDTOBuilder := sdk.NewEntityDTOBuilder(appEntityType, id)
	entityDTOBuilder = entityDTOBuilder.DisplayName(dispName)

	entityDTOBuilder.SellsCommodities(commoditiesSold)

	for provider, commodities := range commoditiesBoughtMap {
		entityDTOBuilder.SetProvider(provider)
		entityDTOBuilder.BuysCommodities(commodities)
	}

	entityDto := entityDTOBuilder.Create()

	appType := task.Name

	ipAddress := slaveIdIp[task.SlaveId] //this.getIPAddress(host, nodeName)

	appData := &sdk.EntityDTO_ApplicationData{
		Type:      &appType,
		IpAddress: &ipAddress,
	}
	entityDto.ApplicationData = appData
	return entityDto

}

// Build entityDTO that contains all the necessary info of a pod.
func buildTaskContainerEntityDTO(slaveIdIpMap map[string]string, task *util.Task, commoditiesSold, commoditiesBought []*sdk.CommodityDTO) (*sdk.EntityDTO, error) {
	taskName := task.Name
	id := task.Id
	dispName := task.Name

	entityDTOBuilder := sdk.NewEntityDTOBuilder(sdk.EntityDTO_CONTAINER, id)
	entityDTOBuilder.DisplayName(dispName)

	slaveId := task.SlaveId
	if slaveId == "" {
		return nil, fmt.Errorf("Cannot find the hosting slave ID for task %s", taskName)
	}
	glog.V(4).Infof("Pod %s is hosted on %s", dispName, slaveId)

	entityDTOBuilder.SellsCommodities(commoditiesSold)
	//	providerUid := nodeUidTranslationMap[slaveId]
	entityDTOBuilder = entityDTOBuilder.SetProviderWithTypeAndID(sdk.EntityDTO_VIRTUAL_MACHINE, slaveId)
	entityDTOBuilder.BuysCommodities(commoditiesBought)
	ipAddress := slaveIdIpMap[task.SlaveId]
	entityDTOBuilder = entityDTOBuilder.SetProperty("ipAddress", ipAddress)
	glog.V(3).Infof("Pod %s will be stitched to VM with IP %s", dispName, ipAddress)

	entityDto := entityDTOBuilder.Create()
	return entityDto, nil
}

func getSlaveIP(s util.Slave) string {
	//"slave(1)@10.10.174.92:5051"
	var ipportArray []string
	slaveIP := ""
	ipLong := s.Pid
	arr := strings.Split(ipLong, "@")
	if len(arr) > 1 {
		ipport := arr[1]
		ipportArray = strings.Split(ipport, ":")
	}
	if len(ipportArray) > 0 {
		slaveIP = ipportArray[0]
	}
	return slaveIP
}

func buildVMEntityDTO(slaveIP, nodeID, displayName string, commoditiesSold []*sdk.CommodityDTO) *sdk.EntityDTO {
	entityDTOBuilder := sdk.NewEntityDTOBuilder(sdk.EntityDTO_VIRTUAL_MACHINE, nodeID)
	entityDTOBuilder.DisplayName(displayName)
	entityDTOBuilder.SellsCommodities(commoditiesSold)
	// TODO stitch
	ipAddress := slaveIP //nodeProbe.getIPForStitching(displayName)
	entityDTOBuilder = entityDTOBuilder.SetProperty("IP", ipAddress)
	glog.V(4).Infof("Parse node: The ip of vm to be reconcile with is %s", ipAddress)
	metaData := generateReconcilationMetaData()

	entityDTOBuilder = entityDTOBuilder.ReplacedBy(metaData)
	entityDto := entityDTOBuilder.Create()
	return entityDto
}

func generateReconcilationMetaData() *sdk.EntityDTO_ReplacementEntityMetaData {
	replacementEntityMetaDataBuilder := sdk.NewReplacementEntityMetaDataBuilder()
	replacementEntityMetaDataBuilder.Matching("IP")
	replacementEntityMetaDataBuilder.PatchSelling(sdk.CommodityDTO_CPU_ALLOCATION)
	replacementEntityMetaDataBuilder.PatchSelling(sdk.CommodityDTO_MEM_ALLOCATION)
	metaData := replacementEntityMetaDataBuilder.Build()
	return metaData
}
