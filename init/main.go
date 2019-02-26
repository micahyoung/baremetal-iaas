package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/diskfs/go-diskfs/disk"

	"github.com/micahyoung/baremetal-iaas/stemcell"
)

func main() {

	fmt.Printf("\n")
	var err error
	var hardDiskFile *os.File

	hardDiskFile, err = os.OpenFile("/dev/sda", os.O_RDWR|os.O_SYNC, 0600)
	if err != nil {
		panic(err)
	}
	defer hardDiskFile.Close()

	var stemcellTarGzFile *os.File
	stemcellTarGzFile, err = os.Open("/stemcell.tar.gz")
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

		fmt.Printf("root disk: %d bytes\n", copiedBytes)

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

	fmt.Printf("Partition type: %#+v\n", table.Type())

	fmt.Printf("rebooting in 5 seconds\n")
	time.Sleep(5 * time.Second)
	fmt.Printf("rebooting\n")

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)

	// find stemcell
	// find config
	// extract/unpack os
	// write agent-settings-config
	//
}
