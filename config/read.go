package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Listening          string
	GeoIPCityFile      string
	GeoIPASNFile       string
	DataDir            string
	EnableRegistration bool
}

var (
	ActualConfig = Config{}
)

func ReadConfig(file string) Config {

	viper.AddConfigPath("/etc/k20a/")
	viper.AddConfigPath("$HOME/.k20a/")
	viper.AddConfigPath("$HOME/.config/k20a/")
	viper.AddConfigPath("data")
	viper.AddConfigPath(".")
	viper.SetConfigName(file)

	viper.SetDefault("Listening", ":3000")
	viper.SetDefault("GeoIPCityFile", "/var/lib/GeoIP/GeoLite2-City.mmdb")
	viper.SetDefault("GeoIPASNFile", "/var/lib/GeoIP/GeoLite2-ASN.mmdb")
	viper.SetDefault("DataDir", "data")
	viper.SetDefault("EnableRegistration", true)

	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
	ActualConfig.Listening = viper.GetString("Listening")
	ActualConfig.GeoIPCityFile = viper.GetString("GeoIPCityFile")
	ActualConfig.GeoIPASNFile = viper.GetString("GeoIPASNFile")
	ActualConfig.DataDir = viper.GetString("DataDir")
	ActualConfig.EnableRegistration = viper.GetBool("EnableRegistration")

	log.Printf("using config: %+v", ActualConfig)
	return ActualConfig
}
