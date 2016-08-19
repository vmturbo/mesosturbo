package util

import "time"

type Attributes struct {
	Rack string `json:"rack"`
}

type Resources struct {
	Disk  float64 `json:"disk"`
	Mem   float64 `json:"mem"`
	CPUs  float64 `json:"cpus"`
	Ports string  `json:"ports"`
}

type PortUtil struct {
	Number   float64
	Capacity float64
	Used     float64
}

type CalculatedUse struct {
	Disk                 float64
	Mem                  float64
	CPUs                 float64
	CPUsumSystemUserSecs float64
	UsedPorts            []PortUtil
}

type Statistics struct {
	CPUsLimit         float64 `json:"cpus_limit"`
	MemLimitBytes     float64 `json:"mem_limit_bytes"`
	MemRSSBytes       float64 `json:"mem_rss_bytes"`
	CPUsystemTimeSecs float64 `json:"cpus_system_time_secs"`
	CPUuserTimeSecs   float64 `json:"cpus_user_time_secs"`
	DiskLimitBytes    float64 `json:"disk_limit_bytes"`
	DiskUsedBytes     float64 `json:"disk_used_bytes"`
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
	Attributes       Attributes `json:"attributes"`
}

type MesosAPIResponse struct {
	AllPorts          []string
	MApps             *MarathonApps
	Version           string      `json:"version"`
	Id                string      `json:"id"`
	ActivatedSlaves   float64     `json:"activated_slaves"`
	DeActivatedSlaves float64     `json:"deactivated_slaves"`
	Slaves            []Slave     `json:"slaves"`
	Frameworks        []Framework `json:"frameworks"`
	TaskMasterAPI     MasterTasks
	SlaveIdIpMap      map[string]string
	MapTaskStatistics map[string]Statistics
	Leader            string `json:"leader"`
	//Monitor
	TimeSinceLastDisc *time.Time
	SlaveUseMap       map[string]*CalculatedUse
	// TODO use this?
	MapSlaveToTasks map[string][]Task
	//cluster
	Cluster     ClusterInfo
	ClusterName string `json:"cluster"`
}

type ClusterInfo struct {
	ClusterName string
	MasterIP    string
	MasterId    string
}

type PortMapping struct {
	ContainerPort int `json:"containerPort"`
	HostPort      int `json:"hostPort"`
	ServicePort   int `json:"servicePort"`
}

type ContDocker struct {
	ForcePullImage bool          `json:"force_pull_image"`
	Image          string        `json:"image"`
	Network        string        `json:"network"`
	Privileged     bool          `json:"privileged"`
	PortMappings   []PortMapping `json:"portMappings"`
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

type Label struct {
	//	Key   string `json:"key"`
	//	Value string `json:"value"`
	State string `json:"key"`
}

type Task struct {
	Container   Container `json:"container"`
	Discovery   Discovery `json:"discovery"`
	ExecutorId  string    `json:"executor_id"`
	FrameworkId string    `json:"framework_id"`
	Id          string    `json:"id"`
	Labels      []Label   `json:"labels"`
	Name        string    `json:"name"`
	Resources   Resources `json:"resources"`
	SlaveId     string    `json:"slave_id"`
	State       string    `json:"state"`
	Statuses    []Status  `json:"statuses"`
}

// assumed to be framework from slave , not from master state
type Framework struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Pid       string    `json:"pid"`
	Hostname  string    `json:"hostname"`
	Active    bool      `json:"active"`
	Role      string    `json:"role"`
	Resources Resources `json:"resources"`
	Tasks     []Task    `json:"tasks"`
}

type MasterTasks struct {
	Tasks []Task `json:"tasks"`
}

type App struct {
	Name         string     `json:"id"`
	Constraints  [][]string `json:"constraints"`
	RequirePorts bool       `json:"requirePorts"`
	Container    Container  `json:"container"`
}

type MarathonApps struct {
	Apps []App `json:"apps"`
}
