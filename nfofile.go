package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
)

func createNFOFile(vi *videoInfo) error {
	type episode struct {
		XMLName    xml.Name `xml:"episodedetails"`
		Title      string   `xml:"title"`
		Season     int      `xml:"season"`
		Episode    int      `xml:"episode"`
		Plot       string   `xml:"plot,omitempty"`
		UserRating float64  `xml:"userrating,omitempty"`
		Aired      string   `xml:"aired,omitempty"`
	}
	e := &episode{
		Title:      vi.Title,
		Season:     vi.Season,
		Episode:    vi.Episode,
		Plot:       vi.Description,
		UserRating: vi.Rating,
		Aired:      vi.UploadDate,
	}

	output, err := xml.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}

	contents := fmt.Sprintf(
		"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\" ?>\n%s\n",
		output,
	)

	ext := path.Ext(vi.Filename)
	nfofile := vi.Filename[0:len(vi.Filename)-len(ext)] + ".nfo"
	return ioutil.WriteFile(nfofile, []byte(contents), 0644)
}
