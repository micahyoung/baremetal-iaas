package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	ext4 "github.com/dsoprea/go-ext4"
)

func main() {
	if err := foo(); err != nil {
		panic(err)
	}
}

func foo() error {
	var err error

	inodeNumber := ext4.InodeRootDirectory

	f, err := os.Open("test.img")
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

	for {
		fullPath, directoryEntry, err := dw.Next()

		if matched, _ := regexp.MatchString("boot/vmlinuz-*", fullPath); matched {
			fmt.Printf("%d\n", directoryEntry.Data().RecLen)

			var inode *ext4.Inode
			inodeNumber := int(directoryEntry.Data().Inode)
			inode, err = ext4.NewInodeWithReadSeeker(bgd, f, inodeNumber)
			en := ext4.NewExtentNavigatorWithReadSeeker(f, inode)

			var r io.Reader
			r = ext4.NewInodeReader(en)

			var content []byte
			if content, err = ioutil.ReadAll(r); err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", fullPath, content)

			return nil
		}
		if err == io.EOF {
			break
		} else if err != nil {
			continue
			// log.Panic(err)
		}
	}
	return nil
}
