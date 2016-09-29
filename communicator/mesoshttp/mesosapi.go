package mesoshttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	"github.com/vmturbo/mesosturbo/communicator/metadata"
	"github.com/vmturbo/mesosturbo/communicator/util"
	"io/ioutil"
	"net/http"
)

type MesosHTTPClient struct {
	MesosMasterBase string
}

func (mesos *MesosHTTPClient) DCOSLoginRequest(metadata *metadata.ConnectionClient, dcos_token string) error {

	var jsonStr []byte
	url := "http://" + metadata.MesosIP + "/acs/api/v1/auth/login"

	if dcos_token == "" {
		glog.V(3).Infof(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `"}`)
		jsonStr = []byte(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `"}`)
	} else {
		glog.V(3).Infof(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `","token":"` + dcos_token + `"}`)
		jsonStr = []byte(`{"uid":"` + metadata.DCOS_Username + `","password":"` + metadata.DCOS_Password + `","token":"` + dcos_token + `"}`)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		glog.Errorf("Error in POST request: %s \n", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		glog.Errorf("Error in POST request: %s \n", err)
		return err
	} else {
		// Get token if response if OK
		defer resp.Body.Close()
		if resp.Status == "" {
			glog.Errorf("Empty response status \n")
			return errors.New("Empty response status \n")
		}

		glog.Infof(" Status is : %s \n", resp.Status)

		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Errorf("Error in ioutil.ReadAll: %s \n", err)
				return err
			}
			byteContent := []byte(body)
			var tokenResp = new(util.TokenResponse)
			err = json.Unmarshal(byteContent, &tokenResp)
			if err != nil {
				glog.Errorf("error in json unmarshal : %s . \r\nLogin failed , please try again with correct credentials.\n", err)
				return err
			}
			metadata.Token = tokenResp.Token
			return nil
		} else {
			glog.Errorf("Please check DCOS credentials and start mesosturbo again.\n")
			return errors.New("DCOS authorization credentials are not correct, check mesosturbo arguments --dcos-uid , --dcos-pwd, or --token! \n")
		}

	}
}
