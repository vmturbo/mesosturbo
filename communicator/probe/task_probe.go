package probe

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
	"strconv"
	"strings"
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
	Task           *util.Task
	Cluster        *util.ClusterInfo
	Constraints    [][]string
	ConstraintsMap map[string][][]string
	PortsUsed      map[string]util.PortUtil
}

func (probe *TaskProbe) getPortsBought() {
	// Task.Resources.Ports
	var portsTaskUses map[string]util.PortUtil
	portsTaskUses = make(map[string]util.PortUtil)
	usedStr := probe.Task.Resources.Ports
	if usedStr != "" {
		glog.V(3).Infof("=========-------> used ports at container is %+v\n", usedStr)
		portsStr := usedStr[1 : len(usedStr)-1]
		glog.V(3).Infof("=========-------> used ports is %+v\n", portsStr)
		portRanges := strings.Split(portsStr, ",")
		for _, prange := range portRanges {
			glog.V(3).Infof("=========-------> prange is %+v\n", prange)
			ports := strings.Split(prange, "-")
			glog.V(3).Infof("=========-------> port is %+v\n", ports[0])
			portStart, err := strconv.Atoi(strings.Trim(ports[0], " "))
			if err != nil {
				glog.V(3).Infof(" Error: %+v", err)
			}
			if strings.Trim(ports[0], " ") == strings.Trim(ports[1], " ") {
				// ports used by Task
				portsTaskUses[strings.Trim(ports[0], " ")] = util.PortUtil{
					Number:   float64(portStart),
					Capacity: float64(1.0),
					Used:     float64(1.0),
				}
			} else {
				//range from port start to end
				for _, p := range ports {
					port, err := strconv.Atoi(p)
					if err != nil {
						glog.V(3).Infof("Error getting used port. %+v\n", err)
					}
					portsTaskUses[strings.Trim(p, " ")] = util.PortUtil{
						Number:   float64(port),
						Capacity: float64(1.0),
						Used:     float64(1.0),
					}
				}
			}
		}
	}
	probe.PortsUsed = portsTaskUses
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
	memAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key(task.Id).
		Capacity(float64(taskResourceStat.memAllocationCapacity)).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, memAllocationComm)
	cpuAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key(task.Id).
		Capacity(float64(taskResourceStat.cpuAllocationCapacity)).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, cpuAllocationComm)
	diskAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
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
	cpuAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, cpuAllocationCommBought)
	memAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, memAllocationCommBought)
	diskAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key("Mesos").
		Used(taskResourceStat.diskAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, diskAllocationCommBought)
	clusterCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CLUSTER).
		Key(taskProbe.Cluster.ClusterName).
		Create()
	glog.V(3).Infof("-------> cluster Commodity bought by Container is %s \n", taskProbe.Cluster.ClusterName)
	commoditiesBought = append(commoditiesBought, clusterCommBought)
	glog.V(3).Infof("========> size %d and labels  %+v", len(taskProbe.Task.Labels), taskProbe.Task.Labels)
	// this is only for constraint type CLUSTER
	/*	for _, c := range taskProbe.Constraints {
			if c[1] == "CLUSTER" {
				key := c[0]
				val := c[2]
				glog.V(3).Infof("========> key %s and value  ", key)
				vmpmAccessCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VMPM_ACCESS).
					Key(key + ":" + val).
					Create()
				commoditiesBought = append(commoditiesBought, vmpmAccessCommBought)
			}
		}
	*/ // TODO other constraint operator types
	taskProbe.getPortsBought()
	glog.V(3).Infof("\n\n\n")
	for k, v := range taskProbe.PortsUsed {
		glog.V(3).Infof(" -------->>>> ports used by task %+v  and  %+v \n", k, v)
		networkCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_NETWORK).
			Key(k).
			Create()
		commoditiesBought = append(commoditiesBought, networkCommBought)

	}

	return commoditiesBought
}

// Build commodityDTOs for commodity sold by the app
func (TaskProbe *TaskProbe) GetCommoditiesSoldByApp(task *util.Task, taskResourceStat *TaskResourceStat) []*sdk.CommodityDTO {
	appName := task.Name
	var commoditiesSold []*sdk.CommodityDTO
	transactionComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_TRANSACTION).
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
	// From Container
	containerName := task.Id
	containerProvider := sdk.CreateProvider(sdk.EntityDTO_CONTAINER, containerName)
	var commoditiesBought []*sdk.CommodityDTO
	cpuAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, cpuAllocationCommBought)
	memAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, memAllocationCommBought)
	diskAllocationCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key(task.Id).
		Used(taskResourceStat.diskAllocationUsed).
		Create()
	commoditiesBought = append(commoditiesBought, diskAllocationCommBought)

	commoditiesBoughtMap[containerProvider] = commoditiesBought

	// from Virtual Machine
	slaveProvider := sdk.CreateProvider(sdk.EntityDTO_VIRTUAL_MACHINE, task.SlaveId)
	var commoditiesBoughtFromSlave []*sdk.CommodityDTO

	vCpuCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VCPU).
		//		Key(task.SlaveId).
		Used(taskResourceStat.cpuAllocationUsed).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, vCpuCommBought)

	vMemCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VMEM).
		//		Key(task.SlaveId).
		Used(taskResourceStat.memAllocationUsed).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, vMemCommBought)

	appCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_APPLICATION).
		Key(task.SlaveId).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, appCommBought)
	clusterCommBought := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CLUSTER).
		Key(taskProbe.Cluster.ClusterName).
		Create()
	commoditiesBoughtFromSlave = append(commoditiesBoughtFromSlave, clusterCommBought)

	commoditiesBoughtMap[slaveProvider] = commoditiesBoughtFromSlave
	return commoditiesBoughtMap
}

/*
func (TaskProbe *TaskProbe) buildConstraintMap() map[string][][]Strin{
	for i, cons := range taskProbe.Constraints{
		this.ConstraintsMap[]
}
*/
