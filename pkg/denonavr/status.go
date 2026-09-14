package denonavr

import (
	"encoding/xml"
	"io"

	log "github.com/sirupsen/logrus"
)

type DenonZoneStatus struct {
	XMLName         xml.Name `xml:"item"`
	Zone            string   `xml:"Zone>value"`
	Power           string   `xml:"Power>value"`
	InputFuncList   []string `xml:"InputFuncList>value"`
	RenameSource    []string `xml:"RenameSource>value>value"`
	SourceDelete    []string `xml:"SourceDelete>value"`
	InputFuncSelect string   `xml:"InputFuncSelect>value"`
	VolumeDisplay   string   `xml:"VolumeDisplay>value"`
	RestorerMode    string   `xml:"RestorerMode>value"`
	SurrMode        string   `xml:"SurrMode>value"`
	MasterVolume    string   `xml:"MasterVolume>value"`
	Mute            string   `xml:"Mute>value"`
	Model           string   `xml:"Model>value"`
}

type DenonNetAudioStatus struct {
	XMLName xml.Name `xml:"item"`
	SzLine  []string `xml:"szLine>value"`
}

func (d *DenonAVR) getZoneStatus(zone DenonZone) DenonZoneStatus {
	var url string
	switch zone {
	case MainZone:
		url = "http://" + d.Host + STATUS_URL
	case Zone2:
		url = "http://" + d.Host + STATUS_Z2_URL
	case Zone3:
		url = "http://" + d.Host + STATUS_Z3_URL
	}

	zoneStatus, err := d.getZoneStatusFromDevice(url)
	if err != nil {
		// Keep the last known status instead of crashing or blanking it out
		// on a transient network error.
		return d.zoneStatus[zone]
	}
	d.zoneStatus[zone] = *zoneStatus

	return d.zoneStatus[zone]

}

func (d *DenonAVR) getNetAudioStatus() {
	url := "http://" + d.Host + NET_AUDIO_STATUR_URL
	if status, err := d.getNetAudioStatusFromDevice(url); err == nil {
		d.netAudioStatus = status
	}
}

// Return the Status from a Zone
func (d *DenonAVR) getZoneStatusFromDevice(url string) (*DenonZoneStatus, error) {
	status := DenonZoneStatus{} // Somehow the values in the array are added instead of replaced. Not sure if this is the solution, but it works...
	resp, err := httpClient.Get(url)
	if err != nil {
		log.WithError(err).Error("Failed to get data from Denon AVR")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Cannot read response body")
		return nil, err
	}

	if err := xml.Unmarshal(body, &status); err != nil {
		log.WithError(err).Info("Could not unmarshall")
		return nil, err
	}

	return &status, nil
}

// Return the Status from a Zone
func (d *DenonAVR) getNetAudioStatusFromDevice(url string) (DenonNetAudioStatus, error) {
	status := DenonNetAudioStatus{} // Somehow the values in the array are added instead of replaced. Not sure if this is the solution, but it works...
	resp, err := httpClient.Get(url)
	if err != nil {
		log.WithError(err).Error("Failed to get data from Denon AVR")
		return status, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Cannot read response body")
		return status, err
	}

	if err := xml.Unmarshal(body, &status); err != nil {
		log.WithError(err).Info("Could not unmarshall")
		return status, err
	}

	return status, nil
}
