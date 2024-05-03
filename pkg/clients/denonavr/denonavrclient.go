package denonavrclient

import (
	"fmt"
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
		Version: "0.2.11",
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
	client.SetDriverUserDataFunc = client.handleSetDriverUserData

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
	c.mediaPlayer.AddFeature(entities.InfoPlayerEntityFeatures)

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

	simpleCommands := []string{"OUTPUT_MONITOR1", "OUTPUT_MONITOR2", "OUTPUT_MONITORAUTO"}

	c.mediaPlayer.AddOption(entities.SimpleCommandsMediaPlayerEntityOption, simpleCommands)

}

func (c *DenonAVRClient) denonHandleSetup(setup_data integration.SetupData) {

	c.IntegrationDriver.SetDriverSetupState(integration.SetupEvent, integration.SetupState, "", nil)

	telnetEnabled, err := strconv.ParseBool(c.IntegrationDriver.SetupData["telnet"])
	if err != nil {
		telnetEnabled = false
	}

	if telnetEnabled {
		var userAction = integration.RequireUserAction{
			Confirmation: integration.ConfirmationPage{
				Title: integration.LanguageText{
					En: "Be aware",
				},
				Message1: integration.LanguageText{
					En: "When using telnet, no other connection to your Denon AVR can be made via telnet",
				},
			},
		}

		// Start the setup with some require user data
		time.Sleep(1 * time.Second)
		c.IntegrationDriver.SetDriverSetupState(integration.SetupEvent, integration.WaitUserActionState, "", &userAction)
	} else {
		// No required User action so finish
		time.Sleep(1 * time.Second)
		c.IntegrationDriver.SetDriverSetupState(integration.StopEvent, integration.OkState, "", nil)
		c.FinishIntegrationSetup()
	}

}

func (c *DenonAVRClient) handleSetDriverUserData(user_data map[string]string, confirm bool) {

	log.Debug("Denon handle set driver user data")

	// confirm seems to be set to false always, maybe just the presence of the field tells me,
	// confirmation was sent?
	if len(user_data) == 0 {
		log.Debug("Telnet enabled, test if we can connect via telnet")

		telnetEnabled, err := strconv.ParseBool(c.IntegrationDriver.SetupData["telnet"])
		if err != nil {
			telnetEnabled = false
		}

		if telnetEnabled {
			denon := denonavr.NewDenonAVR(c.IntegrationDriver.SetupData["ipaddr"], telnetEnabled)
			if telnet, err := denon.ConnectTelnet(); err != nil {
				c.IntegrationDriver.SetDriverSetupState(integration.StopEvent, integration.ErrorState, integration.ConnectionRefusedError, nil)
				return
			} else {
				// Close connection again. So later we don't run into connection errors
				if err := telnet.Close(); err != nil {
					log.WithError(err).Debug("Telnet connection (already) closed")
				}
			}
		}

		c.IntegrationDriver.SetDriverSetupState(integration.StopEvent, integration.OkState, "", nil)
		c.FinishIntegrationSetup()

	}
}

func (c *DenonAVRClient) setupDenon() error {
	log.Debug("Create new Denon Client")
	if c.IntegrationDriver.SetupData != nil && c.IntegrationDriver.SetupData["ipaddr"] != "" {
		telnetEnabled, err := strconv.ParseBool(c.IntegrationDriver.SetupData["telnet"])
		if err != nil {
			telnetEnabled = false
		}
		c.denon = denonavr.NewDenonAVR(c.IntegrationDriver.SetupData["ipaddr"], telnetEnabled)
	} else {
		err := fmt.Errorf("cannot setup Denon Client, missing setupData")
		return err
	}

	return nil
}

func (c *DenonAVRClient) configureDenon() {

	log.Debug("Configure Denon Integration")

	// Configure the Entity Change Func

	// Buttons
	c.moni1Button.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoni1Out)
	c.moni2Button.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoni2Out)
	c.moniAutoButton.MapCommand(entities.PushButtonEntityCommand, c.denon.SetMoniAutoOut)

	c.mediaPlayer.MapCommand(entities.MediaPlayerEntityCommand("OUTPUT_MONITOR1"), c.denon.SetMoni1Out)
	c.mediaPlayer.MapCommand(entities.MediaPlayerEntityCommand("OUTPUT_MONITOR2"), c.denon.SetMoni2Out)
	c.mediaPlayer.MapCommand(entities.MediaPlayerEntityCommand("OUTPUT_MONITORAUTO"), c.denon.SetMoniAutoOut)

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
	c.mediaPlayer.AddCommand(entities.InfoMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("InfoMediaPlayerEntityCommand called")
		return c.denon.CursorControl(denonavr.DenonCursorControlMenuInfo)
	})

	// Sound Mode
	c.mediaPlayer.AddCommand(entities.SelectSoundModeMediaPlayerEntityCommand, func(mediaPlayer entities.MediaPlayerEntity, params map[string]interface{}) int {
		log.WithField("entityId", mediaPlayer.Id).Debug("SelectSoundModeMediaPlayerEntityCommand called")
		return c.denon.SetSoundModeMainZone(params["mode"].(string))
	})

}

func (c *DenonAVRClient) denonClientLoop() {
	log.Debug("Start Denon Client Loop")

	if !c.IntegrationSetupFinished() {

		// Migration, if ipaddr already available, call FinishIntegrationSetup
		if c.IntegrationDriver.SetupData != nil && c.IntegrationDriver.SetupData["ipaddr"] != "" {
			c.FinishIntegrationSetup()
		} else {
			log.Info("Cannot handle connect, integration setup not yet finished")
			return
		}
	}

	if c.denon == nil {
		// Initialize Denon Client
		if err := c.setupDenon(); err != nil {
			log.WithError(err).Error("Setup/Connection of Denon failed")
			return
		}
	}

	// Start the Denon Liste Loop if already configured
	if c.denon != nil {

		// Configure Denon Client
		c.configureDenon()

		go func() {
			log.WithFields(log.Fields{
				"Denon IP": c.denon.Host}).Info("Start Denon AVR Client Loop")
			if err := c.denon.StartListenLoop(); err != nil {
				log.WithError(err).Error("Denon AVR Client Loop ended with errors")
				c.Messages <- "error"
			}
		}()

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
			c.denon = nil
			c.SetDeviceState(integration.DisconnectedDeviceState)
			return
		case "error":
			// The denon Listen loop ended with some errors
			c.denon = nil
			c.SetDeviceState(integration.ErrorDeviceState)
		}
	}

}
