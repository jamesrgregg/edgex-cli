// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/edgexfoundry-holding/edgex-cli/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type DeviceConfig struct {
	Name            string
	Profile         string
	Description     string
	Service         string
	Labels          []string
	AddressableName string
	Protocols       map[string]models.ProtocolProperties
	AutoEvents      []models.AutoEvent
}

type DeviceFile struct {
	DeviceList []DeviceConfig
}

func NewCommand() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add devices",
		Long:  `Create the devices described in the given TOML file.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fname := cmd.Flag("file").Value.String()
			return processFile(fname)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Toml file containing device(s) configuration (required)")
	cmd.MarkFlagRequired("file")
	return cmd
}

func addDevice(dev models.Device) (string, error) {
	url := config.Conf.Clients["Metadata"].Url()
	mdc := metadata.NewDeviceClient(
		local.New(url + clients.ApiDeviceRoute),
	)
	resp, err := mdc.Add(context.Background(), &dev)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func processFile(fname string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Invalid TOML")
		}
	}()

	var content = &DeviceFile{}
	file, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}

	err = toml.Unmarshal(file, content)
	if err != nil {
		return
	}

	for _, d := range content.DeviceList {
		millis := time.Now().UnixNano() / int64(time.Millisecond)
		dev := models.Device{
			Name:           d.Name,
			Profile:        models.DeviceProfile{Name: d.Profile},
			Protocols:      d.Protocols,
			Labels:         d.Labels,
			Service:        models.DeviceService{Name: d.Service, Addressable: models.Addressable{Name: d.AddressableName}},
			AdminState:     models.Unlocked,
			OperatingState: models.Enabled,
			AutoEvents:     d.AutoEvents,
		}
		dev.Origin = millis
		id, err := addDevice(dev)
		if err != nil {
			fmt.Println("Error: ", err.Error())
		} else {
			fmt.Println("Created with ID: ", id)
		}
	}
	return nil
}
