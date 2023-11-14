package denonavr

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/ziutek/telnet"
)

type DenonCommand string
type DenonZone string

const (
	DenonCommandPower          DenonCommand = "PW"
	DennonCommandZoneMain      DenonCommand = "ZM"
	DenonCommandMainZoneVolume DenonCommand = "MV"
	DenonCommandMainZoneMute   DenonCommand = "MU"
	DenonCommandSelectInput    DenonCommand = "SI"
	DenonCommandCursorControl  DenonCommand = "MN"
	DenonCommandNS             DenonCommand = "NS"
	DenonCommandMS             DenonCommand = "MS"
	DenonCommandVS             DenonCommand = "VS"
	DenonCommandZone2          DenonCommand = "Z2"
	DenonCommandZone3          DenonCommand = "Z3"
)

const (
	MainZone DenonZone = "MAIN"
	Zone2    DenonZone = "Z2"
	Zone3    DenonZone = "Z3"
)

const (
	STATUS_URL           string = "/goform/formMainZone_MainZoneXmlStatus.xml"
	STATUS_Z2_URL        string = "/goform/formZone2_Zone2XmlStatus.xml"
	STATUS_Z3_URL        string = "/goform/formZone3_Zone3XmlStatus.xml"
	MAINZONE_URL         string = "/goform/formMainZone_MainZoneXml.xml"
	COMMAND_URL          string = "/goform/formiPhoneAppDirect.xml"
	NET_AUDIO_STATUR_URL string = "/goform/formNetAudio_StatusXml.xml"
)

type DenonXML struct {
	XMLName          xml.Name     `xml:"item"`
	FriendlyName     string       `xml:"FriendlyName>value"`
	Power            string       `xml:"Power>value"`
	ZonePower        string       `xml:"ZonePower>value"`
	RenameZone       string       `xml:"RenameZone>value"`
	TopMenuLink      string       `xml:"TopMenuLink>value"`
	VideoSelectDisp  string       `xml:"VideoSelectDisp>value"`
	VideoSelect      string       `xml:"VideoSelect>value"`
	VideoSelectOnOff string       `xml:"VideoSelectOnOff>value"`
	VideoSelectList  []ValueLists `xml:"VideoSelectLists>value"`
	ECOModeDisp      string       `xml:"ECOModeDisp>value"`
	ECOMode          string       `xml:"ECOMode>value"`
	ECOModeList      []ValueLists `xml:"ECOModeLists>value"`
	AddSourceDisplay string       `xml:"AddSourceDisplay>value"`
	ModelId          string       `xml:"ModelId>value"`
	BrandId          string       `xml:"BrandId>value"`
	SalesArea        string       `xml:"SalesArea>value"`
	InputFuncSelect  string       `xml:"InputFuncSelect>value"`
	NetFuncSelect    string       `xml:"NetFuncSelect>value"`
	SelectSurround   string       `xml:"selectSurround>value"`
	VolumeDisplay    string       `xml:"VolumeDisplay>value"`
	MasterVolume     string       `xml:"MasterVolume>value"`
	Mute             string       `xml:"Mute>value"`
}

type ValueLists struct {
	Index string `xml:"index,attr"`
	Table string `xml:"table,attr"`
}

type DenonAVR struct {
	Host string

	telnet *telnet.Conn

	mainZoneData DenonXML

	// Zone Status
	zoneStatus     map[DenonZone]DenonZoneStatus
	netAudioStatus DenonNetAudioStatus

	// Attributes
	attributes     map[string]interface{}
	attributeMutex sync.Mutex

	updateTrigger  chan string
	controlChannel chan string

	// Telnet
	telnetEnabled bool
	telnetEvents  chan *TelnetEvent
	telnetMutex   sync.Mutex

	entityChangedFunction map[string][]func(interface{})
}

func NewDenonAVR(host string, telnetEnabled bool) *DenonAVR {

	denonavr := DenonAVR{}

	denonavr.Host = host

	denonavr.mainZoneData = DenonXML{}
	denonavr.zoneStatus = make(map[DenonZone]DenonZoneStatus)
	denonavr.netAudioStatus = DenonNetAudioStatus{}

	denonavr.entityChangedFunction = make(map[string][]func(interface{}))

	denonavr.attributes = make(map[string]interface{})

	denonavr.updateTrigger = make(chan string)
	denonavr.controlChannel = make(chan string)
	denonavr.telnetEvents = make(chan *TelnetEvent)

	denonavr.telnetEnabled = telnetEnabled

	return &denonavr
}

// Add a new function that is called when a attribute of this entity has changed
func (d *DenonAVR) AddHandleEntityChangeFunc(attribute string, f func(interface{})) {
	d.entityChangedFunction[attribute] = append(d.entityChangedFunction[attribute], f)
}

