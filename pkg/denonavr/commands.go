package denonavr

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

func (d *DenonAVR) sendCommandToDevice(cmd DenonCommand, payload string) (int, error) {

	if d.telnetEnabled {

		err := d.sendTelnetCommand(cmd, payload)
		if err != nil {
			// Fallback to HTTP
			return d.sendHTTPCommand(cmd, payload)
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

	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error sending command: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Trigger a update to get updated data handled in the Listen Loop
	d.updateTrigger <- "update"

	return resp.StatusCode, nil
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
