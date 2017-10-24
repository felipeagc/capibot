package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/otium/ytdl"
)

const (
	apiURL = "https://www.googleapis.com/youtube/v3/search"
)

var (
	youtubeKey = os.Getenv("GOOGLE_KEY")
)

type youtubeThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type youtubeThumbnails struct {
	Default youtubeThumbnail `json:"default"`
	Medium  youtubeThumbnail `json:"medium"`
	High    youtubeThumbnail `json:"high"`
}

type youtubeVideoID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

type youtubeSnippet struct {
	PublishedAt          time.Time         `json:"publishedAt"`
	ChannelID            string            `json:"channelId"`
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	ChannelTitle         string            `json:"channelTitle"`
	LiveBroadcastContent string            `json:"liveBroadcastContent"`
	Thumbnails           youtubeThumbnails `json:"thumbnails"`
}

type youtubeSearchResponse struct {
	Items []youtubeSearchItem `json:"items"`
}

type youtubeSearchItem struct {
	ID      youtubeVideoID `json:"id"`
	Snippet youtubeSnippet `json:"snippet"`
}

// YoutubeResult contains the information relevant to a youtube search result
type YoutubeResult struct {
	Title        string
	ThumbnailURL string
	Description  string
	PublishedAt  time.Time
	VideoID      string
}

// YoutubeSearch returns video results for a search query
// It excludes livestreams, playlists, etc.
// It also query escapes the search query.
func YoutubeSearch(query string) ([]YoutubeResult, error) {
	query = url.QueryEscape(query)
	url := apiURL + "?part=snippet&maxResults=10&q=" + query + "&type=video&key=" + youtubeKey
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	s := new(youtubeSearchResponse)
	err = json.NewDecoder(resp.Body).Decode(s)
	if err != nil {
		return nil, err
	}

	results := make([]YoutubeResult, 0)

	for _, item := range s.Items {
		if item.Snippet.LiveBroadcastContent != "none" {
			continue
		}
		if item.ID.Kind != "youtube#video" {
			continue
		}
		result := YoutubeResult{
			Title:        item.Snippet.Title,
			ThumbnailURL: item.Snippet.Thumbnails.Default.URL,
			Description:  item.Snippet.Description,
			PublishedAt:  item.Snippet.PublishedAt,
			VideoID:      item.ID.VideoID,
		}
		results = append(results, result)
	}

	return results, nil
}

// YoutubeGetInfo gets information about a video specified though a URL
func YoutubeGetInfo(url string) (*YoutubeResult, error) {
	videoInfo, err := ytdl.GetVideoInfo(url)
	if err != nil {
		return nil, err
	}

	result := YoutubeResult{
		Title:        videoInfo.Title,
		ThumbnailURL: videoInfo.GetThumbnailURL(ytdl.ThumbnailQualityDefault).String(),
		Description:  videoInfo.Description,
		PublishedAt:  videoInfo.DatePublished,
		VideoID:      videoInfo.ID,
	}

	return &result, nil
}
