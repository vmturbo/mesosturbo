package action

type MesosClient struct {
	MesosMasterIP   string
	MesosMasterPort string
	Action          string
	DestinationId   string
	TaskId          string
}

type TaskProvider struct {
	Id string
}

type PendingTask struct {
	TaskProvider   TaskProvider
	Kill_requested bool
	Name           string
	Id             string
}
