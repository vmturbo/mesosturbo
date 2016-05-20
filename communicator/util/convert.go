package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetSlaveIP(s Slave) string {
	//"slave(1)@10.10.174.92:5051"
	var ipportArray []string
	slaveIP := ""
	ipLong := s.Pid
	arr := strings.Split(ipLong, "@")
	if len(arr) > 1 {
		ipport := arr[1]
		ipportArray = strings.Split(ipport, ":")
	}
	if len(ipportArray) > 0 {
		slaveIP = ipportArray[0]
	}
	return slaveIP
}

func CreateSlaveIpIdMap(resp *http.Response) (map[string]string, error) {
	fmt.Println("----> in parseAPICallResponse")
	if resp == nil {
		return nil, fmt.Errorf("response sent in is nil")
	}
	//	glog.V(3).Infof(" from glog response body is %s", resp.Body)

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}

	fmt.Println(" response was: %s", string(content))

	//	glog.V(4).Infof("response content is %s", string(content))
	byteContent := []byte(content)
	var jsonMesosMaster = new(MesosAPIResponse)
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
		slaveIP := GetSlaveIP(slaves[i])
		SlaveIpIdMap[slaveIP] = s.Id
	}
	return SlaveIpIdMap, nil
}
