package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

type Video struct {
	MediaName string
	MediaID   string
	VideoID   string
}

func main() {
	fmt.Println("Parsing trailers")

	trailers, _ := FindTrailers()

	fmt.Println("Generating feed")

	feed := &feeds.RssFeed{
		Title:          "Trailers RSS Feed",
		Link:           "http://mediafeeds.malone.me.uk",
		Description:    "An RSS feed of IMDB trailers.",
		ManagingEditor: "CjMalone@mail.com (Cj Malone)",
		Category:       "Trailers",
	}

	feed.Items = make([]*feeds.RssItem, len(trailers))

	for i := 0; i < len(trailers); i++ {
		fileInfo, err := os.Stat("/var/www/mediafeeds/Trailers/" + trailers[i].VideoID + ".mp4")
		if err != nil {
			cmd := exec.Command("youtube-dl", "-o", "/var/www/mediafeeds/Trailers/"+trailers[i].VideoID+".mp4", "http://www.imdb.com/video/imdb/"+trailers[i].VideoID+"/imdb/embed")

			var out bytes.Buffer
			cmd.Stdout = &out

			fmt.Printf("Downloading %s\n", trailers[i].MediaName)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				fmt.Println(out.String())
			}
		}

		fileInfo, err = os.Stat("/var/www/mediafeeds/Trailers/" + trailers[i].VideoID + ".mp4")
		fileLength := "0"
		if err != nil {
			fmt.Println(err)
		} else {
			fileLength = strconv.FormatInt(fileInfo.Size(), 10)
		}

		feed.Items[i] = &feeds.RssItem{
			Title:       trailers[i].MediaName,
			Link:        "http://www.imdb.com/title/" + trailers[i].MediaID,
			Description: "A trailer for \"" + trailers[i].MediaName + "\" (" + trailers[i].MediaID + ")",
			Author:      "CjMalone@mail.com (Cj Malone)",
			Category:    "Trailer",
			Enclosure: &feeds.RssEnclosure{
				Url:    "http://mediafeeds.malone.me.uk/Trailers/" + trailers[i].VideoID + ".mp4",
				Length: fileLength,
				Type:   "video/mp4",
			},
			Guid: trailers[i].VideoID,
			PubDate: time.Now().Format(time.RFC1123Z),
		}
	}
	rss, err := feeds.ToXML(feed)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Feed created")

	rssFile, err := os.OpenFile("/var/www/mediafeeds/trailers.xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	defer rssFile.Close()

	if err != nil {
		fmt.Println(err)
	}

	rssFile.WriteString(rss)
}

func FindTrailers() ([]Video, error) {
	doc, err := goquery.NewDocument("http://www.imdb.com/trailers")
	if err != nil {
		return nil, err
	}

	recentTab := doc.Find("#recAddTab")
	lengthStr, exists := recentTab.Find(".gridlist-item").Last().Attr("data-index")
	if !exists {
		return nil, nil //TODO
	}
	length, err := strconv.Atoi(lengthStr)
	var trailers = make([]Video, length)
	current := 0

	recentTab.Find(".gridlist-item").Each(func(i int, s *goquery.Selection) {
		rawTitle := s.Find(".trailer-caption").Text()
		mediaLink, mediaLinkExists := s.Find(".trailer-caption").ChildrenFiltered("a").Attr("href")
		videoLink, videoLinkExists := s.Find(".video-link").Attr("href")
		if !mediaLinkExists || !videoLinkExists {
			return //TODO
		}

		title := rawTitle[2 : len(rawTitle)-5]
		var trailer Video
		trailer.MediaName = title
		trailer.MediaID = mediaLink[7:16] //TODO
		splitLink := strings.Split(videoLink, "/")
		trailer.VideoID = splitLink[3]

		trailers[current] = trailer

		current += 1
	})
	return trailers, nil
}
