package main

import (
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/cmd/simulation/builder"
	"github.com/vmturbo/mesosturbo/pkg/action"
)

func main() {
	simbuilder := builder.NewSimulatorBuilder()
	// TODO flags
	// TODO init with flags
	simulator, err := simbuilder.Build()
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
	//	actor := action.RequestMesosAction(simulator.MesosClient())
	_, err = action.RequestMesosAction(simulator)
	if err != nil {
		glog.Errorf("error %s \n", err)
	}
}
