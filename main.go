package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const cfgFileName = ".rootinha-bot.yaml"

var cfgFile string

func main() {

	r := New()

	cmd := newCmd(r)

	if err := cmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func newCmd(r *Rootinha) *cobra.Command {
	cobra.OnInitialize(func() {
		if cfgFile == "" {
			cfgFile = cfgFileName
		}

		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			logrus.WithError(err).Error("error reading the config file")
			return
		}

		logrus.WithField("filename", viper.ConfigFileUsed()).Info("using config file...", viper.ConfigFileUsed())
	})

	rootCmd := &cobra.Command{
		Use:   "rootinha-bot",
		Short: "",
		Long:  `.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mapstructure.Decode(viper.AllSettings()["bot"], r.Config)
			r.Config.Compile()

			return r.Start()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", cfgFileName))

	return rootCmd
}