package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	keywords *[]string
)

func init() {
	// init environment variables
	youtubeApiKey := "YOUTUBE_API_KEY"
	if err := viper.BindEnv(youtubeApiKey); err != nil || viper.Get(youtubeApiKey) == nil {
		if err == nil {
			err = fmt.Errorf("Fatal error missing environment variable: %s.\nPlease reference to README.md Environment Variable section.", youtubeApiKey)
		}
		panic(err)
	}

	// init config file
	viper.SetConfigFile("config.toml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s.\nPlease reference to README.md Config File section.", err))
	}

	// init command line flags
	keywords = pflag.StringArray("keywords", nil, "")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil || keywords == nil || len(*keywords) == 0 {
		if err == nil {
			err = fmt.Errorf("Missing keywords flag.\n Usage: go-youtube-crawler --keywords term1 --keywords term2 ... ")
		}
		panic(err)
	}

	// init logging
	zerolog.TimeFieldFormat = ""
	log.Debug().Interface("all", viper.AllSettings()).Msg("Show all config settings")
}

func main() {

}