// Call the registred entity change function with the new value for a attribute
func (d *DenonAVR) callEntityChangeFunction(attribute string, newValue interface{}) {
	if len(d.entityChangedFunction[attribute]) > 0 {
		for _, f := range d.entityChangedFunction[attribute] {
			go f(newValue)
		}
	}
}

func (d *DenonAVR) getMainZoneDataFromDevice() {

	d.mainZoneData = DenonXML{} // Somehow the values in the array are added instead of replaced. Not sure if this is the solution, but it works...
	resp, err := http.Get("http://" + d.Host + MAINZONE_URL)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Cannot read response body")
	}

	if err := xml.Unmarshal(body, &d.mainZoneData); err != nil {
		log.WithError(err).Info("Could not unmarshall")
	}
}

func (d *DenonAVR) StopListenLoop() {
	log.Info("Stop Denon Listen Loop")
	d.controlChannel <- "disconnect"
}

func (d *DenonAVR) StartListenLoop() error {

	log.Info("Start Denon Listen Loop")

	updateInterval := 5 * time.Second
	ticker := time.NewTicker(updateInterval)

	telnetControlChannel := make(chan string)

	defer func() {
		ticker.Stop()
		log.Debug("Denon Listen Loop stopped")
	}()

	// Start listening to telnet
	if d.telnetEnabled {
		go func() {
			// Try reconnecting to telnet if reconnect is true
			for {
				log.Debug("Starting Telnet Loop")
				if err, reconnect := d.listenTelnet(telnetControlChannel); err != nil {
					log.WithError(err).Debug("listenTelnet returned with error")
					close(telnetControlChannel)
					d.controlChannel <- "telnet_error"
					if !reconnect {
						break
					}
				}
			}
		}()
	}

	// do an intial update to make sure we have up to date values
	d.updateAndNotify()

	for {
		select {
		case <-d.updateTrigger:
			// force manual update
			d.updateAndNotify()
		case <-ticker.C:
			// Update every 5 Seconds
			d.updateAndNotify()
		case msg := <-d.controlChannel:
			switch msg {
			case "disconnect":
				log.Debug("return listen loop due to disconnect")
				// Make sure telnet also disconnects
				telnetControlChannel <- "disconnect"
				return nil
			case "telnet_error":
				log.Debug("return listen loop with error")
				return fmt.Errorf("telnet connection error")
			}

		}
	}
}

func (d *DenonAVR) updateAndNotify() {

	// Don't wait on each Call, handle them individually
	go d.updateMainZoneDataAndNotify()
	go d.updateZoneStatusAndNotify(MainZone)
	go d.updateZoneStatusAndNotify(Zone2)
	go d.updateZoneStatusAndNotify(Zone3)
}

func (d *DenonAVR) updateMainZoneDataAndNotify() {

	d.getMainZoneDataFromDevice()

	d.SetAttribute("POWER", d.mainZoneData.Power)

	d.getNetAudioStatus()

	// Media Title
	d.getMediaTitle()

	// Media Image URL
	d.getMediaImageURL()
}

func (d *DenonAVR) updateZoneStatusAndNotify(zone DenonZone) {

	// Get Data from Denon AVR
	zoneStatus := d.getZoneStatus(zone)
	zoneName := d.getZoneName(zone)

	d.SetAttribute(zoneName+"Power", zoneStatus.Power)
	d.SetAttribute(zoneName+"Volume", zoneStatus.MasterVolume)
	d.SetAttribute(zoneName+"Mute", zoneStatus.Mute)
	d.SetAttribute(zoneName+"SurroundMode", zoneStatus.SurrMode)

	// We use the renamed input sources
	inputFuncSelectList := d.GetZoneInputFuncList(zone)
	// map[string]string are unorderen and range gives a different result on each run
	inputList := make([]string, 0, len(inputFuncSelectList))
	for _, renamedSource := range inputFuncSelectList {
		inputList = append(inputList, renamedSource)
	}
	// sort the slice by keys
	sort.Strings(inputList)
	d.SetAttribute(zoneName+"InputFuncList", inputList)

	inputFuncSelect := zoneStatus.InputFuncSelect
	// Rename Source with the SOURCE_MAPPING if necessary
	for source, origin := range SOURCE_MAPPING {
		if origin == zoneStatus.InputFuncSelect {
			inputFuncSelect = source
			break
		}
	}
	// And then custom renames
	if inputFuncSelectList[inputFuncSelect] != "" {
		inputFuncSelect = inputFuncSelectList[inputFuncSelect]
	}
	d.SetAttribute(zoneName+"InputFuncSelect", inputFuncSelect)

}

func (d *DenonAVR) getZoneName(zone DenonZone) string {

	switch zone {
	case MainZone:
		return "MainZone"

	case Zone2:
		return "Zone2"

	case Zone3:
		return "Zone3"
	}

	return ""
}
