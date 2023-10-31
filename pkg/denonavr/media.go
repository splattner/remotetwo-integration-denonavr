package denonavr

import (
	"fmt"
	"hash/fnv"

	"k8s.io/utils/strings/slices"
)

func (d *DenonAVR) Play() int {
	status, _ := d.sendCommandToDevice(DenonCommandNS, "9A")

	return status
}

func (d *DenonAVR) Pause() int {
	status, _ := d.sendCommandToDevice(DenonCommandNS, "9B")

	return status
}

// Get the current Media Title
// Title of the Playing media or the current Input Function
func (d *DenonAVR) getMediaTitle() string {
	media_title := ""

	if d.IsOn() {
		if slices.Contains(PLAYING_SOURCES, d.mainZoneData.InputFuncSelect) {
			// This is a source that is playing audio
			media_title = d.netAudioStatus.SzLine[1]
		} else {
			// Not a playing source
			media_title = d.mainZoneData.InputFuncSelect
		}
	}

	d.SetAttribute("media_title", media_title)
	return d.attributes["media_title"].(string)
}

// Get the current Media Title
// Title of the Playing media or the current Input Function
func (d *DenonAVR) getMediaImageURL() string {
	media_image_url := ""

	if d.IsOn() {
		if slices.Contains(PLAYING_SOURCES, d.mainZoneData.InputFuncSelect) {
			// This is a source that is playing audio
			// fot the moment, also set this to the input func

			hash := fnv.New32a()
			hash.Write([]byte(d.getMediaTitle()))
			media_image_url = fmt.Sprintf("http://%s:%d/NetAudio/art.asp-jpg?%d", d.Host, 80, hash.Sum32())
		} else {
			media_image_url = fmt.Sprintf("http://%s:%d/", d.Host, 80) + "img/album%20art_S.png"
		}
	}

	d.SetAttribute("media_image_url", media_image_url)
	return d.attributes["media_image_url"].(string)
}
