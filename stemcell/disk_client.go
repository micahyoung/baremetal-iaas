package stemcell

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	ext4 "github.com/dsoprea/go-ext4"
)

var imageRootDiskKernelPath = "/vmlinuz"
var imageRootDiskInitRDPath = "/initrd.img"

type DiskClient struct{}

func NewDiskClient() *DiskClient {
	return &DiskClient{}
}

func (c *DiskClient) ExtractRootDiskBootFiles(rootDiskPath string, kernelFile *os.File, initRDFile *os.File) error {
	var err error

	var rootDiskFile *os.File

	var rootPartitionTempFile *os.File
	if rootPartitionTempFile, err = ioutil.TempFile("", "baremetal-root-partition.img"); err != nil {
		return err
	}
	defer os.RemoveAll(rootPartitionTempFile.Name())

	if rootDiskFile, err = os.Open(rootDiskPath); err != nil {
		return err
	}

	skipMbrBytes := int64(63 * 512)
	skippedBytes, err := rootDiskFile.Seek(skipMbrBytes, 0)
	if err != nil {
		return err
	}

	fmt.Printf("skipped bytes: %d\n", skippedBytes)

	copiedBytes, err := io.Copy(rootPartitionTempFile, rootDiskFile)
	if err != nil {
		return err
	}
	defer rootPartitionTempFile.Close()

	fmt.Printf("copied bytes: %d\n", copiedBytes)

	inodeNumber := ext4.InodeRootDirectory

	f := rootPartitionTempFile
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Seek(ext4.Superblock0Offset, io.SeekStart)
	if err != nil {
		return err
	}

	sb, err := ext4.NewSuperblockWithReader(f)
	if err != nil {
		return err
	}

	bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	if err != nil {
		return err
	}

	bgd, err := bgdl.GetWithAbsoluteInode(inodeNumber)
	if err != nil {
		return err
	}

	dw, err := ext4.NewDirectoryWalk(f, bgd, inodeNumber)
	if err != nil {
		return err
	}

	allEntries := make([]string, 0)

	for {
		fullPath, de, err := dw.Next()

		fmt.Printf("%s\n", fullPath)
		if err == io.EOF {
			break
		} else if err != nil {
			continue
			// log.Panic(err)
		}

		description := fmt.Sprintf("%s: %s", fullPath, de.String())
		allEntries = append(allEntries, description)
	}

	sort.Strings(allEntries)

	for _, entryDescription := range allEntries {
		fmt.Println(entryDescription)
	}

	return nil
}
