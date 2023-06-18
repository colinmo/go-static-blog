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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/spf13/cobra"
	"golang.org/x/image/webp"
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
	if !isElementExists(ThumbnailOptions.Type, []string{"jpeg", "gif", "png", "webp"}) {
		return fmt.Errorf("can only use gif, jpeg, webp, or png as thumbnail type [%s]", ThumbnailOptions.Type)
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
			log.Fatalf("%s\n", err.Error())
		}
		// Lets go
		_ = letsGoThumbnail()
	},
}

func letsGoThumbnail() error {
	if ThumbnailOptions.Filename == "" {
		mediaPath := filepath.Join(ConfigData.BaseDir, "media/")
		PrintIfNotSilent(fmt.Sprintf("Recursive: %v\b", ThumbnailOptions.Regenerate))
		total := recursiveMediaThumbnailer(mediaPath)
		PrintIfNotSilent(fmt.Sprintf("Changed %d files in %s\n", total, mediaPath))
	} else {
		makeThumbnail(ThumbnailOptions.Filename)
	}
	return nil
}

func canMakeThumbnailFor(filename string) bool {
	// Check that it's an image file - by extension and detected type
	ext := filepath.Ext(filename)
	contents, err := os.Open(filename)
	if err != nil {
		return false
	}
	body := make([]byte, 512)
	_, err = contents.Read(body)
	contents.Close()
	if err != nil {
		return false
	}
	fileType := http.DetectContentType(body)
	return isElementExists(strings.ToLower(ext), []string{".jpg", ".jpeg", ".gif", ".png", ".webp"}) &&
		(len(filename) < len(ThumbnailOptions.Extension) || filename[len(filename)-len(ThumbnailOptions.Extension):] != ThumbnailOptions.Extension) &&
		isElementExists(fileType, []string{"image/jpeg", "image/gif", "image/png", "image/webp"})
}

func recursiveMediaThumbnailer(directory string) int {
	changedCount := 0

	files, _ := os.ReadDir(directory)
	for _, file := range files {
		name := file.Name()
		dirAndName := filepath.Join(directory, name)
		if file.IsDir() {
			changedCount += recursiveMediaThumbnailer(dirAndName)
		} else {
			ok := makeThumbnail(dirAndName)
			if ok == nil {
				changedCount++
			}
		}
	}
	return changedCount
}

func makeThumbnail(filename string) error {
	if canMakeThumbnailFor(filename) {
		thumbnailFilename := getThumbnailFilename(filename)
		file, err := os.Open(thumbnailFilename)
		madeFile := false
		if err != nil {
			file, _ = os.Create(thumbnailFilename)
			madeFile = true
		}
		defer file.Close()
		// If we're regenerating or if the file does not exist
		if ThumbnailOptions.Regenerate || madeFile {
			img, err := readImage(filename)
			if err != nil {
				return err
			}
			img = resize.Thumbnail(uint(ThumbnailOptions.Width), uint(ThumbnailOptions.Height), img, resize.Lanczos3)
			return writeImage(img, thumbnailFilename)
		}
		return nil
	}
	return fmt.Errorf("cannot make thumbnail for %s", filename)
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

	var img image.Image
	if len(name) > 5 && name[len(name)-5:] == ".webp" {
		img, err = webp.Decode(fd)
	} else {
		img, _, err = image.Decode(fd)
	}
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
