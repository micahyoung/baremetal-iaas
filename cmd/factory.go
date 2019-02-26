package cmd

import (
	"github.com/micahyoung/baremetal-iaas/configure"
	"github.com/micahyoung/baremetal-iaas/stemcell"
)

type commandFactory struct {
	stemcellClient  *stemcell.StemcellClient
	diskClient      *stemcell.DiskClient
	configGenerator *configure.Generator
}

func NewCommandFactory(stemcellClient *stemcell.StemcellClient, diskClient *stemcell.DiskClient, configGenerator *configure.Generator) *commandFactory {
	return &commandFactory{stemcellClient, diskClient, configGenerator}
}
