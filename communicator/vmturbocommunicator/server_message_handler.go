package vmturbocommunicator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	vmtmeta "github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/probe"
	"github.com/vmturbo/mesosturbo/communicator/util"
	vmtapi "github.com/vmturbo/mesosturbo/communicator/vmtapi"
	"github.com/vmturbo/mesosturbo/pkg/action"
	comm "github.com/vmturbo/vmturbo-go-sdk/communicator"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
)

// impletements sdk.ServerMessageHandler
type MesosServerMessageHandler struct {
	meta              *vmtmeta.VMTMeta
	wsComm            *comm.WebSocketCommunicator
	vmtComm           *VMTCommunicator
	lastDiscoveryTime *time.Time
	slaveUseMap       map[string]*util.CalculatedUse
	taskUseMap        map[string]*util.CalculatedUse
}

// Use the vmt restAPI to add a Mesos target.
func (handler *MesosServerMessageHandler) AddTarget() {
	glog.V(4).Infof("------> in AddTarget()")
	vmtUrl := handler.wsComm.VmtServerAddress

	extCongfix := make(map[string]string)
	extCongfix["Username"] = handler.meta.OpsManagerUsername
	extCongfix["Password"] = handler.meta.OpsManagerPassword
	vmturboApi := vmtapi.NewVmtApi(vmtUrl, extCongfix)

	// Add Mesos target.
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

	// Discover Mesos target.
	vmturboApi.DiscoverTarget(handler.meta.NameOrAddress)
}

// If server sends a validation request, validate the request.
// TODO Validate all the request. aka, no matter what usr/passwd is provided, always pass validation.
// The correct bahavior is to set ErrorDTO when validation fails.
func (handler *MesosServerMessageHandler) Validate(serverMsg *comm.MediationServerMessage) {
	//Always send Validated for now
	glog.V(3).Infof("Mesos validation request from Server")

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
	//Discover the Mesos topology
	glog.V(3).Infof("Discover topology request from server.")
	// 1. Get message ID
	messageID := serverMsg.GetMessageID()
	var stopCh chan struct{} = make(chan struct{})
	go util.Until(func() { handler.keepDiscoverAlive(messageID) }, time.Second*10, stopCh)
	defer close(stopCh)

	// 2. Build discoverResponse
	mesosProbe, err := handler.NewMesosProbe(handler.taskUseMap)
	if err != nil {
		glog.Errorf("Error getting state from master : %s", err)
		return
	}
	nodeEntityDtos, err := ParseNode(mesosProbe, handler.slaveUseMap)
	if err != nil {
		// TODO, should here still send out msg to server?
		glog.Errorf("Error parsing nodes: %s. Will return.", err)
		return
	}
	containerEntityDtos, err := ParseTask(mesosProbe, handler.taskUseMap)
	if err != nil {
		// TODO, should here still send out msg to server? Or set errorDTO?
		glog.Errorf("Error parsing pods: %s. Will return.", err)
		return
	}

	entityDtos := nodeEntityDtos
	entityDtos = append(entityDtos, containerEntityDtos...)
	//	entityDtos = append(entityDtos, serviceEntityDtos...)
	discoveryResponse := &comm.DiscoveryResponse{
		EntityDTO: entityDtos,
	}

	// 3. Build Client message
	clientMsg := comm.NewClientMessageBuilder(messageID).SetDiscoveryResponse(discoveryResponse).Create()
	curtime := time.Now()
	handler.lastDiscoveryTime = &curtime
	//	glog.V(3).Infof(" Discovery msg : %+v", clientMsg)
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
				glog.V(3).Infof(" destination IP is %s and task is  %s \n", slaveId, containerId)
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
			glog.Errorf("Unsupported entity destination type for moving")
			return nil, fmt.Errorf("Unsupported entity destination type for moving")
		}
		glog.Errorf("Unsupported entity source type for moving")
		return nil, fmt.Errorf("Unsupported entity source type for moving")
	} // else if provision TODO
	glog.Errorf("Unsupporter action type")
	return nil, fmt.Errorf("Unsupported action type")
}

