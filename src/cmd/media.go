/*
Copyright © 2021 Colin Morris <relapse@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
)

type ThumbnailOptionsS struct {
	Width      int
	Height     int
	Extension  string
	Type       string
	Filename   string
	Regenerate bool
}

var ThumbnailOptions ThumbnailOptionsS

func defaultsForMe() error {
	// Defaults
	if ThumbnailOptions.Width == 0 {
		ThumbnailOptions.Width = int(ConfigData.Thumbnails.Width)
	}
	if ThumbnailOptions.Height == 0 {
		ThumbnailOptions.Height = int(ConfigData.Thumbnails.Height)
	}
	if ThumbnailOptions.Extension == "" {
		ThumbnailOptions.Extension = ConfigData.Thumbnails.Extension
	}
	if !isElementExists(ThumbnailOptions.Type, []string{"jpeg", "gif", "png"}) {
		return errors.New(fmt.Sprintf("Can only use gif, jpeg, or png as thumbnail type [%s]\n", ThumbnailOptions.Type))
	}
	return nil
}

// thumbCmd represents the Media command
var thumbCmd = &cobra.Command{
	Use:   "thumbnail",
	Short: "Creates thumbnails",
	Long:  `Creates thumbnails`,
	Run: func(cmd *cobra.Command, args []string) {
		err := defaultsForMe()
		if err != nil {
			log.Fatalf("Can only use gif, jpeg, or png as thumbnail type [%s]\n", ThumbnailOptions.Type)
		}
		// Lets go
		_ = letsGoThumbnail()
	},
}

func letsGoThumbnail() error {
	if ThumbnailOptions.Filename == "" {
		total := recursiveMediaThumbnailer(ConfigData.BaseDir + "media/")
		fmt.Printf("Changed %d files in %s\n", total, ConfigData.BaseDir+"media/")
	} else {
		makeThumbnail(ThumbnailOptions.Filename)
	}
	return nil
}

func recursiveMediaThumbnailer(directory string) int {
	changedCount := 0

	files, _ := ioutil.ReadDir(directory)
	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			changedCount += recursiveMediaThumbnailer(directory + name + "/")
		} else if !(len(name) > len(ThumbnailOptions.Extension) &&
			name[len(name)-len(ThumbnailOptions.Extension):] == ThumbnailOptions.Extension) {
			// fmt.Printf("Skipping existing thumb %s\n", name)
			if isElementExists(filepath.Ext(name), []string{".jpg", ".jpeg", ".gif", ".png"}) {
				// Check that it's an image file - by extension and detected type
				ok := makeThumbnail(directory + name)
				if ok == nil {
					changedCount++
				}
			}
		}
	}
	return changedCount
}

func makeThumbnail(filename string) error {
	contents, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Could not process file %s\n%v\n", filename, err)
	}
	body := make([]byte, 512)
	_, err = contents.Read(body)
	contents.Close()
	if err != nil {
		log.Fatalf("Cound not read file %s\n", filename)
	}
	fileType := http.DetectContentType(body)
	if isElementExists(fileType, []string{"image/jpeg", "image/gif", "image/png"}) {
		thumbnailFilename := getThumbnailFilename(filename)
		file, err := os.Open(thumbnailFilename)
		if err != nil {
			file, err = os.Create(thumbnailFilename)
		}
		defer file.Close()
		// If we're regenerating or if the file does not exist
		if ThumbnailOptions.Regenerate || errors.Is(err, os.ErrNotExist) {
			img, err := readImage(filename)
			if err != nil {
				log.Fatalf("Could not read the base image %s\n", filename)
			}
			img = resize.Thumbnail(uint(ThumbnailOptions.Width), uint(ThumbnailOptions.Height), img, resize.Lanczos3)
			return writeImage(img, thumbnailFilename)
		}
	} else {
		return errors.New(fmt.Sprintf("Can't make a thumbnail for %s\n", filename))
	}
	return nil
}

func getThumbnailFilename(filename string) string {
	return filename[0:strings.LastIndex(filename, ".")] + ThumbnailOptions.Extension
}

func readImage(name string) (image.Image, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	img, _, err := image.Decode(fd)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func writeImage(img image.Image, name string) error {
	fd, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fd.Close()

	if ThumbnailOptions.Type == "jpeg" {
		x := jpeg.Options{
			Quality: 90,
		}
		return jpeg.Encode(fd, img, &x)
	}
	if ThumbnailOptions.Type == "png" {
		return png.Encode(fd, img)
	}
	if ThumbnailOptions.Type == "gif" {
		x := gif.Options{}
		return gif.Encode(fd, img, &x)
	}
	return errors.New("bad thumbnail type")
}

func isElementExists(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(thumbCmd)
	thumbCmd.Flags().IntVarP(&ThumbnailOptions.Width, "width", "", 0, "Thumbnail width (pixels)")
	thumbCmd.Flags().IntVarP(&ThumbnailOptions.Height, "height", "", 0, "Thumbnail height (pixels)")
	thumbCmd.Flags().StringVarP(&ThumbnailOptions.Extension, "extension", "e", "", "Extension to add to the filename file")
	thumbCmd.Flags().StringVarP(&ThumbnailOptions.Type, "type", "t", "jpeg", "Image type of thumbnail")
	thumbCmd.Flags().StringVarP(&ThumbnailOptions.Filename, "filename", "f", "", "File to process (default: all media files)")
	thumbCmd.Flags().BoolVarP(&ThumbnailOptions.Regenerate, "regenerate", "r", false, "Regenerate the images (otherwise it only creates a thumbnail if the image doesn't have one yet)")
}
