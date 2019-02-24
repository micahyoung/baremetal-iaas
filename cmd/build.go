package cmd

import (
	"errors"
	"fmt"
	"io"
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
	defer func() {
		kernelTempFile.Close()
		os.RemoveAll(kernelTempFile.Name())
	}()

	var initRDTempFile *os.File
	if initRDTempFile, err = ioutil.TempFile("", "baremetal-initrd.img"); err != nil {
		return err
	}
	defer func() {
		initRDTempFile.Close()
		os.RemoveAll(initRDTempFile.Name())
	}()

	var rootDiskTempFile *os.File
	if rootDiskTempFile, err = ioutil.TempFile("", "baremetal-disk.img"); err != nil {
		return err
	}
	defer func() {
		rootDiskTempFile.Close()
		os.RemoveAll(rootDiskTempFile.Name())
	}()

	err = f.stemcellClient.ExtractStemcellRootDisk(stemcellTarballPath, func(imageFileReader io.Reader) error {
		var err error

		copiedBytes, err := io.Copy(rootDiskTempFile, imageFileReader)
		if err != nil {
			return err
		}

		fmt.Printf("root disk: %s (%d bytes)\n", rootDiskTempFile.Name(), copiedBytes)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract root disk from stemcell: %s", stemcellTarballPath)
		return err
	}

	err = f.diskClient.ExtractRootDiskKernel(rootDiskTempFile, func(inodeReader io.Reader) error {
		copiedBytes, err := io.Copy(kernelTempFile, inodeReader)
		if err != nil {
			return err
		}

		fmt.Printf("kernel: %s (%d bytes)\n", kernelTempFile.Name(), copiedBytes)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract kernel file from root disk: %s", kernelTempFile.Name())
		return err
	}

	err = f.diskClient.ExtractRootDiskInitRD(rootDiskTempFile, func(inodeReader io.Reader) error {
		copiedBytes, err := io.Copy(initRDTempFile, inodeReader)
		if err != nil {
			return err
		}

		fmt.Printf("initrd.img: %s (%d bytes)\n", initRDTempFile.Name(), copiedBytes)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract initrd files from root disk: %s", initRDTempFile.Name())
		return err
	}

	return nil
}
