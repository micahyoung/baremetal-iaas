package stemcell

import (
	"io"
	"regexp"

	ext4 "github.com/dsoprea/go-ext4"
)

var imageRootDiskKernelPattern = "boot/vmlinuz-*"
var imageRootDiskInitRDPattern = "boot/initrd.img-*"

type DiskClient struct{}

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func NewDiskClient() *DiskClient {
	return &DiskClient{}
}

func (c *DiskClient) ExtractRootDiskKernel(rootDiskReadSeeker io.ReaderAt, callback func(io.Reader) error) error {
	return c.findFileReaderInRootDiskReader(rootDiskReadSeeker, imageRootDiskKernelPattern, callback)
}

func (c *DiskClient) ExtractRootDiskInitRD(rootDiskReadSeeker io.ReaderAt, callback func(io.Reader) error) error {
	return c.findFileReaderInRootDiskReader(rootDiskReadSeeker, imageRootDiskInitRDPattern, callback)
}

func (c *DiskClient) findFileReaderInRootDiskReader(rootDiskReadSeeker io.ReaderAt, searchFilePattern string, callback func(io.Reader) error) error {
	var err error

	// Partition hacks:  https://en.wikipedia.org/wiki/Master_boot_record#Partition_table_entries
	skipMbrBytes := int64(63 * 512) // skip mbr record
	maxBytes := int64(2 * TB)       // max partition size

	partitionReader := io.NewSectionReader(rootDiskReadSeeker, skipMbrBytes, maxBytes)

	if err != nil {
		return err
	}

	_, err = partitionReader.Seek(ext4.Superblock0Offset, io.SeekStart)
	if err != nil {
		return err
	}

	superBlock, err := ext4.NewSuperblockWithReader(partitionReader)
	if err != nil {
		return err
	}

	blockGroupDescriptorList, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(partitionReader, superBlock)
	if err != nil {
		return err
	}

	blockGroupDescriptor, err := blockGroupDescriptorList.GetWithAbsoluteInode(ext4.InodeRootDirectory)
	if err != nil {
		return err
	}

	directoryWalk, err := ext4.NewDirectoryWalk(partitionReader, blockGroupDescriptor, ext4.InodeRootDirectory)
	if err != nil {
		return err
	}

	for {
		fullPath, directoryEntry, err := directoryWalk.Next()

		if matched, _ := regexp.MatchString(searchFilePattern, fullPath); matched {
			var fileInode *ext4.Inode
			fileInodeNumber := int(directoryEntry.Data().Inode)
			fileInode, err = ext4.NewInodeWithReadSeeker(blockGroupDescriptor, partitionReader, fileInodeNumber)
			extentNavigator := ext4.NewExtentNavigatorWithReadSeeker(partitionReader, fileInode)

			var inodeReader io.Reader
			inodeReader = ext4.NewInodeReader(extentNavigator)

			return callback(inodeReader)
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}
