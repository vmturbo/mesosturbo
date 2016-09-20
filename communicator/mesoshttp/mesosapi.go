package mesoshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"io/ioutil"
	"net/http"
	"strings"
)

type MesosHTTPClient struct {
	MesosMasterBase string
}

func (mesos *MesosHTTPClient) MesosPostRequest(metadata *metadata.ConnectionClient, dcos_token string) error {

	var jsonStr []byte
	url := "http://" + metadata.MesosIP + "/acs/api/v1/auth/login"

	if dcos_token == "" {
		fmt.Println(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `"}`)
		jsonStr = []byte(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `"}`)
	} else {
		fmt.Println(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `","token":"` + dcos_token + `"}`)
		jsonStr = []byte(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `","token":"` + dcos_token + `"}`)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("Error in POST request: %s", err)
		return err
	} else {
		// Get token if response if OK
		if resp.Status == "" {
			fmt.Println("Empty response status")
			return err
		}
		fmt.Printf(" Status is : %s \n", resp.Status)
		statusmsg := strings.Split(resp.Status, " ")
		if statusmsg[0] == "200" {
			fmt.Println(" We got 200, response Status:", statusmsg[0])
			fmt.Println("response Headers:", resp.Header)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error after ioutil.ReadAll: %s", err)
				return err
			}
			byteContent := []byte(body)
			var tokenResp = new(util.TokenResponse)
			err = json.Unmarshal(byteContent, &tokenResp)
			if err != nil {
				fmt.Printf("error in json unmarshal : %s . \r\nLogin failed , please try again with correct credentials.", err)
				return err
			}
			metadata.Token = tokenResp.Token
			fmt.Println("--> token we got is ")
			fmt.Println(tokenResp.Token)
			return nil
		} else {
			fmt.Println("Please check DCOS credentials and start mesosturbo again.")
			return err
		}

	}
}
