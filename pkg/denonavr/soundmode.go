package denonavr

import "strings"

// thx to https://github.com/ol-iver/denonavr/blob/main/denonavr/const.py

// Sound modes
const ALL_ZONE_STEREO = "ALL ZONE STEREO"

var SOUND_MODE_MAPPING = map[string][]string{
	"MUSIC": {
		"PLII MUSIC",
		"DTS NEO:6 MUSIC",
		"DTS NEO:6 M",
		"DTS NEO:X M",
		"DOLBY D +NEO:X M",
		"DTS NEO:X MUSIC",
		"DOLBY PL2 MUSIC",
		"DOLBY PL2 M",
		"PLIIX MUSIC",
		"DOLBY PL2 X MUSIC",
	},
	"MOVIE": {
		"PLII MOVIE",
		"PLII CINEMA",
		"DTS NEO:X CINEMA",
		"DTS NEO:X C",
		"DTS NEO:6 CINEMA",
		"DTS NEO:6 C",
		"DOLBY D +NEO:X C",
		"PLIIX CINEMA",
		"DOLBY PLII MOVIE",
		"MULTI IN + VIRTUAL:X",
		"DOLBY PL2 CINEMA",
		"DOLBY PL2 C",
		"DOLBY PL2 X MOVIE",
	},
	"GAME": {
		"PLII GAME",
		"DOLBY D +NEO:X G",
		"DOLBY PL2 GAME",
		"DOLBY PL2 G",
		"DOLBY PL2 X GAME",
	},
	"AUTO":        {"None"},
	"STANDARD":    {"None2"},
	"VIRTUAL":     {"VIRTUAL"},
	"MATRIX":      {"MATRIX"},
	"ROCK ARENA":  {"ROCK ARENA"},
	"JAZZ CLUB":   {"JAZZ CLUB"},
	"VIDEO GAME":  {"VIDEO GAME"},
	"MONO MOVIE":  {"MONO MOVIE"},
	"DIRECT":      {"DIRECT"},
	"PURE DIRECT": {"PURE_DIRECT", "PURE DIRECT"},
	"DOLBY DIGITAL": {
		"DOLBY DIGITAL",
		"DOLBY D + DOLBY SURROUND",
		"DOLBY D+DS",
		"DOLBY DIGITAL +",
		"STANDARD(DOLBY)",
		"DOLBY SURROUND",
		"DOLBY D + +DOLBY SURROUND",
		"NEURAL",
		"DOLBY HD",
		"DOLBY HD + DOLBY SURROUND",
		"MULTI IN + DSUR",
		"MULTI IN + NEURAL:X",
		"MULTI IN + DOLBY SURROUND",
		"DOLBY D + NEURAL:X",
		"DOLBY DIGITAL + NEURAL:X",
		"DOLBY DIGITAL + + NEURAL:X",
		"DOLBY ATMOS",
		"DOLBY AUDIO - DOLBY SURROUND",
		"DOLBY TRUEHD",
		"DOLBY AUDIO - DOLBY DIGITAL PLUS",
		"DOLBY AUDIO - TRUEHD + DSUR",
		"DOLBY AUDIO - DOLBY TRUEHD",
		"DOLBY AUDIO - TRUEHD + NEURAL:X",
		"DOLBY AUDIO - DD + DSUR",
		"DOLBY AUDIO - DD+   + NEURAL:X",
		"DOLBY AUDIO - DD+   + DSUR",
		"DOLBY AUDIO - DOLBY DIGITAL",
		"DOLBY AUDIO-DSUR",
		"DOLBY AUDIO-DD+DSUR",
	},
	"DTS SURROUND": {
		"DTS SURROUND",
		"DTS NEURAL:X",
		"STANDARD(DTS)",
		"DTS + NEURAL:X",
		"MULTI CH IN",
		"DTS-HD MSTR",
		"DTS VIRTUAL:X",
		"DTS-HD + NEURAL:X",
		"DTS-HD",
		"DTS + VIRTUAL:X",
		"DTS + DOLBY SURROUND",
		"DTS-HD + DOLBY SURROUND",
		"DTS-HD + DSUR",
		"DTS:X MSTR",
	},
	"AURO3D":     {"AURO-3D"},
	"AURO2DSURR": {"AURO-2D SURROUND"},
	"MCH STEREO": {
		"MULTI CH STEREO",
		"MULTI_CH_STEREO",
		"MCH STEREO",
		"MULTI CH IN 7.1",
	},
	"STEREO":        {"STEREO"},
	ALL_ZONE_STEREO: {"ALL ZONE STEREO"},
}

func (d *DenonAVR) GetSoundModeList() []string {

	var soundModeList []string

	for mode := range SOUND_MODE_MAPPING {
		soundModeList = append(soundModeList, mode)
	}

	return soundModeList
}

func (d *DenonAVR) GetSurroundMode(zone DenonZone) string {

	return d.getZoneSurroundMode(d.zoneStatus[zone])
}

func (d *DenonAVR) getZoneSurroundMode(zoneStatus DenonZoneStatus) string {

	var surroundMode string

	rawSurroundMode := strings.TrimRight(zoneStatus.SurrMode, " ")

	if strings.Contains(strings.ToUpper(rawSurroundMode), "DTS") {
		surroundMode = "DTS SURROUND"
	}

	if strings.Contains(strings.ToUpper(rawSurroundMode), "DOLBY DIGITAL") {
		surroundMode = "DOLBY DIGITAL"
	}

	if strings.Contains(strings.ToUpper(rawSurroundMode), "MUSIC") {
		surroundMode = "MUSIC"
	}

	if strings.Contains(strings.ToUpper(rawSurroundMode), "AURO3D") {
		surroundMode = "AURO3D"
	}

	if strings.Contains(strings.ToUpper(rawSurroundMode), "MOVIE") || strings.Contains(strings.ToUpper(rawSurroundMode), "CINEMA") {
		surroundMode = "MOVIE"
	}

	if strings.Contains(strings.ToUpper(rawSurroundMode), "AURO3D") {
		surroundMode = "AURO3D"
	}

	if strings.ToUpper(rawSurroundMode) == "AURO-2D SURROUND" {
		surroundMode = "AURO2DSURR"
	}

	if strings.ToUpper(rawSurroundMode) == "None" {
		surroundMode = "AUTO"
	}

	if strings.ToUpper(rawSurroundMode) == "None2" {
		surroundMode = "STANDARD"
	}

	if surroundMode == "" {
		surroundMode = rawSurroundMode
	}

	return surroundMode
}

func (d *DenonAVR) SetSoundModeMainZone(mode string) int {

	status, _ := d.sendCommandToDevice(DenonCommandMS, strings.ToUpper(mode))
	return status

}
