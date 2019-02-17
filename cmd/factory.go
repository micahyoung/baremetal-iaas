package cmd

import (
	"github.com/micahyoung/baremetal-iaas/stemcell"
)

type commandFactory struct {
	stemcellClient *stemcell.StemcellClient
	diskClient     *stemcell.DiskClient
}

func NewCommandFactory(stemcellClient *stemcell.StemcellClient, diskClient *stemcell.DiskClient) *commandFactory {
	return &commandFactory{stemcellClient, diskClient}
}