/*
func CreateSlaveIpIdMap(resp *http.Response) (map[string]string, error) {
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
*/
// Receives an action request from server and call ActionExecutor to execute action.
func (handler *MesosServerMessageHandler) HandleAction(serverMsg *comm.MediationServerMessage) {
	//	messageID := serverMsg.GetMessageID()
	actionRequest := serverMsg.GetActionRequest()
	actionItemDTO := actionRequest.GetActionItemDTO()
	glog.V(3).Infof("The received ActionItemDTO is %v", actionItemDTO)

	fullUrl := "http://" + handler.meta.MesosActionIP + ":5050" + "/state"
	glog.V(4).Infof("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)
	glog.V(4).Infof("%+v", req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	respMap, err := util.CreateSlaveIpIdMap(resp)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
	}
	defer resp.Body.Close()

	simulator, err := handler.ActionBuilder(actionItemDTO, respMap)
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
	_, err = action.RequestMesosAction(simulator)
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
	// response
	handler.vmtComm.SendActionResponse(sdk.ActionResponseState_SUCCEEDED, int32(100), serverMsg.GetMessageID(), "Success")

	/*
		err := actionExecutor.ExcuteAction(actionItemDTO, messageID)
		if err != nil {
			glog.Errorf("Error execute action: %s", err)
		}
	*/
}

func (handler *MesosServerMessageHandler) NewMesosProbe(previousUseMap map[string]*util.CalculatedUse) (*util.MesosAPIResponse, error) {
	fullUrl := "http://" + handler.meta.MesosActionIP + ":5050" + "/state"
	glog.V(4).Infof("The full Url is ", fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)

	glog.V(4).Infof("%+v", req)
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
		glog.V(3).Infof("Number of slaves %d \n", len(respContent.Slaves))
		s := &respContent.Slaves[idx]
		s.Resources.Mem = s.Resources.Mem * float64(1024)
		s.UsedResources.Mem = s.UsedResources.Mem * float64(1024)
		s.OfferedResources.Mem = s.OfferedResources.Mem * float64(1024)
		glog.V(3).Infof("=======> SLAVE idk: %d name: %s, mem: %.2f, cpu: %.2f, disk: %.2f \n", idx, s.Name, s.Resources.Mem, s.Resources.CPUs, s.Resources.Disk)
	}
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return nil, err
	}
	glog.V(3).Infof("Get Succeed: %v\n", respContent)
	defer resp.Body.Close()

	if respContent.Frameworks == nil {
		glog.Errorf("Error getting Frameworks response: %s", err)
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

	//We pass the entire http response as the respContent object
	taskContent, err := parseAPITasksResponse(respContent)
	if err != nil {
		glog.Errorf("Error getting response: %s", err)
		return nil, err
	}
	glog.V(3).Infof("Number of tasks \n", len(taskContent.Tasks))

	for j := range taskContent.Tasks {
		t := taskContent.Tasks[j]
		// MEM UNITS KB
		t.Resources.Mem = t.Resources.Mem * float64(1024)
		//	fmt.Printf("----> tasks from mesos: # %d, name : %s, state: %s\n", j, t.Name, t.State)
		//	glog.V(3).Infof("=======> TASK name: %s, mem: %.2f, cpu: %.2f, disk: %.2f \n", t.Name, t.Resources.Mem, t.Resources.CPUs, t.Resources.Disk)

	}
	respContent.TaskMasterAPI = *taskContent
	glog.V(4).Infof("tasks response is %+v \n", resp.Body)
	defer resp.Body.Close()

	// STATS
	var mapTaskRes map[string]util.Statistics
	mapTaskRes = make(map[string]util.Statistics)
	var mapSlaveUse map[string]*util.CalculatedUse
	mapSlaveUse = make(map[string]*util.CalculatedUse)
	var mapTaskUse map[string]*util.CalculatedUse
	mapTaskUse = make(map[string]*util.CalculatedUse)
	for i := range respContent.Slaves {
		s := respContent.Slaves[i]
		fullUrl := "http://" + util.GetSlaveIP(s) + ":5051" + "/monitor/statistics.json"
		req, err := http.NewRequest("GET", fullUrl, nil)
		req.Close = true
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			glog.V(4).Infof("Error getting response: %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		stringResp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.V(4).Infof("error %s", err)
		}
		byteContent := []byte(stringResp)
		var usedRes = new([]util.Executor)
		err = json.Unmarshal(byteContent, &usedRes)
		if err != nil {
			glog.V(4).Infof("JSON error %s", err)
		}
		var arrOfExec []util.Executor
		arrOfExec = *usedRes
		mapSlaveUse[s.Id] = &util.CalculatedUse{
			CPUs: float64(0),
		}
		for j := range arrOfExec {
			executor := arrOfExec[j]
			// TODO check if this is taskId
			taskId := executor.Source
			mapTaskRes[taskId] = executor.Statistics
			// TASK MONITOR
			if _, ok := mapTaskUse[taskId]; !ok {
				var prevSecs float64
				// CPU use CALCULATION STARTS

				curSecs := executor.Statistics.CPUsystemTimeSecs + executor.Statistics.CPUuserTimeSecs
				if handler.lastDiscoveryTime == nil {
					glog.V(4).Infof("last time from handler is nil")
				}
				_, ok := previousUseMap[taskId]
				if previousUseMap == nil || !ok {
					glog.V(4).Infof(" map was nil !!")
					prevSecs = curSecs

				} else {
					prevSecs = previousUseMap[taskId].CPUsumSystemUserSecs
					glog.V(4).Infof("previous system + user : %f and time %+v\n", prevSecs, respContent.TimeSinceLastDisc)
				}
				diffSecs := curSecs - prevSecs
				if diffSecs < 0 {
					diffSecs = float64(0.0)
				}
				glog.V(4).Infof(" t1 - t0 : %f \n", diffSecs)
				var lastTime time.Time
				if handler.lastDiscoveryTime == nil {
					lastTime = time.Now()
				} else {
					lastTime = *handler.lastDiscoveryTime
				}
				diffTime := time.Since(lastTime)
				//	fmt.Printf(" last time on record : %+v \n", lastTime)
				diffT := diffTime.Seconds()
				//	fmt.Printf("time since last discovery in sec : %f \n", diffT)
				usedCPUfraction := diffSecs / diffT
				// ratio * cores * 1000kHz
				glog.V(4).Infof("-------------> Fraction of CPU utilization: %f \n", usedCPUfraction)

				// s.Resources is # of cores
				// usedCPU is in MHz
				usedCPU := usedCPUfraction * s.Resources.CPUs * float64(1000)
				mapTaskUse[taskId] = &util.CalculatedUse{
					CPUs:                 usedCPU,
					CPUsumSystemUserSecs: curSecs,
				}
				glog.V(4).Infof("------------> Capacity in CPUs, directly from Mesos %f \n", s.Resources.CPUs)
				glog.V(4).Infof("------------->Used CPU in MHz : %f \n", usedCPU)

				// Sum the used CPU in MHz for each slave
				mapSlaveUse[s.Id].CPUs = usedCPU + mapSlaveUse[s.Id].CPUs
			}
		} // task loop
	} // slave loop

	// map task to resources
	handler.taskUseMap = mapTaskUse
	handler.slaveUseMap = mapSlaveUse
	respContent.MapTaskStatistics = mapTaskRes
	respContent.SlaveUseMap = mapSlaveUse
	respContent.Cluster.MasterIP = handler.meta.MesosActionIP
	respContent.Cluster.ClusterName = respContent.ClusterName
	return respContent, nil
}

