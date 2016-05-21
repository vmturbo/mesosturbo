package action

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type migration struct {
	destination_node_id string
	task_ids            []string
}

func RequestMesosAction(mesosClient *MesosClient) (string, error) {
	baseUrl := "http://" + mesosClient.MesosMasterIP + ":" + mesosClient.MesosMasterPort + "/" + mesosClient.Action + "?"
	//fullUrl := baseUrl + "destination_node_id=32f951d7-52f8-4842-ae1f-eb8d7ec6ac94-S0&task_ids=basic-0.6432abd7-179f-11e6-9521-52540006b4aa"
	fmt.Println(" --> The full Url is ", baseUrl)
	var jsonStr []byte
	if mesosClient.Action == "MigrateTasks" {
		jsonStr = []byte(`{"destination_node_id":"` + mesosClient.DestinationId + `", "task_ids": ["` + mesosClient.TaskId + `"]}`)
	}
	if mesosClient.Action == "AssignTasks" {
		jsonStr = []byte(`{"node_id":"` + mesosClient.DestinationId + `", "task_ids": ["` + mesosClient.TaskId + `"]}`)
	}
	fmt.Printf("payload is :  %+v \n", `{"node_id":"`+mesosClient.DestinationId+`", "task_ids": ["`+mesosClient.TaskId+`"]}`)
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf(" --> error %s \n", err)
	}
	defer resp.Body.Close()
	fmt.Printf("----> request is : %+v\n", req)
	fmt.Printf("response status %s Headers: %s \n", resp.Header, resp.Status)
	if !strings.Contains(resp.Header.Get("Status"), "202 Accepted") {
		fmt.Println("Error while migrating tasks \n")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return string(body), nil
}

func RequestPendingTasks(mesosClient *MesosClient) ([]*PendingTask, error) {
	// 10.10.174.96:5555/GetPendingTasks
	baseUrl := "http://" + mesosClient.MesosMasterIP + ":" + mesosClient.MesosMasterPort + "/" + "GetPendingTasks"
	//fullUrl := baseUrl + "destination_node_id=32f951d7-52f8-4842-ae1f-eb8d7ec6ac94-S0&task_ids=basic-0.6432abd7-179f-11e6-9521-52540006b4aa"
	fmt.Println(" --> The full Url is ", baseUrl)
	req, err := http.NewRequest("GET", baseUrl, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf(" --> error %s \n", err)
		return nil, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error after ioutil.ReadAll: %s", err)
		return nil, err
	}
	var pendingTasks = new([]*PendingTask)
	byteContent := []byte(content)
	err = json.Unmarshal(byteContent, &pendingTasks)
	if err != nil {
		fmt.Printf("JSON error in getPendingTasks %s", err)
		return nil, err
	}
	var pendingTaskArray []*PendingTask
	pendingTaskArray = *pendingTasks
	defer resp.Body.Close()
	return pendingTaskArray, nil
}
