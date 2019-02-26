package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/micahyoung/baremetal-iaas/cmd"
	"github.com/micahyoung/baremetal-iaas/configure"
	"github.com/micahyoung/baremetal-iaas/stemcell"
)

func main() {
	app := cli.NewApp()

	stemcellClient := stemcell.NewStemcellClient()
	diskClient := stemcell.NewDiskClient()
	configGenerator := &configure.Generator{}
	cmdFactory := cmd.NewCommandFactory(stemcellClient, diskClient, configGenerator)

	app.Commands = []cli.Command{
		{
			Name:    "configure",
			Aliases: []string{"c"},
			Usage:   "configure",
			Action:  cmdFactory.ConfigureAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "stemcell, s",
					Usage: "Stemcell tar gz file",
				},
				cli.StringFlag{
					Name:  "server-base-url, u",
					Usage: "Server Base URL where build directory files are hosted",
				},

				cli.StringFlag{
					Name:  "build-directory, d",
					Usage: "Directory containing vmlinuz, initrd.img, root-disk.img and agent-settings.json",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
