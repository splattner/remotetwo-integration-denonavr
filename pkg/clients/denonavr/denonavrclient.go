package denonavrclient

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/splattner/goucrt/pkg/entities"
	"github.com/splattner/goucrt/pkg/integration"
	"github.com/splattner/remotetwo-integration-denonavr/pkg/denonavr"
)

// Denon AVR Client Implementation
type DenonAVRClient struct {
	integration.Client
	denon *denonavr.DenonAVR

	moni1Button    *entities.ButtonEntity
	moni2Button    *entities.ButtonEntity
	moniAutoButton *entities.ButtonEntity

	mediaPlayer *entities.MediaPlayerEntity

	mapOnState map[bool]entities.MediaPlayerEntityState
}

func NewDenonAVRClient(i *integration.Integration) *DenonAVRClient {
	client := DenonAVRClient{}

	client.IntegrationDriver = i
	// Start without a connection
	client.DeviceState = integration.DisconnectedDeviceState

	client.Messages = make(chan string)

	inputSetting_ipaddr := integration.SetupDataSchemaSettings{
		Id: "ipaddr",
		Label: integration.LanguageText{
			En: "IP Address of your Denon Receiver",
		},
		Field: integration.SettingTypeText{
			Text: integration.SettingTypeTextDefinition{
				Value: "",
			},
		},
	}

	inputSetting_telnet := integration.SetupDataSchemaSettings{
		Id: "telnet",
		Label: integration.LanguageText{
			En: "Use telnet to communicate with your DenonAVR Device",
		},
		Field: integration.SettingTypeCheckbox{
			Checkbox: integration.SettingTypeCheckboxDefinition{
				Value: false,
			},
		},
	}

	metadata := integration.DriverMetadata{
		DriverId: "denonavr",
		Developer: integration.Developer{
			Name: "Sebastian Plattner",
		},
		Name: integration.LanguageText{
			En: "Denon AVR",
		},
		Version: "0.2.8",
		SetupDataSchema: integration.SetupDataSchema{
			Title: integration.LanguageText{
				En: "Configuration",
				De: "Konfiguration",
			},
			Settings: []integration.SetupDataSchemaSettings{inputSetting_ipaddr, inputSetting_telnet},
		},
		Icon: "custom:denon.png",
	}

	client.IntegrationDriver.SetMetadata(&metadata)

	// set the client specific functions
	client.InitFunc = client.initDenonAVRClient
	client.SetupFunc = client.denonHandleSetup
	client.ClientLoopFunc = client.denonClientLoop

	client.mapOnState = map[bool]entities.MediaPlayerEntityState{
		true:  entities.OnMediaPlayerEntityState,
		false: entities.OffMediaPlayerEntityState,
	}

	return &client
}

func (c *DenonAVRClient) initDenonAVRClient() {

	log.Debug("Initialize DenonAVR CLient")

	// Media Player
	c.mediaPlayer = entities.NewMediaPlayerEntity("mediaplayer", entities.LanguageText{En: "Denon AVR"}, "", entities.ReceiverMediaPlayerDeviceClass)
	c.mediaPlayer.AddFeature(entities.OnOffMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.ToggleMediaPlayerEntityyFeatures)
	c.mediaPlayer.AddFeature(entities.VolumeMediaPlayerEntityyFeatures)
	c.mediaPlayer.AddFeature(entities.VolumeUpDownMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.MuteMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.UnmuteMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.MuteToggleMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.SelectSourceMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.SelectSoundModeMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.DPadMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.MediaTitleMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.MediaImageUrlMediaPlayerEntityFeatures)
	c.mediaPlayer.AddFeature(entities.MenuMediaPlayerEntityFeatures)

	if err := c.IntegrationDriver.AddEntity(c.mediaPlayer); err != nil {
		log.WithError(err).Error("Cannot add Entity")
	}

	// Butons
	c.moni1Button = entities.NewButtonEntity("moni1", entities.LanguageText{En: "Monitor Out 1"}, "")
	if err := c.IntegrationDriver.AddEntity(c.moni1Button); err != nil {
		log.WithError(err).Error("Cannot add Entity")
	}

	c.moni2Button = entities.NewButtonEntity("moni2", entities.LanguageText{En: "Monitor Out 2"}, "")
	if err := c.IntegrationDriver.AddEntity(c.moni2Button); err != nil {
		log.WithError(err).Error("Cannot add Entity")
	}

	c.moniAutoButton = entities.NewButtonEntity("moniauto", entities.LanguageText{En: "Monitor Out Auto"}, "")
	if err := c.IntegrationDriver.AddEntity(c.moniAutoButton); err != nil {
		log.WithError(err).Error("Cannot add Entity")
	}

}

func (c *DenonAVRClient) denonHandleSetup(setup_data integration.SetupData) {

	c.IntegrationDriver.SetDriverSetupState(integration.SetupEvent, integration.SetupState, "", nil)
	time.Sleep(1 * time.Second)

	// No required User action so finish
	c.IntegrationDriver.SetDriverSetupState(integration.StopEvent, integration.OkState, "", nil)

}

func (c *DenonAVRClient) setupDenon() {
	if c.IntegrationDriver.SetupData != nil && c.IntegrationDriver.SetupData["ipaddr"] != "" {
		telnetEnabled, err := strconv.ParseBool(c.IntegrationDriver.SetupData["telnet"])
		if err != nil {
			telnetEnabled = false
		}
		c.denon = denonavr.NewDenonAVR(c.IntegrationDriver.SetupData["ipaddr"], telnetEnabled)
	} else {
		log.Error("Cannot setup Denon, missing setupData")
	}
}

