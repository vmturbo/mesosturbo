package probe

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
)

type NodeResourceStat struct {
	cpuAllocationCapacity float64
	cpuAllocationUsed     float64
	memAllocationCapacity float64
	memAllocationUsed     float64
	vCpuCapacity          float64
	vCpuUsed              float64
	vMemCapacity          float64
	vMemUsed              float64
}

// models the probe for a mesos master state , containing metadata about
// all of the slaves
type NodeProbe struct {
	MasterState *util.MesosAPIResponse
}

// Get current stat of node resources, such as capacity and used values.
func (nodeProbe *NodeProbe) getNodeResourceStat(slaveInfo *util.Slave, useMap map[string]*util.CalculatedUse) (*NodeResourceStat, error) {
	// The return cpu frequency is in KHz, we need MHz
	// TODO pam , how to get slave frequency?
	//	cpuFrequency := 400 // machineInfo.CpuFrequency / 1000
	// TODO pam	nodeMachineInfoMap[node.Name] = machineInfo

	// Here we only need the root container.
	// To get a valid cpu usage, there must be at least 2 valid stats.
	//	intervalInNs := currentStat.Timestamp.Sub(prevStat.Timestamp).Nanoseconds()
	//	glog.V(4).Infof("interval is %d", intervalInNs)
	//	currentStat := containerStats[len(containerStats)-1]
	//	prevStat := containerStats[len(containerStats)-2]
	//	rawUsage := int64(currentStat.Cpu.Usage.Total - prevStat.Cpu.Usage.Total)
	//	glog.V(4).Infof("rawUsage is %d", rawUsage)
	//	intervalInNs := currentStat.Timestamp.Sub(prevStat.Timestamp).Nanoseconds()
	//	glog.V(4).Infof("interval is %d", intervalInNs)
	//	rootCurCpu := slaveInfo.UsedResouces.CPUs//float64(rawUsage) * 1.0 / float64(intervalInNs)
	//	rootCurMem := slaveInfo.UsedResources.Mem//float64(currentStat.Memory.Usage) / 1024 // Mem is returned in B

	// Get the node Cpu and Mem capacity.
	nodeCpuCapacity := slaveInfo.Resources.CPUs * float64(2000) //float64(slaveInfo.Resources.CPUs) * float64(cpuFrequency)
	nodeMemCapacity := slaveInfo.Resources.Mem                  //float64(slaveInfo.Resources.Mem) / 1024 // Mem is returned in B
	glog.V(4).Infof("Discovered node is %f\n", slaveInfo.Id)
	glog.V(4).Infof("Node CPU capacity is %f \n", nodeCpuCapacity)
	glog.V(4).Infof("Node Mem capacity is %f \n", nodeMemCapacity)
	// Find out the used value for each commodity
	cpuUsed := useMap[slaveInfo.Id].CPUs   //float64(rootCurCpu) * float64(cpuFrequency)
	memUsed := slaveInfo.UsedResources.Mem //float64(rootCurMem)
	glog.V(4).Infof("Discovered node is %f\n", slaveInfo.Id)
	glog.V(4).Infof("=======> Node CPU used is %f \n", cpuUsed)
	glog.V(4).Infof("Node Mem used is %f \n", memUsed)

	return &NodeResourceStat{
		cpuAllocationCapacity: nodeCpuCapacity,
		cpuAllocationUsed:     cpuUsed,
		memAllocationCapacity: nodeMemCapacity,
		memAllocationUsed:     memUsed,
		vCpuCapacity:          nodeCpuCapacity,
		vCpuUsed:              cpuUsed,
		vMemCapacity:          nodeMemCapacity,
		vMemUsed:              memUsed,
	}, nil

}

// Get current stat of node resources, such as capacity and used values.

func (nodeProbe *NodeProbe) CreateCommoditySold(slaveInfo *util.Slave, useMap map[string]*util.CalculatedUse) ([]*sdk.CommodityDTO, error) {
	var commoditiesSold []*sdk.CommodityDTO
	nodeResourceStat, err := nodeProbe.getNodeResourceStat(slaveInfo, useMap)
	if err != nil {
		return commoditiesSold, err
	}

	//TODO: create const value for keys
	memAllocationComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key("Mesos").
		Capacity(float64(nodeResourceStat.memAllocationCapacity)).
		Used(nodeResourceStat.memAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, memAllocationComm)
	cpuAllocationComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key("Mesos").
		Capacity(float64(nodeResourceStat.cpuAllocationCapacity)).
		Used(nodeResourceStat.cpuAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, cpuAllocationComm)
	vMemComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VMEM).
		//Key(slaveInfo.Id).
		Capacity(nodeResourceStat.vMemCapacity).
		Used(nodeResourceStat.vMemUsed).
		Create()
	commoditiesSold = append(commoditiesSold, vMemComm)
	vCpuComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_VCPU).
		//Key(slaveInfo.Id).
		Capacity(float64(nodeResourceStat.vCpuCapacity)).
		Used(nodeResourceStat.vCpuUsed).
		Create()
	commoditiesSold = append(commoditiesSold, vCpuComm)
	appComm := sdk.NewCommodtiyDTOBuilder(sdk.CommodityDTO_APPLICATION).
		Key(slaveInfo.Id).
		Create()
	commoditiesSold = append(commoditiesSold, appComm)
	return commoditiesSold, nil
}
