package ucrt

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goucrtcmd "github.com/splattner/goucrt/pkg/cmd"
	"github.com/splattner/goucrt/pkg/integration"
	denonavrclient "github.com/splattner/remotetwo-integration-denonavr/pkg/clients/denonavr"
	"github.com/splattner/remotetwo-integration-denonavr/pkg/cmd"

	log "github.com/sirupsen/logrus"
)

func NewCommand(name string) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   name,
		Short: "Remote Two integration for Denon AV Receiver",
		Long:  `Denon AVR Integration for a Unfolded Circle Remote Two`,
		Run: func(c *cobra.Command, args []string) {

			log.SetOutput(os.Stdout)

			debug := viper.GetBool("debug")
			if debug {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.InfoLevel)
			}

			var config integration.Config
			if err := viper.Unmarshal(&config); err != nil {
				log.WithError(err).Error("Cannot unmarshal config with viper")
			}

			i, err := integration.NewIntegration(config)
			cmd.CheckError(err)

			myclient := denonavrclient.NewDenonAVRClient(i)

			myclient.InitClient()

			cmd.CheckError(i.Run())

		},
	}

	if err := goucrtcmd.BindStandardFlags(rootCmd); err != nil {
		log.WithError(err).Error("Cannot bind standard flags")
	}

	return rootCmd
}
