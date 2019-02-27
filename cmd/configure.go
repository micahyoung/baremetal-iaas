package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/micahyoung/baremetal-iaas/settings"

	"github.com/urfave/cli"
)

var embeddedScriptName = `embed.ipxe`
var serverScriptName = `server.ipxe`
var settingsJSONName = `settings.json`

func (f *commandFactory) ConfigureAction(c *cli.Context) error {
	var err error
	stemcellTarballPath := c.String("stemcell")
	serverBaseURL := c.String("server-base-url")
	buildDirPath := c.String("build-directory")

	if stemcellTarballPath == "" {
		return fmt.Errorf("missing stemcell file")
	}
	if buildDirPath == "" {
		return fmt.Errorf("missing build dir")
	}
	if serverBaseURL == "" {
		return fmt.Errorf("missing server base URL")
	}

	if _, err = os.Stat(stemcellTarballPath); os.IsNotExist(err) {
		return fmt.Errorf("stemcell does not exist: %s", stemcellTarballPath)
	}
	if _, err = os.Stat(buildDirPath); os.IsNotExist(err) {
		return fmt.Errorf("build directory does not exist: %s", buildDirPath)
	}

	//TODO validate server IP format
	if _, err = os.Stat(buildDirPath); os.IsNotExist(err) {
		return fmt.Errorf("build directory does not exist: %s", buildDirPath)
	}

	settings := settings.NewSettings()

	clonedStemcellFilePath := filepath.Join(buildDirPath, settings.StemcellTarGzPath)
	kernelFilePath := filepath.Join(buildDirPath, settings.KernelPath)
	initRDFilePath := filepath.Join(buildDirPath, settings.InitRDPath)

	var clonedStemcellFile *os.File
	if clonedStemcellFile, err = os.Create(clonedStemcellFilePath); err != nil {
		return err
	}
	defer clonedStemcellFile.Close()

	var stemcellTarGzFile *os.File
	if stemcellTarGzFile, err = os.Open(stemcellTarballPath); err != nil {
		return err
	}
	defer stemcellTarGzFile.Close()

	if _, err = io.Copy(clonedStemcellFile, stemcellTarGzFile); err != nil {
		return err
	}

	var rootDiskTempFile *os.File
	if rootDiskTempFile, err = ioutil.TempFile("", "baremetal-root-disk.img"); err != nil {
		return err
	}
	defer func() {
		fmt.Printf("cleaned up root disk temp file") 
		rootDiskTempFile.Close()
		os.Remove(rootDiskTempFile.Name())
	}()

	clonedStemcellFile.Seek(0, io.SeekStart)
	err = f.stemcellClient.ExtractStemcellRootDisk(clonedStemcellFile, func(imageFileReader io.Reader) error {
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

	var kernelFile *os.File
	if kernelFile, err = os.Create(kernelFilePath); err != nil {
		return err
	}
	defer kernelFile.Close()

	var initRDFile *os.File
	if initRDFile, err = os.Create(initRDFilePath); err != nil {
		return err
	}
	defer initRDFile.Close()

	err = f.diskClient.ExtractRootDiskKernel(rootDiskTempFile, func(inodeReader io.Reader) error {
		var err error

		copiedBytes, err := io.Copy(kernelFile, inodeReader)
		if err != nil {
			return err
		}

		fmt.Printf("kernel: %s (%d bytes)\n", kernelFile.Name(), copiedBytes)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract kernel file from root disk: %s", kernelFile.Name())
		return err
	}

	err = f.diskClient.ExtractRootDiskInitRD(rootDiskTempFile, func(inodeReader io.Reader) error {
		var err error

		copiedBytes, err := io.Copy(initRDFile, inodeReader)
		if err != nil {
			return err
		}

		fmt.Printf("initrd.img: %s (%d bytes)\n", initRDFile.Name(), copiedBytes)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract initrd files from root disk: %s", initRDFile.Name())
		return err
	}

	var uuid string
	uuid, err = f.diskClient.VolumeUUID(rootDiskTempFile)
	if err != nil {
		return err
	}

	var embeddedScriptContent string
	if embeddedScriptContent, err = f.configGenerator.EmbeddedScriptContent(serverBaseURL, serverScriptName); err != nil {
		return err
	}

	var serverScriptContent string
	if serverScriptContent, err = f.configGenerator.ServerScriptContent(settings, uuid); err != nil {
		return err
	}

	embeddedScriptPath := filepath.Join(buildDirPath, embeddedScriptName)
	ioutil.WriteFile(embeddedScriptPath, []byte(embeddedScriptContent), 0666)

	serverScriptPath := filepath.Join(buildDirPath, serverScriptName)
	ioutil.WriteFile(serverScriptPath, []byte(serverScriptContent), 0666)

	return nil
}
