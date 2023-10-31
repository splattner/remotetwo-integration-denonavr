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

func (d *DenonAVR) handleTelnetEvents() {

	for {
		event := <-d.telnetEvents
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
	}
}

func (d *DenonAVR) listenTelnet() {

	defer func() {
		log.Debug("Closing Telnet connection")
		d.telnet.Close()
	}()

	go d.handleTelnetEvents()

	var err error

	for {
		d.telnet, err = telnet.DialTimeout("tcp", d.Host+":23", 5*time.Second)
		if err != nil {
			log.WithError(err).Info("failed to connect to telnet")
			continue
		}

		if err = d.telnet.Conn.(*net.TCPConn).SetKeepAlive(true); err != nil {
			log.WithError(err).Error("failed to enable tcp keep alive")
		}

		if err = d.telnet.Conn.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second); err != nil {
			log.WithError(err).Error("failed to set tcp keep alive period")
		}

		log.WithField("host", d.Host+":23").Debug("Telnet connected")

		for {
			data, err := d.telnet.ReadString('\r')
			if err != nil {
				log.WithError(err).Errorf("failed to read form telnet")
				break
			}
			data = strings.Trim(data, " \n\r")

			parsedData := strings.Split(data, " ")
			event := TelnetEvent{}
			event.RawData = data
			event.Command = parsedData[0]
			if len(parsedData) > 1 {
				event.Payload = parsedData[1]
			}

			// Fire Event for handling
			d.telnetEvents <- &event
		}
	}
}

func (d *DenonAVR) sendTelnetCommand(cmd DenonCommand, payload string) error {

	d.telnetMutex.Lock()
	defer d.telnetMutex.Unlock()

	log.WithFields(log.Fields{
		"cmd":     string(cmd),
		"payload": payload,
	}).Debug("Send Telnet command")

	_, err := d.telnet.Write([]byte(string(cmd) + payload + "\r"))

	return err
}
