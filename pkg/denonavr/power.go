package denonavr

import (
	log "github.com/sirupsen/logrus"
)

func (d *DenonAVR) TurnOn() error {
	if _, err := d.sendCommandToDevice(DenonCommandPower, "ON"); err != nil {
		return err
	}
	_, err := d.sendCommandToDevice(DennonCommandZoneMain, "ON")

	return err
}

func (d *DenonAVR) TurnOff() error {

	if _, err := d.sendCommandToDevice(DenonCommandPower, "STANDBY"); err != nil {
		return err
	}
	_, err := d.sendCommandToDevice(DennonCommandZoneMain, "OFF")

	return err
}

func (d *DenonAVR) TogglePower() error {

	if d.IsOn() {
		return d.TurnOff()
	}

	return d.TurnOn()
}

func (d *DenonAVR) IsOn() bool {

	mainzonepower, err := d.GetAttribute("MainZonePower")
	if err != nil {
		log.WithError(err).Error("MainZonePower attribute not found")
		return false
	}

	switch mainzonepower.(string) {
	case "ON":
		return true
	default:
		return false
	}
}
