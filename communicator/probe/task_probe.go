package probe

import (
	"fmt"

	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
)

type TaskResourceStat struct {
	cpuAllocationCapacity float64
	cpuAllocationUsed     float64
	memAllocationCapacity float64
	memAllocationUsed     float64
}

// models the probe for a mesos master state , containing metadata about
// all of the slaves
type TaskProbe struct {
	Task *util.Task
}

/*
func parseUsedResources(taskId string, resp *http.Response) *util.Resources{
	content, err := ioutil.ReadAll(resp.Body)
	byteContent := []byte(content)

	var usedRes = new([]util.Executor)
//	err := json.NewDecoder(resp.Body).Decode(&usedRes)
	err = json.Unmarshal(byteContent, &usedRes)
	if (err != nil){
		fmt.Printf("JSON erroriiiiiiii %s", err)
	}
	var res *util.Resources
	var arrOfExec []util.Executor
	arrOfExec = *usedRes
	for i := range arrOfExec{
		e := arrOfExec[i]
		if(e.Id == taskId){
			res = &util.Resources{
				Disk: float64(0),
				Mem: e.Statistics.MemLimitBytes/float64(1024.00),
				CPUs: e.Statistics.CPUsLimit,
			}
			//e.Statistics.MemLimitBytes
			//e.Statistics.MemRSSBytes
			break
		}
	}
	return res
}

func (probe *TaskProbe) GetUsedResourcesForTask(ipAddress string) *util.Resources {
	fullUrl := "http://" + ipAddress + ":5051" + "/monitor/statistics.json"
	req, err := http.NewRequest("GET", fullUrl, nil)
	req.Close = true
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
               fmt.Println("Error getting response: %s", err)
               return nil
        }
	defer resp.Body.Close()
	stringResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error %s", err)
	}
	fmt.Println("response is %+v",string(stringResp))
	used := parseUsedResources(probe.Task.Id, resp)
	return used
}
*/

// Get current stat of node resources, such as capacity and used values.
func (probe *TaskProbe) GetTaskResourceStat(mapT map[string]util.Statistics, task *util.Task, taskUseMap map[string]*util.CalculatedUse) (*TaskResourceStat, error) {
	// TODO! Here we assume when user defines a pod, resource requirements are also specified.
	// The metrics we care about now are Cpu and Mem.
	//requests := task.Resources.Limits
	memCapacity := mapT[task.Id].MemLimitBytes / float64(1024.00)
	cpuCapacity := task.Resources.CPUs * float64(1000.00)

	fmt.Println("Discovered task is " + task.Id)
	fmt.Printf("Container capacity is %f \n", cpuCapacity)
	fmt.Printf("Container Mem capacity is %f \n", memCapacity)

	// this flag is defined at package level, in probe.go
	//	if localTestingFlag {
	//		cpuUsed = float64(10000)
	//	}
	if taskUseMap != nil {
		//	fmt.Printf(" task used map is %+v and value is %+v \n", taskUseMap, taskUseMap[task.Id])
	} else {
		fmt.Println("task map is nil")
	}
	memUsed := mapT[task.Id].MemRSSBytes / float64(1024.00)
	cpuUsed := taskUseMap[task.Id].CPUs

	return &TaskResourceStat{
		cpuAllocationCapacity: cpuCapacity,
		cpuAllocationUsed:     cpuUsed,
		memAllocationCapacity: memCapacity,
		memAllocationUsed:     memUsed,
	}, nil
}

// Build commodityDTOs for commodity sold by the pod
func (TaskProbe *TaskProbe) GetCommoditiesSoldByContainer(task *util.Task, taskResourceStat *TaskResourceStat) []*sdk.CommodityDTO {
	var commoditiesSold []*sdk.CommodityDTO
	memAllocationComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key(task.Id).
		Capacity(float64(taskResourceStat.memAllocationCapacity)).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, memAllocationComm)
	cpuAllocationComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key(task.Id).
		Capacity(float64(taskResourceStat.cpuAllocationCapacity)).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, cpuAllocationComm)
	return commoditiesSold
}

// Build commodityDTOs for commodity sold by the pod
func (taskProbe *TaskProbe) GetCommoditiesBoughtByContainer(task *util.Task, taskResourceStat *TaskResourceStat) []*sdk.CommodityDTO {
	var commoditiesBought []*sdk.CommodityDTO
	cpuAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key("Container").
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, cpuAllocationCommBought)
	memAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key("Container").
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, memAllocationCommBought)
	// TODO vmpm  access commodity
	return commoditiesBought
}

// Build commodityDTOs for commodity sold by the app
func (TaskProbe *TaskProbe) GetCommoditiesSoldByApp(task *util.Task, taskResourceStat *TaskResourceStat) []*sdk.CommodityDTO {
	appName := task.Name
	var commoditiesSold []*sdk.CommodityDTO
	transactionComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_TRANSACTION).
		Key(appName).
		Capacity(float64(0)).
		Used(float64(0)).
		Create()
	commoditiesSold = append(commoditiesSold, transactionComm)
	return commoditiesSold
}

// Build commodityDTOs for commodity bought by the app
func (taskProbe *TaskProbe) GetCommoditiesBoughtByApp(task *util.Task, taskResourceStat *TaskResourceStat) map[*sdk.ProviderDTO][]*sdk.CommodityDTO {
	commoditiesBoughtMap := make(map[*sdk.ProviderDTO][]*sdk.CommodityDTO)
	// TODO check about name
	containerName := task.Id
	containerProvider := sdk.CreateProvider(sdk.EntityDTO_CONTAINER, containerName)
	var commoditiesBought []*sdk.CommodityDTO
	cpuAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, cpuAllocationCommBought)
	memAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, memAllocationCommBought)

	commoditiesBoughtMap[containerProvider] = commoditiesBought

	slaveProvider := sdk.CreateProvider(sdk.EntityDTO_VIRTUAL_MACHINE, task.SlaveId)
	var commoditiesBoughtFromSlave []*sdk.CommodityDTO

	vCpuCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VCPU).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, vCpuCommBought)

	vMemCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VMEM).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, vMemCommBought)

	appCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_APPLICATION).
		Key(task.SlaveId).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, appCommBought)
	commoditiesBoughtMap[slaveProvider] = commoditiesBoughtFromSlave
	return commoditiesBoughtMap
}
