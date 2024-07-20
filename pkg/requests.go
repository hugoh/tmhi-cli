package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
)

func reboot(username, password, ip string, dryRun bool) {
	loginData := login(username, password, ip)
	if !loginData.Success {
		log.Println("Cannot reboot without successful login flow")
		return
	}
	rebootRequest := map[string]interface{}{
		"uri": fmt.Sprintf("http://%s/reboot_web_app.cgi", ip),
		"headers": map[string]string{
			"Cookie": fmt.Sprintf("sid=%s", loginData.SID),
		},
		"body": map[string]string{
			"csrf_token": loginData.CSRFToken,
		},
	}
	logrus.Debug(fmt.Sprintf("Reboot request: %+v", rebootRequest))
	rebootMsg := "T-Mobile Internet Router reboot successfully requested"
	if !dryRun {
		httpPost(rebootRequest, func(rebootResp *http.Response) {
			defer rebootResp.Body.Close()
			var respData map[string]interface{}
			json.NewDecoder(rebootResp.Body).Decode(&respData)
			if rebootResp.StatusCode == http.StatusOK {
				logrus.Debug(fmt.Sprintf("Reboot response: %+v", respData))
				log.Println(rebootMsg)
			} else {
				log.Println(fmt.Sprintf("Reboot request failed: %+v", respData))
			}
		})
	} else {
		log.Println(fmt.Sprintf("[DRY-RUN] %s [/DRY-RUN]", rebootMsg))
	}
}

func httpPost(request map[string]interface{}, handler func(resp *http.Response)) {
	uri := request["uri"].(string)
	body := request["body"].(map[string]string)
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Println("Error marshalling request body:", err)
		return
	}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating HTTP request:", err)
		return
	}
	for k, v := range request["headers"].(map[string]string) {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making HTTP request:", err)
		return
	}
	handler(resp)
}
