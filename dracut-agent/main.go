package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/diskfs/go-diskfs/disk"

	"github.com/micahyoung/baremetal-iaas/stemcell"
)

var stemcellPath = "/stemcell.tar.gz"
var primaryDisk = "/dev/sda"
var logPath = "/baremetal-dracut-agent.log"
var logFile *os.File

func setupHooks() error {
	var err error

	stages := []string{
		"cmdline",
		"pre-udev",
		"pre-trigger",
		"initqueue",
		"pre-mount",
		"mount",
		"pre-pivot",
		"cleanup",
	}

	for _, stage := range stages {
		hookPath := fmt.Sprintf("/lib/dracut/hooks/%s/baremetal-dracut-agent.sh", stage)

		fmt.Fprintf(logFile, "BAREMETAL (setup-hooks) %s\n", hookPath)

		_, err = os.Stat(hookPath)
		if err != nil {
			content := fmt.Sprintf("#!/bin/bash\n/bin/dracut-cmdline-ask %s\n", stage)
			err = ioutil.WriteFile(hookPath, []byte(content), 0777)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func initqueue() error {
	var err error

	//confirm first run
	//TODO: check actual partition for UUID
	fmt.Fprintf(logFile, "checking stemcell: %s\n", stemcellPath)
	_, err = os.Stat(stemcellPath)
	if err != nil {
		//TODO on second invocation
		// confirm /sysroot is present
		// write vcap settings
		fmt.Fprintf(logFile, "stemcell not present or already written\n")

		return nil
	}

	//TODO: replace with watch for /dev/sda
	for i := 10 * time.Second; i > time.Duration(0); i -= time.Second {
		fmt.Fprintf(logFile, "writing stemcell %s to disk %s in %d second(s)\n", stemcellPath, primaryDisk, i/time.Second)
		time.Sleep(1 * time.Second)
	}

	fmt.Fprintf(logFile, "writing stemcell %s to disk %s\n", stemcellPath, primaryDisk)

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

		fmt.Fprintf(logFile, "done writing stemcell disk: %d bytes\n", copiedBytes)

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

	fmt.Fprintf(logFile, "DEBUG: Partition type: %#+v\n", table.Type())

	fmt.Fprintf(logFile, "removing stemcell\n", stemcellPath)
	err = os.Remove(stemcellPath)
	if err != nil {
		panic(err)
	}

	return nil
}

func prePivot() error {
	var err error

	mountTestPath := "/sysroot/baz"
	content := []byte("here")
	err = ioutil.WriteFile(mountTestPath, content, 0777)
	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	var err error

	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	if len(os.Args) == 1 {
		fmt.Fprintf(logFile, "BAREMETAL (setup-hooks) setting up hooks\n")
		setupHooks()
	} else {
		switch os.Args[1] {
		case "cmdline":
			fmt.Fprintf(logFile, "BAREMETAL (cmdline)\n")
		case "pre-udev":
			fmt.Fprintf(logFile, "BAREMETAL (pre-udev)\n")
		case "pre-trigger":
			fmt.Fprintf(logFile, "BAREMETAL (pre-trigger)\n")
		case "initqueue":
			fmt.Fprintf(logFile, "BAREMETAL (initqueue)\n")
			initqueue()
		case "mount":
			fmt.Fprintf(logFile, "BAREMETAL (mount)\n")
		case "pre-pivot":
			fmt.Fprintf(logFile, "BAREMETAL (pre-pivot)\n")
			prePivot()
		case "cleanup":
			fmt.Fprintf(logFile, "BAREMETAL (cleanup)\n")
		}
	}

	if _, err = os.Stat(logPath); err == nil {
		if _, err = os.Stat("/sysroot/bin"); err == nil {
			logContent, err := ioutil.ReadFile(logPath)
			if err != nil {
				panic(err)
			}

			err = ioutil.WriteFile("/sysroot/baremetal-dracut-agent.log", logContent, 0666)
			if err != nil {
				panic(err)
			}
		}
	}
}
