package cmd

import (
	"github.com/google/logger"
	"gopkg.in/yaml.v2"
)

var def = `
mesosactionip: 10.10.174.91
mesosactionport: 5555
serveraddress: 129.236.238.130:8080
targettype: Mesos
nameoraddress: my_mesos
username: mesos_user1
targetidentifier: my_mesos1
password: fake_password
localaddress: http://10.10.174.96/
websocketusername: vmtRemoteMediation
websocketpassword: vmtRemoteMediation
opsmanagerusername: administrator
opsmanagerpassword: a
`

func ReadYaml() map[interface{}]interface{} {
	s := Settings{}

	err := yaml.Unmarshall([]byte(def), &s)
	if err != nil {
		logger.Fatalf("failed to open the yaml, error: %v", err)
	}

	d, err := yaml.Marshal(&s)
	if err != nil {
		logger.Fatalf("error : %v", err)
	}

	m := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(def), &m)
	if err != nil {
		logger.Fatalf("error: %v", err)
	}
	return m
}
