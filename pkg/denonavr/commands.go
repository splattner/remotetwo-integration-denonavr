package denonavr

import (
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

func (d *DenonAVR) sendCommandToDevice(cmd DenonCommand, payload string) (int, error) {

	if d.telnetEnabled {

		err := d.sendTelnetCommand(cmd, payload)
		if err != nil {
			return 404, err
		}

		return 200, nil
	}

	return d.sendHTTPCommand(cmd, payload)
}

func (d *DenonAVR) sendHTTPCommand(denonCommandType DenonCommand, command string) (int, error) {

	url := "http://" + d.Host + COMMAND_URL + "?" + url.QueryEscape(string(denonCommandType)+command)
	log.WithFields(log.Fields{
		"type":    string(denonCommandType),
		"command": command,
		"url":     url}).Info("Send Command to Denon Device")

	req, err := http.Get(url)
	if err != nil {
		return req.StatusCode, fmt.Errorf("error sending command: %w", err)
	}

	// Trigger a update to get updated data handled in the Listen Loop
	d.updateTrigger <- "update"

	return req.StatusCode, nil
}

func (d *DenonAVR) SetMoni1Out() error {
	_, err := d.sendCommandToDevice(DenonCommandVS, "MONI1")

	return err
}

func (d *DenonAVR) SetMoni2Out() error {
	_, err := d.sendCommandToDevice(DenonCommandVS, "MONI2")

	return err
}

func (d *DenonAVR) SetMoniAutoOut() error {
	_, err := d.sendCommandToDevice(DenonCommandVS, "MONIAUTO")

	return err
}
