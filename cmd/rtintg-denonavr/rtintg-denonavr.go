package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/splattner/remotetwo-integration-denonavr/pkg/cmd"
	denonavr "github.com/splattner/remotetwo-integration-denonavr/pkg/cmd/denonavr"

	log "github.com/sirupsen/logrus"
)

const RELEASEDATE string = "31.10.2023"

func main() {

	baseName := filepath.Base(os.Args[0])

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("UC_")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Info("Unable to read config")
	}

	err := denonavr.NewCommand(baseName).Execute()
	cmd.CheckError(err)

}
