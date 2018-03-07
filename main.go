package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/tumblr/tumblr.go"
	tumblr_go "github.com/tumblr/tumblrclient.go"
)

func main() {
	//vi := downloadURL("https://www.youtube.com/watch?v=C0DPdy98e4c", 3, 2)

	//err := createNFOFile(vi)
	//if err != nil {
	//    panic(err)
	//}

	client := tumblr_go.NewClientWithToken(
		os.Getenv("TUMBLR_CONSUMER_KEY"),
		os.Getenv("TUMBLR_CONSUMER_SECRET"),
		os.Getenv("TUMBLR_TOKEN"),
		os.Getenv("TUMBLR_TOKEN_SECRET"),
	)
	resp, err := tumblr.GetPosts(client, "softwaredevvideos", url.Values{})
	if err != nil {
		panic(err)
	}

	allPosts, err := resp.All()
	if err != nil {
		panic(err)
	}

	for _, post := range allPosts {
		switch pt := post.(type) {
		case *tumblr.LinkPost:
			fmt.Printf(" link   %d %v %v\n", pt.Id, pt.Url, pt.Tags)
		case *tumblr.VideoPost:
			fmt.Printf(" video  %d %v %v\n", pt.Id, pt.PermalinkUrl, pt.Tags)
		}
	}
}
