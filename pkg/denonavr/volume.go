package denonavr

import (
	"fmt"
	"math"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (d *DenonAVR) SetVolume(volume float64) error {

	// The Volume command need the following
	// 10.5 -> MV105
	// 11 -> MV11

	var convertedVolume string

	if volume != math.Trunc(volume) {
		convertedVolume = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", volume*10), "0"), ".")
	} else {
		convertedVolume = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", volume), "0"), ".")
	}

	_, err := d.sendCommandToDevice(DenonCommandMainZoneVolume, convertedVolume)
	return err

}

func (d *DenonAVR) MainZoneMute() error {

	_, err := d.sendCommandToDevice(DenonCommandMainZoneMute, "ON")
	return err
}

func (d *DenonAVR) MainZoneUnMute() error {

	_, err := d.sendCommandToDevice(DenonCommandMainZoneMute, "OFF")
	return err
}

func (d *DenonAVR) MainZoneMuteToggle() error {

	if d.MainZoneMuted() {
		return d.MainZoneUnMute()

	}
	return d.MainZoneMute()
}

func (d *DenonAVR) MainZoneMuted() bool {

	mainZoneMute, err := d.GetAttribute("MainZoneMute")
	if err != nil {
		log.WithError(err).Debug("MainZoneMute attribute not found")
		return false
	}

	switch mainZoneMute.(string) {
	case "on":
		return true
	default:
		return false
	}
}

func (d *DenonAVR) SetVolumeUp() error {

	_, err := d.sendCommandToDevice(DenonCommandMainZoneVolume, "UP")
	return err

}

func (d *DenonAVR) SetVolumeDown() error {
	_, err := d.sendCommandToDevice(DenonCommandMainZoneVolume, "DOWN")
	return err
}
