package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/diskfs/go-diskfs/disk"

	"github.com/micahyoung/baremetal-iaas/stemcell"
)

var stemcellPath = "/stemcell.tar.gz"
var primaryDisk = "/dev/sda"

func main() {
	var err error

	//TODO: replace with watch for /dev/sda
	for i := 10 * time.Second; i > time.Duration(0); i -= time.Second {
		fmt.Printf("writing stemcell %s to disk %s in %d second(s)\n", stemcellPath, primaryDisk, i/time.Second)
		time.Sleep(1 * time.Second)
	}

	//confirm first run
	//TODO: check actual partition for UUID
	fmt.Printf("checking stemcell: %s\n", stemcellPath)
	_, err = os.Stat(stemcellPath)
	if err != nil {
		//TODO on second invocation
		// confirm /sysroot is present
		// write vcap settings
		fmt.Println("stemcell not present or already written")

		os.Exit(0)
	}

	fmt.Printf("writing stemcell %s to disk %s\n", stemcellPath, primaryDisk)

	var hardDiskFile *os.File

	hardDiskFile, err = os.OpenFile(primaryDisk, os.O_RDWR|os.O_SYNC, 0600)
	if err != nil {
		panic(err)
	}
	defer hardDiskFile.Close()

	var stemcellTarGzFile *os.File
	stemcellTarGzFile, err = os.Open(stemcellPath)
	if err != nil {
		panic(err)
	}
	defer stemcellTarGzFile.Close()

	stemcellClient := stemcell.NewStemcellClient()
	err = stemcellClient.ExtractStemcellRootDisk(stemcellTarGzFile, func(imageFileReader io.Reader) error {
		var err error

		copiedBytes, err := io.Copy(hardDiskFile, imageFileReader)
		if err != nil {
			return err
		}

		fmt.Printf("done writing stemcell disk: %d bytes\n", copiedBytes)

		return nil
	})

	diskData := &disk.Disk{
		File:              hardDiskFile,
		LogicalBlocksize:  512,
		PhysicalBlocksize: 512,
	}
	table, err := diskData.GetPartitionTable()
	if err != nil {
		panic(err)
	}

	fmt.Printf("DEBUG: Partition type: %#+v\n", table.Type())

	fmt.Printf("removing stemcell\n", stemcellPath)
	err = os.Remove(stemcellPath)
	if err != nil {
		panic(err)
	}

	// syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)

	// find stemcell
	// find config
	// extract/unpack os
	// write agent-settings-config
	//
}
