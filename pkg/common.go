package pkg

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func doReboot(client *http.Client, req *http.Request, dryRun bool) error {
	if !dryRun {
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending reboot request: %w", err)
		}
		defer resp.Body.Close()

		logrus.WithFields(LogHTTPResponseFields(resp)).Debug("reboot response")
		if !HTTPRequestSuccessful(resp) {
			logrus.WithFields(LogHTTPResponseFields(resp)).Error("reboot failed")
			return ErrRebootFailed
		}
	}
	logrus.Info("successfully requested gateway rebooted")
	return nil
}
