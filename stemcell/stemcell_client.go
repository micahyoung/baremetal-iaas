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

func (c *StemcellClient) ExtractStemcellRootDisk(stemcellPath string, rootDiskFile *os.File) error {
	var err error
	var stemcellTarGzFile *os.File

	if stemcellTarGzFile, err = os.Open(stemcellPath); err != nil {
		return err
	}
	defer stemcellTarGzFile.Close()

	var stemcellGzipReader *gzip.Reader
	if stemcellGzipReader, err = gzip.NewReader(stemcellTarGzFile); err != nil {
		return err
	}

	stemcellTarReader := tar.NewReader(stemcellGzipReader)

	for {
		stemcellTarHeader, err := stemcellTarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		stemcellFileName := stemcellTarHeader.Name

		switch stemcellTarHeader.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fmt.Println("Stemcell File Name: ", stemcellFileName)

			if stemcellFileName == imageTarGzName {
				imageGzipReader, err := gzip.NewReader(stemcellTarReader)
				if err != nil {
					return err
				}

				imageTarReader := tar.NewReader(imageGzipReader)

				for {
					imageTarHeader, err := imageTarReader.Next()

					if err == io.EOF {
						break
					}

					if err != nil {
						return err
					}

					imageFileName := imageTarHeader.Name

					switch stemcellTarHeader.Typeflag {
					case tar.TypeDir:
						continue
					case tar.TypeReg:
						fmt.Println("Image File Name: ", imageFileName)

						if imageFileName == imageRootDiskName {
							var bytesWritten int64
							if bytesWritten, err = io.Copy(rootDiskFile, imageTarReader); err != nil {
								return err
							}

							fmt.Printf("wrote %d bytes to %s\n", bytesWritten, rootDiskFile.Name())
						}

					default:
						fmt.Printf("%s : %c %s %s\n",
							"Yikes! Unable to figure out type",
							stemcellTarHeader.Typeflag,
							"in file",
							stemcellFileName,
						)
					}

				}

			}
		default:
			fmt.Printf("%s : %c %s %s\n",
				"Yikes! Unable to figure out type",
				stemcellTarHeader.Typeflag,
				"in file",
				stemcellFileName,
			)
		}
	}

	return nil
}
