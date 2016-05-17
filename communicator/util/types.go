package util

import "time"

type Resources struct {
	Disk float64 `json:"disk"`
	Mem  float64 `json:"mem"`
	CPUs float64 `json:"cpus"`
}

type CalculatedUse struct {
	Disk                 float64
	Mem                  float64
	CPUs                 float64
	CPUsumSystemUserSecs float64
}

type Statistics struct {
	CPUsLimit         float64 `json:"cpus_limit"`
	MemLimitBytes     float64 `json:"mem_limit_bytes"`
	MemRSSBytes       float64 `json:"mem_rss_bytes"`
	CPUsystemTimeSecs float64 `json:"cpus_system_time_secs"`
	CPUuserTimeSecs   float64 `json:"cpus_user_time_secs"`
}

type Executor struct {
	Id         string     `json:"executor_id"`
	Source     string     `json:"source"`
	Statistics Statistics `json:"statistics"`
}

type Slave struct {
	Id               string    `json:"id"`
	Pid              string    `json:"pid"`
	Resources        Resources `json:"resources"`
	UsedResources    Resources `json:"used_resources"`
	OfferedResources Resources `json:"offered_resources"`
	Name             string    `json:"hostname"`
	Calculated       CalculatedUse
}

type MesosAPIResponse struct {
	Version           string  `json:"version"`
	Id                string  `json:"id"`
	ActivatedSlaves   float64 `json:"activated_slaves"`
	DeActivatedSlaves float64 `json:"deactivated_slaves"`
	Slaves            []Slave `json:"slaves"`
	//Frameworks        []Framework `json:"frameworks"`
	TaskMasterAPI     MasterTasks
	SlaveIdIpMap      map[string]string
	MapTaskStatistics map[string]Statistics
	//Monitor
	TimeSinceLastDisc *time.Time
	SlaveUseMap       map[string]*CalculatedUse
	// TODO use this?
	MapSlaveToTasks map[string][]Task
}

type ContDocker struct {
	ForcePullImage bool   `json:"force_pull_image"`
	Image          string `json:"image"`
	Network        string `json:"network"`
	Privileged     bool   `json:"privileged"`
}

type Container struct {
	Docker ContDocker `json:"docker"`
	Type   string     `json"type"`
}

type PortInfo struct {
	Number   int64  `json:"number"`
	Protocol string `json:"protocol"`
}

type DiscPorts struct {
	Ports []PortInfo `json:"ports"`
}

type Discovery struct {
	Name       string    `json:"name"`
	Ports      DiscPorts `json:"ports"`
	Visibility string    `json:"visibility"`
}

type NetworkInfo struct {
	IPaddress string `json:"ip_address"`
}

type NetworkInfos struct {
	Infos []NetworkInfo `json:"network_infos"`
}

type Status struct {
	Container_Status NetworkInfos `json:"container_status"`
	State            string       `json:"state"`
	//	Timestamp `json:"timestamp"`
}

type Task struct {
	Container   Container `json:"container"`
	Discovery   Discovery `json:"discovery"`
	ExecutorId  string    `json:"executor_id"`
	FrameworkId string    `json:"framework_id"`
	Id          string    `json:"id"`
	Resources   Resources `json:"resources"`
	SlaveId     string    `json:"slave_id"`
	Name        string    `json:"name"`
	Statuses    []Status  `json:"statuses"`
	State       string    `json:"state"`
}

type MasterTasks struct {
	Tasks []Task `json:"tasks"`
}
