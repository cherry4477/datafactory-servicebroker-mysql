package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	conf "github.com/cloudfoundry-samples/go_service_broker/config"
	utils "github.com/cloudfoundry-samples/go_service_broker/utils"
	webs "github.com/cloudfoundry-samples/go_service_broker/web_server"
)

type Options struct {
	ConfigPath string
	Cloud      string
}

var options Options

func init() {
	defaultConfigPath := utils.GetPath([]string{"assets", "config.json"})
	flag.StringVar(&options.ConfigPath, "c", defaultConfigPath, "use '-c' option to specify the config file path")

	flag.StringVar(&options.Cloud, "cloud", utils.SQL, "use '--cloud' option to specify the cloud client to use: AWS or SoftLayer (SL)")

	flag.Parse()
}

func main() {
	err := checkCloudName(options.Cloud)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	_, err = conf.LoadConfig(options.ConfigPath)
	if err != nil {
		panic(fmt.Sprintf("Error loading config file [%s]...", err.Error()))
	}

	server, err := webs.CreateServer(options.Cloud)
	if err != nil {
		panic(fmt.Sprintf("Error creating server [%s]...", err.Error))
	}

	server.Start()
}

// Private func

func checkCloudName(name string) error {

	switch name {
	case utils.AWS, utils.SOFTLAYER, utils.SL, utils.SQL:
		return nil
	}

	return errors.New(fmt.Sprintf("Invalid cloud name: %s", name))
}