func ParseNode(m *util.MesosAPIResponse, slaveUseMap map[string]*util.CalculatedUse) ([]*sdk.EntityDTO, error) {
	glog.V(4).Infof("in ParseNode\n")
	result := []*sdk.EntityDTO{}
	for i := range m.Slaves {
		s := m.Slaves[i]
		// build sold commodities
		slaveProbe := &probe.NodeProbe{
			MasterState: m,
			Cluster:     &m.Cluster,
		}
		commoditiesSold, err := slaveProbe.CreateCommoditySold(&s, slaveUseMap)
		if err != nil {
			glog.Errorf("error is : %s\n", err)
			return result, err
		}
		slaveIP := util.GetSlaveIP(s)
		m.SlaveIdIpMap[s.Id] = slaveIP
		entityDTO := buildVMEntityDTO(slaveIP, s.Id, s.Name+"foo", commoditiesSold)
		result = append(result, entityDTO)
	}
	glog.V(4).Infof(" entity DTOs : %d\n", len(result))
	return result, nil
}

func ParseTask(m *util.MesosAPIResponse, taskUseMap map[string]*util.CalculatedUse) ([]*sdk.EntityDTO, error) {
	result := []*sdk.EntityDTO{}
	taskList := m.TaskMasterAPI.Tasks
	for i := range taskList {
		if _, ok := taskUseMap[taskList[i].Id]; !ok {
			continue
		}
		taskProbe := &probe.TaskProbe{
			Task:    &taskList[i],
			Cluster: &m.Cluster,
		}
		if taskProbe.Task.State != "TASK_RUNNING" {
			glog.V(4).Infof("=====> not running task is %s and state %s\n", taskProbe.Task.Name, taskProbe.Task.State)
			continue
		}
		glog.V(4).Infof("=====> task is %s and state %s\n", taskProbe.Task.Name, taskProbe.Task.State)

		//ipAddress := slaveIdIpMap[taskProbe.Task.SlaveId]
		//usedResources := taskProbe.GetUsedResourcesForTask(ipAddress)
		taskResource, err := taskProbe.GetTaskResourceStat(m.MapTaskStatistics, taskProbe.Task, taskUseMap)
		if err != nil {
			glog.Errorf("error is : %s", err)
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
	glog.V(4).Infof("Task entity DTOs : %d", len(result))
	return result, nil
}

func parseAPITasksResponse(resp *util.MesosAPIResponse) (*util.MasterTasks, error) {
	glog.V(4).Infof("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	glog.V(3).Infof(" Number of frameworks is %d\n", len(resp.Frameworks))

	allTasks := make([]util.Task, 0)
	for i := range resp.Frameworks {
		if resp.Frameworks[i].Tasks != nil {
			ftasks := resp.Frameworks[i].Tasks
			for j := range ftasks {
				allTasks = append(allTasks, ftasks[j])
			}
			glog.V(3).Infof(" Number of tasks is %d\n", len(resp.Frameworks[i].Tasks))
		}
	}
	tasksObj := &util.MasterTasks{
		Tasks: allTasks,
	}
	return tasksObj, nil
}

func parseAPIStateResponse(resp *http.Response) (*util.MesosAPIResponse, error) {
	glog.V(4).Infof("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	glog.V(3).Infof(" from glog response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
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
	return jsonMesosMaster, nil
}

func buildTaskAppEntityDTO(slaveIdIp map[string]string, task *util.Task, commoditiesSold []*sdk.CommodityDTO, commoditiesBoughtMap map[*sdk.ProviderDTO][]*sdk.CommodityDTO) *sdk.EntityDTO {
	appEntityType := sdk.EntityDTO_APPLICATION
	id := task.Name + "::" + "APP:" + task.Id
	dispName := "APP:" + task.Name + "foofoo"
	entityDTOBuilder := sdk.NewEntityDTOBuilder(appEntityType, id+"foo")
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
	replacementEntityMetaDataBuilder.PatchSelling(sdk.CommodityDTO_VCPU)
	replacementEntityMetaDataBuilder.PatchSelling(sdk.CommodityDTO_VMEM)
	replacementEntityMetaDataBuilder.PatchSelling(sdk.CommodityDTO_APPLICATION)
	metaData := replacementEntityMetaDataBuilder.Build()
	return metaData
}
