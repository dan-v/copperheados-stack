package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/dan-v/unofficial-copperheados/chos"
)

var name, region, device, ami, sshKey, spotPrice string
var remove bool

var RootCmd = &cobra.Command{
	Use:   "chosdeploy",
	Short: "Deploy build and update environment for CopperheadOS",
	Args: func(cmd *cobra.Command, args []string) error {
		if device != "marlin" && device != "sailfish" {
			return errors.New("Must specify either marlin or sailfish for device type")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !remove {
			chos.AWSApply(
				chos.ChosConfig{
					Name:      name,
					Region:    region,
					Device:    device,
					AMI:       ami,
					SSHKey:    sshKey,
					SpotPrice: spotPrice,
				},
			)
		} else {
			chos.AWSDestroy(
				chos.ChosConfig{
					Name:   name,
					Region: region,
				},
			)
		}
	},
}

func init() {
	RootCmd.Flags().StringVarP(&name, "name", "n", "", "Name for build environment")
	RootCmd.MarkFlagRequired("name")
	RootCmd.Flags().StringVarP(&region, "region", "r", "", "Region for build environment")
	RootCmd.MarkFlagRequired("region")
	RootCmd.Flags().StringVarP(&device, "device", "d", "", "marlin|sailfish")
	RootCmd.MarkFlagRequired("device")
	RootCmd.Flags().StringVar(&sshKey, "ssh-key", "", "SSH key name to use")
	RootCmd.Flags().StringVar(&spotPrice, "spot-price", ".80", "Spot price to use")
	RootCmd.Flags().StringVar(&ami, "ami", "ami-0def3275", "AMI to use for build environment")
	RootCmd.Flags().BoolVar(&remove, "remove", false, "Remove environment")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
