package denonavr

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/utils/strings/slices"
)

var SOURCE_MAPPING = map[string]string{
	"TV AUDIO":       "TV",
	"iPod/USB":       "USB/IPOD",
	"Bluetooth":      "BT",
	"Blu-ray":        "BD",
	"CBL/SAT":        "SAT/CBL",
	"NETWORK":        "NET",
	"Media Player":   "MPLAY",
	"AUX":            "AUX1",
	"Tuner":          "TUNER",
	"FM":             "TUNER",
	"SpotifyConnect": "Spotify Connect",
}

var CHANGE_INPUT_MAPPING = map[string]string{
	"Favorites":      "FAVORITES",
	"Flickr":         "FLICKR",
	"Internet Radio": "IRADIO",
	"Media Server":   "SERVER",
	"Online Music":   "NET",
	"Spotify":        "SPOTIFY",
}

var TELNET_SOURCES = []string{
	"CD",
	"PHONO",
	"TUNER",
	"DVD",
	"BD",
	"TV",
	"SAT/CBL",
	"MPLAY",
	"GAME",
	"HDRADIO",
	"NET",
	"PANDORA",
	"SIRIUSXM",
	"SOURCE",
	"LASTFM",
	"FLICKR",
	"IRADIO",
	"IRP",
	"SERVER",
	"FAVORITES",
	"AUX1",
	"AUX2",
	"AUX3",
	"AUX4",
	"AUX5",
	"AUX6",
	"AUX7",
	"BT",
	"USB/IPOD",
	"USB DIRECT",
	"IPOD DIRECT",
}

var TELNET_MAPPING = map[string]string{
	"FAVORITES": "Favorites",
	"FLICKR":    "Flickr",
	"IRADIO":    "Internet Radio",
	"IRP":       "Internet Radio",
	"SERVER":    "Media Server",
	"SPOTIFY":   "Spotify",
}

var NETAUDIO_SOURCES = []string{
	"AirPlay",
	"Online Music",
	"Media Server",
	"iPod/USB",
	"Bluetooth",
	"Internet Radio",
	"Favorites",
	"SpotifyConnect",
	"Flickr",
	"NET/USB",
	"Music Server",
	"NETWORK",
	"NET",
}

var TUNER_SOURCES = []string{
	"Tuner",
	"TUNER",
}

var HDTUNER_SOURCES = []string{
	"HD Radio",
	"HDRADIO",
}

var PLAYING_SOURCES = append(append(append(NETAUDIO_SOURCES, NETAUDIO_SOURCES...), TUNER_SOURCES...), HDTUNER_SOURCES...)

func (d *DenonAVR) GetZoneInputFuncList(zone DenonZone) map[string]string {

	inputFuncList := make(map[string]string)

	// Only add those not deleted
	// Use renamed value
	for i, input := range d.zoneStatus[zone].InputFuncList {
		// only the ones active or empty (== Online Music)
		if d.zoneStatus[zone].SourceDelete[i] == "USE" || d.zoneStatus[zone].SourceDelete[i] == "" {
			inputFuncList[input] = strings.TrimRight(d.zoneStatus[zone].RenameSource[i], " ")
		}
	}

	return inputFuncList

}

func (d *DenonAVR) SetSelectSourceMainZone(source string) int {

	inputFuncList := d.GetZoneInputFuncList(MainZone)
	log.WithFields(log.Fields{
		"source":        source,
		"inputFuncList": inputFuncList,
		"sourceMapping": SOURCE_MAPPING}).Debug("Select Source Main Zone")

	var selectedSource string
	for sourceOrigin, renamedSource := range inputFuncList {
		if renamedSource == source {
			selectedSource = sourceOrigin
			break
		}
	}

	if SOURCE_MAPPING[selectedSource] != "" {
		selectedSource = SOURCE_MAPPING[selectedSource]
	}

	if slices.Contains(TELNET_SOURCES, selectedSource) {
		status, _ := d.sendCommandToDevice(DenonCommandSelectInput, selectedSource)
		return status
	}

	return 404
}
