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

		tarFileName := tarHeader.Name

		switch tarHeader.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if tarFileName == searchFileName {
				return callback(tarReader)
			}
		default:
			//unknown file type
			continue
		}

	}

	return fmt.Errorf("file not found in tar: %s", searchFileName)
}
