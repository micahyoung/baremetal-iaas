package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/micahyoung/baremetal-iaas/cmd"
	"github.com/micahyoung/baremetal-iaas/stemcell"
)

func main() {
	app := cli.NewApp()

	stemcellClient := stemcell.NewStemcellClient()
	diskClient := stemcell.NewDiskClient()
	cmdFactory := cmd.NewCommandFactory(stemcellClient, diskClient)

	app.Commands = []cli.Command{
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build",
			Action:  cmdFactory.BuildAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "stemcell, s",
					Usage: "Stemcell tar gz file",
				},
				cli.StringFlag{
					Name:  "agent-settings-file, a",
					Usage: "BOSH Agent Settings config JSON file",
				},
				cli.StringFlag{
					Name:  "disk-image-file, o",
					Usage: "Hard drive disk image file",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
