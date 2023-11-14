package denonavr

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/ziutek/telnet"
)

type TelnetEvent struct {
	RawData string
	Command string
	Payload string
}

func (d *DenonAVR) handleTelnetEvents(controlChannel chan string) {

	log.Debug("Start Telnet Event Handler")
	for {
		select {
		case event := <-d.telnetEvents:
			parsedCommand := strings.Split(event.Command, "")
			command := parsedCommand[0] + parsedCommand[1]
			param := strings.Join(parsedCommand[2:], "")

			if event.Command == "OPSTS" {
				// ignore this
				continue
			}

			log.WithFields(log.Fields{
				"cmd":     event.Command,
				"payload": event.Payload,
				"command": DenonCommand(command),
				"param":   param,
			}).Debug("Telnet Event received")

			switch DenonCommand(command) {
			case DenonCommandPower:
				d.SetAttribute("POWER", param)
			case DennonCommandZoneMain:
				d.SetAttribute("MainZonePower", param)
			case DenonCommandMainZoneVolume:
				if param != "MAX" {

					volume, err := strconv.ParseFloat(param, 32)
					if err != nil {
						log.WithError(err).Error("failed to parse volume")
					}

					// The Volume command need the following
					// 10.5 -> MV105
					// 11 -> MV11
					if len(param) == 3 {
						volume = volume / 10
						log.WithField("volume", volume).Debug("Got volume after conversion")
					}

					d.SetAttribute("MainZoneVolume", fmt.Sprintf("%0.1f", volume-80))
				}

			case DenonCommandMainZoneMute:
				d.SetAttribute("MainZoneMute", strings.ToLower(param))
			}
		case msg := <-controlChannel:
			if msg == "disconnect" {
				return
			}
		}
	}
}

func (d *DenonAVR) listenTelnet(controlChannel chan string) (error, bool) {

	log.Debug("Start Telnet listen loop")

	eventHandlerControlChannel := make(chan string)

	defer func() {
		log.Debug("Closing Telnet connection")
		// Make sure eventHandler loop is also closed
		eventHandlerControlChannel <- "disconnect"
		if d.telnet != nil {
			if err := d.telnet.Close(); err != nil {
				log.WithError(err).Debug("Telnet connection (already) closed")
			}
		}
	}()

	go d.handleTelnetEvents(eventHandlerControlChannel)

	var err error

	d.telnet, err = telnet.DialTimeout("tcp", d.Host+":23", 5*time.Second)
	if err != nil {
		log.WithError(err).Info("failed to connect to telnet")
		return err, false
	}

	if err = d.telnet.Conn.(*net.TCPConn).SetKeepAlive(true); err != nil {
		log.WithError(err).Error("failed to enable tcp keep alive")
		return err, false
	}

	if err = d.telnet.Conn.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second); err != nil {
		log.WithError(err).Error("failed to set tcp keep alive period")
		return err, false
	}

	log.WithField("host", d.Host+":23").Debug("Telnet connected")

	dataChannel := make(chan string)

	for {

		go d.telnetReadString(dataChannel)
		select {
		case data := <-dataChannel:
			if data == "" {
				log.Debug("No Data from Telnet received")
				// Return the error but try to reconnect as we just lost the connection
				return fmt.Errorf("failed to read form telnet"), true
			}
			parsedData := strings.Split(data, " ")
			event := TelnetEvent{}
			event.RawData = data
			event.Command = parsedData[0]
			if len(parsedData) > 1 {
				event.Payload = parsedData[1]
			}

			// Fire Event for handling
			d.telnetEvents <- &event
		case msg := <-controlChannel:
			if msg == "disconnect" {
				return nil, false
			}
		}

	}
}

func (d *DenonAVR) telnetReadString(dataChannel chan string) {

	data, err := d.telnet.ReadString('\r')
	data = strings.Trim(data, " \n\r")
	if err != nil {
		log.WithError(err).Errorf("failed to read form telnet")
		close(dataChannel)
		return
	}
	dataChannel <- data
}

func (d *DenonAVR) sendTelnetCommand(cmd DenonCommand, payload string) error {

	d.telnetMutex.Lock()
	defer d.telnetMutex.Unlock()

	log.WithFields(log.Fields{
		"cmd":     string(cmd),
		"payload": payload,
	}).Debug("Send Telnet command")

	if d.telnet != nil {
		_, err := d.telnet.Write([]byte(string(cmd) + payload + "\r"))
		return err
	}

	return fmt.Errorf("cannot send telnet command, no telnet connection available")
}
