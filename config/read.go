package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config contains the configuration options
type Config struct {
	Listening          string
	GeoIPCityFile      string
	GeoIPASNFile       string
	DataDir            string
	EnableRegistration bool
	UseBundledWebApp   bool
}

var (
	// ActualConfig stores the last readed config value
	ActualConfig = Config{}
	file         = "rightana"
)

// ReadConfig reads the config file from the default locations
func ReadConfig() Config {

	viper.AddConfigPath("/etc/rightana/")
	viper.AddConfigPath("$HOME/.rightana/")
	viper.AddConfigPath("$HOME/.config/rightana/")
	viper.AddConfigPath("data")
	viper.AddConfigPath(".")
	viper.SetConfigName(file)

	viper.SetDefault("Listening", ":3000")
	viper.SetDefault("GeoIPCityFile", "/var/lib/GeoIP/GeoLite2-City.mmdb")
	viper.SetDefault("GeoIPASNFile", "/var/lib/GeoIP/GeoLite2-ASN.mmdb")
	viper.SetDefault("DataDir", "data")
	viper.SetDefault("EnableRegistration", true)
	viper.SetDefault("UseBundledWebApp", true)

	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
	ActualConfig.Listening = viper.GetString("Listening")
	ActualConfig.GeoIPCityFile = viper.GetString("GeoIPCityFile")
	ActualConfig.GeoIPASNFile = viper.GetString("GeoIPASNFile")
	ActualConfig.DataDir = viper.GetString("DataDir")
	ActualConfig.EnableRegistration = viper.GetBool("EnableRegistration")
	ActualConfig.UseBundledWebApp = viper.GetBool("UseBundledWebApp")

	log.Printf("using config: %+v", ActualConfig)
	return ActualConfig
}
