/*
Copyright © 2020 luca.capra@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"
	"os"

	"github.com/muka/camd/device"
	"github.com/muka/camd/onvif"
	"github.com/muka/camd/video"
	"github.com/spf13/cobra"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover camera sources",
	Long:  `This command search for camera sources on the local network(s).`,
	Run: func(cmd *cobra.Command, args []string) {

		emitter := make(chan device.Device)

		go func() {
			for {
				select {
				case dev := <-emitter:
					log.Printf("Received device %+v", dev)
				}
			}
		}()

		devices, err := video.ListDevices()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		for _, device := range devices {
			emitter <- device
		}

		err = onvif.Discover(emitter)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// discoverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// discoverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
