package stemcell

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

var imageTarGzName = "image"
var imageRootDiskName = "root.img"

type StemcellClient struct {
}

func NewStemcellClient() *StemcellClient {
	return &StemcellClient{}
}

func (c *StemcellClient) ExtractStemcellRootDisk(stemcellTarGzFile *os.File, callback func(io.Reader) error) error {
	return c.findFileReaderInTarGzReader(stemcellTarGzFile, imageTarGzName, func(stemcellFileReader io.Reader) error {
		return c.findFileReaderInTarGzReader(stemcellFileReader, imageRootDiskName, callback)
	})
}

func (c *StemcellClient) findFileReaderInTarGzReader(fileReader io.Reader, searchFileName string, callback func(io.Reader) error) error {
	gzipReader, err := gzip.NewReader(fileReader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		tarHeader, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		imageFileName := tarHeader.Name

		switch tarHeader.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fmt.Println("Image File Name: ", imageFileName)

			if imageFileName == searchFileName {
				return callback(tarReader)
			}

		default:
			fmt.Printf("%s : %c %s %s\n",
				"Yikes! Unable to figure out type",
				tarHeader.Typeflag,
				"in file",
				imageFileName,
			)
		}

	}

	// need check?
	return nil
}
