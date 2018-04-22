package plexshowupdater

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
)

type episode struct {
	XMLName    xml.Name `xml:"episodedetails"`
	Title      string   `xml:"title"`
	Season     int      `xml:"season"`
	Episode    int      `xml:"episode"`
	Plot       string   `xml:"plot,omitempty"`
	UserRating float64  `xml:"userrating,omitempty"`
	Aired      string   `xml:"aired,omitempty"`
}

// NFOFilenameForVideo returns 'filename' with the extension '.nfo'
func NFOFilenameForVideo(filename string) string {
	ext := path.Ext(filename)
	return filename[0:len(filename)-len(ext)] + ".nfo"
}

// CreateNFOFile creates a Plex episode NFO file for the video in 'filename'.
func CreateNFOFile(title string, season, episodeNum int, description string, rating float64, uploadDate, filename string) error {
	e := &episode{
		Title:      title,
		Season:     season,
		Episode:    episodeNum,
		Plot:       description,
		UserRating: rating,
		Aired:      uploadDate,
	}

	output, err := xml.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}

	contents := fmt.Sprintf(
		"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\" ?>\n%s\n",
		output,
	)

	nfofile := NFOFilenameForVideo(filename)
	return ioutil.WriteFile(nfofile, []byte(contents), 0644)
}
