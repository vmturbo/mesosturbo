package probe

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
)

type TaskResourceStat struct {
	cpuAllocationCapacity  float64
	cpuAllocationUsed      float64
	memAllocationCapacity  float64
	memAllocationUsed      float64
	diskAllocationCapacity float64
	diskAllocationUsed     float64
}

// models the probe for a mesos master state , containing metadata about
// all of the slaves
type TaskProbe struct {
	Task *util.Task
}

// Get current stat of node resources, such as capacity and used values.
func (probe *TaskProbe) GetTaskResourceStat(mapT map[string]util.Statistics, task *util.Task, taskUseMap map[string]*util.CalculatedUse) (*TaskResourceStat, error) {
	// The metrics we care about now are Cpu and Mem.
	// TODO io metrics
	//requests := task.Resources.Limits
	glog.V(3).Infof("---------------------> task.Resources.CPUs is %f\n", task.Resources.CPUs)
	cpuCapacity := task.Resources.CPUs * float64(2000.00)
	memCapacity := mapT[task.Id].MemLimitBytes / float64(1024.00)
	diskCapacity := mapT[task.Id].DiskLimitBytes / float64(1024.00*1024.00)

	glog.V(4).Infof("Discovered task is " + task.Id)
	glog.V(4).Infof("Container capacity is %f \n", cpuCapacity)
	glog.V(4).Infof("Container Mem capacity is %f \n", memCapacity)

	if taskUseMap != nil {
		glog.V(4).Infof("Task use map is nil.\n")
	} else {
		glog.V(4).Infof("task map is nil.\n")
	}
	memUsed := mapT[task.Id].MemRSSBytes / float64(1024.00)
	cpuUsed := taskUseMap[task.Id].CPUs
	// assuming disk unit is KB
	diskUsed := mapT[task.Id].DiskUsedBytes / float64(1024.00*1024.00)

	return &TaskResourceStat{
		cpuAllocationCapacity:  cpuCapacity,
		cpuAllocationUsed:      cpuUsed,
		memAllocationCapacity:  memCapacity,
		memAllocationUsed:      memUsed,
		diskAllocationCapacity: diskCapacity,
		diskAllocationUsed:     diskUsed,
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
	diskAllocationComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key(task.Id).
		Capacity(float64(taskResourceStat.diskAllocationCapacity)).
		Used(taskResourceStat.diskAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, diskAllocationComm)
	return commoditiesSold
}

// Build commodityDTOs for commodity sold by the pod
func (taskProbe *TaskProbe) GetCommoditiesBoughtByContainer(task *util.Task, taskResourceStat *TaskResourceStat) []*sdk.CommodityDTO {
	var commoditiesBought []*sdk.CommodityDTO
	cpuAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, cpuAllocationCommBought)
	memAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, memAllocationCommBought)
	diskAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.diskAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, diskAllocationCommBought)
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
	diskAllocationCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.diskAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, diskAllocationCommBought)

	commoditiesBoughtMap[containerProvider] = commoditiesBought

	slaveProvider := sdk.CreateProvider(sdk.EntityDTO_VIRTUAL_MACHINE, task.SlaveId)
	var commoditiesBoughtFromSlave []*sdk.CommodityDTO

	vCpuCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VCPU).
		//		Key(task.SlaveId).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, vCpuCommBought)

	vMemCommBought := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VMEM).
		//		Key(task.SlaveId).
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
