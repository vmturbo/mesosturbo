package probe

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
	"strconv"
)

type NodeResourceStat struct {
	diskAllocationCapacity float64
	diskAllocationUsed     float64
	cpuAllocationCapacity  float64
	cpuAllocationUsed      float64
	memAllocationCapacity  float64
	memAllocationUsed      float64
	vCpuCapacity           float64
	vCpuUsed               float64
	vMemCapacity           float64
	vMemUsed               float64
}

// models the probe for a mesos master state , containing metadata about
// all of the slaves
type NodeProbe struct {
	MasterState   *util.MesosAPIResponse
	Cluster       *util.ClusterInfo
	AllSlavePorts []string
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
	// Mem is returned in B
	nodeMemCapacity := slaveInfo.Resources.Mem
	nodeDiskCapacity := slaveInfo.Resources.Disk
	glog.V(4).Infof("Discovered node is %f\n", slaveInfo.Id)
	glog.V(4).Infof("Node CPU capacity is %f \n", nodeCpuCapacity)
	glog.V(4).Infof("Node Mem capacity is %f \n", nodeMemCapacity)
	glog.V(4).Infof("Node Disk capacity is %f \n", nodeDiskCapacity)

	// Find out the used value for each commodity
	cpuUsed := useMap[slaveInfo.Id].CPUs
	memUsed := useMap[slaveInfo.Id].Mem
	diskUsed := slaveInfo.UsedResources.Disk
	glog.V(4).Infof("Discovered node is %f\n", slaveInfo.Id)
	glog.V(4).Infof("=======> Node CPU used is %f \n", cpuUsed)
	glog.V(4).Infof("Node Mem used is %f \n", memUsed)
	glog.V(4).Infof("Node Disk used is %f \n", diskUsed)

	return &NodeResourceStat{
		diskAllocationCapacity: nodeDiskCapacity,
		diskAllocationUsed:     diskUsed,
		cpuAllocationCapacity:  nodeCpuCapacity,
		cpuAllocationUsed:      cpuUsed,
		memAllocationCapacity:  nodeMemCapacity,
		memAllocationUsed:      memUsed,
		vCpuCapacity:           nodeCpuCapacity,
		vCpuUsed:               cpuUsed,
		vMemCapacity:           nodeMemCapacity,
		vMemUsed:               memUsed,
	}, nil

}

func (nodeProbe *NodeProbe) CreatePortConstraints(useMap map[string]*util.CalculatedUse) {
	for _, port := range nodeProbe.AllSlavePorts {
		glog.V(3).Infof(" port now is %d", port)
		usedPort, err := strconv.Atoi(port)
		if err != nil {
			glog.V(3).Infof(" error in ports list")
		}
		for _, use := range useMap {
			// check if slave is using port
			if val, ok := use.UsedPorts[port]; ok {
				//use.UsedPorts[port].Used = float64(1.0)
				glog.V(3).Infof(" slave now is used with cap %f \n", val.Used)
			} else {
				use.UsedPorts[port] = util.PortUtil{
					Number:   float64(usedPort),
					Used:     float64(0.0),
					Capacity: float64(1.0),
				}
			}
		}
	}
}

// Get current stat of node resources, such as capacity and used values.

func (nodeProbe *NodeProbe) CreateCommoditySold(slaveInfo *util.Slave, useMap map[string]*util.CalculatedUse) ([]*sdk.CommodityDTO, error) {
	var commoditiesSold []*sdk.CommodityDTO
	nodeResourceStat, err := nodeProbe.getNodeResourceStat(slaveInfo, useMap)
	if err != nil {
		return commoditiesSold, err
	}

	// create labels for VM node
	var labels = []string{}
	if slaveInfo.Attributes.Rack != "" {
		strkey := "rack"
		strval := slaveInfo.Attributes.Rack
		glog.V(3).Infof("----------------> rack is %s", strval)
		labels = append(labels, strkey+":"+strval)
		glog.V(3).Infof("====================> labels : %+v", labels)
	}

	//TODO: create const value for keys
	memAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_MEM_ALLOCATION).
		Key("Mesos").
		Capacity(float64(nodeResourceStat.memAllocationCapacity)).
		Used(nodeResourceStat.memAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, memAllocationComm)
	cpuAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CPU_ALLOCATION).
		Key("Mesos").
		Capacity(float64(nodeResourceStat.cpuAllocationCapacity)).
		Used(nodeResourceStat.cpuAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, cpuAllocationComm)
	diskAllocationComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_STORAGE_ALLOCATION).
		Key("Mesos").
		Capacity(float64(nodeResourceStat.diskAllocationCapacity)).
		Used(nodeResourceStat.diskAllocationUsed).
		Create()
	commoditiesSold = append(commoditiesSold, diskAllocationComm)
	vMemComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VMEM).
		//Key(slaveInfo.Id).
		Capacity(nodeResourceStat.vMemCapacity).
		Used(nodeResourceStat.vMemUsed).
		Create()
	commoditiesSold = append(commoditiesSold, vMemComm)
	vCpuComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VCPU).
		//Key(slaveInfo.Id).
		Capacity(float64(nodeResourceStat.vCpuCapacity)).
		Used(nodeResourceStat.vCpuUsed).
		Create()
	commoditiesSold = append(commoditiesSold, vCpuComm)
	appComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_APPLICATION).
		Key(slaveInfo.Id).
		Create()
	commoditiesSold = append(commoditiesSold, appComm)
	clusterComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_CLUSTER).
		Key(nodeProbe.Cluster.ClusterName).
		Create()
	commoditiesSold = append(commoditiesSold, clusterComm)

	// TODO add port commodity sold for now
	nodeProbe.CreatePortConstraints(useMap)
	glog.V(2).Infof("----> used ports are: %s", useMap[slaveInfo.Id].UsedPorts)
	ports := useMap[slaveInfo.Id].UsedPorts
	for port, portObj := range ports {
		portComm := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_NETWORK).
			Key(port). // port number in string form
			Capacity(portObj.Capacity).
			Used(portObj.Used).
			Create()
		commoditiesSold = append(commoditiesSold, portComm)
	}

	// add labels
	for _, label := range labels {
		vmpmAccessCommBuilder := sdk.NewCommodityDTOBuilder(sdk.CommodityDTO_VMPM_ACCESS)
		vmpmAccessCommBuilder = vmpmAccessCommBuilder.Key(label)
		vmpmAccessComm := vmpmAccessCommBuilder.Create()
		commoditiesSold = append(commoditiesSold, vmpmAccessComm)
	}
	return commoditiesSold, nil
}