func (c *DenonAVRClient) configureDenon() {

	// Configure the Entity Change Func

	// Buttons
	c.moni1Button.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoni1Out)
	c.moni2Button.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoni2Out)
	c.moniAutoButton.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoniAutoOut)

	// Media Player
	c.denon.AddHandleEntityChangeFunc("MainZonePower", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.StateMediaPlayerEntityAttribute, c.mapOnState[c.denon.IsOn()])
	})

	c.denon.AddHandleEntityChangeFunc("MainZoneVolume", func(value interface{}) {

		var volume float64
		if s, err := strconv.ParseFloat(value.(string), 64); err == nil {
			volume = s
		}

		c.mediaPlayer.SetAttribute(entities.VolumeMediaPlayerEntityAttribute, volume+80)
	})

	c.denon.AddHandleEntityChangeFunc("MainZoneMute", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.MutedMediaPlayeEntityAttribute, c.denon.MainZoneMuted())
	})

	c.denon.AddHandleEntityChangeFunc("MainZoneInputFuncList", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.SourceListMediaPlayerEntityAttribute, value.([]string))
	})

	c.denon.AddHandleEntityChangeFunc("MainZoneInputFuncSelect", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.SourceMediaPlayerEntityAttribute, value.(string))
	})

	c.denon.AddHandleEntityChangeFunc("MainZoneSurroundMode", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.SoundModeMediaPlayerEntityAttribute, value.(string))
	})

	// We can set the sound_mode_list without change handler. Its static
	func() {
		c.mediaPlayer.SetAttribute(entities.SoundModeListMediaPlayerEntityAttribute, c.denon.GetSoundModeList())
	}()

	// Media Title
	c.denon.AddHandleEntityChangeFunc("media_title", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.MediaTitleMediaPlayerEntityAttribute, value.(string))
	})

	// Media Image URL
	c.denon.AddHandleEntityChangeFunc("media_image_url", func(value interface{}) {
		c.mediaPlayer.SetAttribute(entities.MediaImageUrlMediaPlayerEntityAttribute, value.(string))
	})

	// Add Commands
	c.mediaPlayer.MapCommand(entities.OnMediaPlayerEntityCommand, c.denon.TurnOn)
	c.mediaPlayer.MapCommand(entities.OffMediaPlayerEntityCommand, c.denon.TurnOff)
	c.mediaPlayer.MapCommand(entities.ToggleMediaPlayerEntityCommand, c.denon.TogglePower)

	c.mediaPlayer.AddCommand(entities.VolumeMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("VolumeMediaPlayerEntityCommand called")

		var volume float64
		if v, err := strconv.ParseFloat(params["volume"].(string), 64); err == nil {
			volume = v
		}
		if err := c.denon.SetVolume(volume); err != nil {
			return 404
		}
		return 200
	})

	// Volume commands
	c.mediaPlayer.MapCommand(entities.VolumeUpMediaPlayerEntityCommand, c.denon.SetVolumeUp)
	c.mediaPlayer.MapCommand(entities.VolumeDownMediaPlayerEntityCommand, c.denon.SetVolumeDown)
	c.mediaPlayer.MapCommand(entities.MuteMediaPlayerEntityCommand, c.denon.MainZoneMute)
	c.mediaPlayer.MapCommand(entities.UnmuteMediaPlayerEntityCommand, c.denon.MainZoneUnMute)
	c.mediaPlayer.MapCommand(entities.MuteToggleMediaPlayerEntityCommand, c.denon.MainZoneMuteToggle)

	// Source commands
	c.mediaPlayer.AddCommand(entities.SelectSourcMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("SelectSourcMediaPlayerEntityCommand called")
		if params["source"] != nil {
			return c.denon.SetSelectSourceMainZone(params["source"].(string))
		}
		return 200
	})

	// Cursor commands
	c.mediaPlayer.AddCommand(entities.CursorUpMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("CursorUpMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlUp)
	})
	c.mediaPlayer.AddCommand(entities.CursorDownMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("CursorDownMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlDown)
	})
	c.mediaPlayer.AddCommand(entities.CursorLeftMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("CursorUpMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlLeft)
	})
	c.mediaPlayer.AddCommand(entities.CursorRightMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("CursorRightMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlRight)
	})
	c.mediaPlayer.AddCommand(entities.CursorEnterMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("CursorEnterMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlEnter)
	})
	c.mediaPlayer.AddCommand(entities.BackMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("BackMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlReturn)
	})
	c.mediaPlayer.AddCommand(entities.MenuMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("MenuMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlMenu)
	})

	// Sound Mode
	c.mediaPlayer.AddCommand(entities.SelectSoundModeMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("SelectSoundModeMediaPlayerEntityCommand called")
		return c.denon.SetSoundModeMainZone(params["mode"].(string))
	})

}

func (c *DenonAVRClient) denonClientLoop() {

	defer func() {
		c.SetDeviceState(integration.DisconnectedDeviceState)
	}()

	if c.denon == nil {
		// Initialize Denon Client
		c.setupDenon()
	} else {
		return
	}

	// Start the Denon Liste Loop if already configured
	if c.denon != nil {

		// Configure Denon Client
		c.configureDenon()

		log.WithFields(log.Fields{
			"Denon IP": c.denon.Host}).Info("Start Denon AVR Client Loop")
		go c.denon.StartListenLoop()

		// Handle connection to device this integration shall control
		// Set Device state to connected when connection is established
		c.SetDeviceState(integration.ConnectedDeviceState)
	}

	// Run Client Loop to handle entity changes from device
	for {
		msg := <-c.Messages
		switch msg {
		case "disconnect":
			c.denon.StopListenLoop()
			return
		}
	}

}
