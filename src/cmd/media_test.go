package cmd

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestGetThumbnailFilename(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	name := getThumbnailFilename("/usr/local/zend/apache2/mep.jpg")
	expect := "/usr/local/zend/apache2/mep-thumb.jpg"

	if name != expect {
		t.Fatalf("Bad filename result %s|%s\n", name, expect)
	}
	name = getThumbnailFilename("/usr/local/zend/apache2/mep.gif")
	expect = "/usr/local/zend/apache2/mep-thumb.jpg"

	if name != expect {
		t.Fatalf("Bad filename result %s|%s\n", name, expect)
	}
}

func TestThumbnailGeneration(t *testing.T) {
	ThumbnailOptions.Extension = "-thumb.jpg"
	ConfigData.BaseDir = `F:\Dropbox\swap\golang\vonblog\features\tests\thumbnail\`
	ThumbnailOptions.Width = 100
	ThumbnailOptions.Height = 100
	ThumbnailOptions.Type = "jpeg"
	ThumbnailOptions.Regenerate = true
	err := makeThumbnail(`F:\Dropbox\swap\golang\vonblog\features\tests\thumbnail\LogVisualiser.png`)
	if err != nil {
		t.Fatalf("Ded %v\n", err)
	}
}

func TestDetectImage(t *testing.T) {
	contents, err := os.Open(`F:\Dropbox\swap\golang\vonblog\features\tests\thumbnail\LogVisualiser.png`)
	if err != nil {
		t.Fatalf("Could not process file %s\n%v\n", `F:\Dropbox\swap\golang\vonblog\features\tests\thumbnail\LogVisualiser.png`, err)
	}
	defer contents.Close()
	body := make([]byte, 512)
	_, err = contents.Read(body)
	if err != nil {
		t.Fatalf("Cound not read file %s\n", `F:\Dropbox\swap\golang\vonblog\features\tests\thumbnail\LogVisualiser.png`)
	}
	fileType := http.DetectContentType(body)
	if fileType != "image/png" {
		fmt.Printf("%v\n", body)
		t.Fatalf("Detect failed for %s\n", fileType)
	}
}
