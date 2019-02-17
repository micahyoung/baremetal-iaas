package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
)

func (f *commandFactory) BuildAction(c *cli.Context) error {
	var err error
	stemcellTarballPath := c.String("stemcell")
	agentSettingsPath := c.String("agent-settings-file")
	outputImagePath := c.String("disk-image-file")

	if stemcellTarballPath == "" {
		return fmt.Errorf("missing stemcell file")
	}
	if agentSettingsPath == "" {
		return fmt.Errorf("missing agent settings file")
	}

	if outputImagePath == "" {
		return errors.New("missing desired output image file")
	}

	if _, err = os.Stat(stemcellTarballPath); os.IsNotExist(err) {
		return fmt.Errorf("stemcell does not exist: %s", stemcellTarballPath)
	}
	if _, err = os.Stat(agentSettingsPath); os.IsNotExist(err) {
		return fmt.Errorf("agent settings file does not exist: %s", agentSettingsPath)
	}
	if _, err = os.Stat(outputImagePath); !os.IsNotExist(err) {
		return fmt.Errorf("refusing to overwrite existing file: %s", outputImagePath)
	}

	var kernelTempFile *os.File
	if kernelTempFile, err = ioutil.TempFile("", "baremetal-vmlinuz"); err != nil {
		return err
	}
	defer os.RemoveAll(kernelTempFile.Name())

	var initRDTempFile *os.File
	if initRDTempFile, err = ioutil.TempFile("", "baremetal-initrd.img"); err != nil {
		return err
	}
	defer os.RemoveAll(initRDTempFile.Name())

	var rootDiskTempFile *os.File
	if rootDiskTempFile, err = ioutil.TempFile("", "baremetal-disk.img"); err != nil {
		return err
	}
	defer func() {
		rootDiskTempFile.Close()
		os.RemoveAll(rootDiskTempFile.Name()) //FIXME not working
	}()

	if err = f.stemcellClient.ExtractStemcellRootDisk(stemcellTarballPath, rootDiskTempFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract root disk from stemcell: %s", stemcellTarballPath)
		return err
	}

	rootDiskTempPath := rootDiskTempFile.Name()
	fmt.Printf("root disk: %s\n", rootDiskTempPath)
	if err = f.diskClient.ExtractRootDiskBootFiles(rootDiskTempPath, kernelTempFile, initRDTempFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract boot files from root disk: %s", rootDiskTempPath)
		return err
	}

	return nil
}
