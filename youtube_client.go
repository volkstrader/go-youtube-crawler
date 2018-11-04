package main

import (
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"strings"
)

// YouTubeClient is an API client for YouTube Data API v3
type YouTubeClient struct {
	apikey   string
	service  *youtube.Service
	pageSize int64
}

// NewClient create client connecting to YouTube Service with API Key
func NewClient(apiKey string, resultsPerPage int64) (*YouTubeClient, error) {
	client := &YouTubeClient{
		apikey:   apiKey,
		pageSize: resultsPerPage,
	}

	httpClient := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}

	service, err := youtube.New(httpClient)
	if err != nil {
		return nil, err
	}

	client.service = service
	return client, nil
}

// SearchVideoByKeywords make http call to YouTube search API for videos id by keywords
func (client *YouTubeClient) SearchVideoByKeywords(keywords []string, nextPageToken string) ([]string, string, error) {
	query := strings.Join(keywords, ",")
	call := client.service.Search.
		List("id").
		Type("video").
		Q(query).
		MaxResults(client.pageSize)

	if nextPageToken != "" {
		call.PageToken(nextPageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	ids := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		ids[i] = item.Id.VideoId
	}

	return ids, resp.NextPageToken, nil
}

// ListVideosByIds make http call to YouTube videos API for video metadata by list of ids, no more than 50 ids per call
func (client *YouTubeClient) ListVideosByIds(ids []string) ([]*youtube.Video, error) {
	query := strings.Join(ids, ",")
	call := client.service.Videos.
		List("snippet").
		Id(query).
		MaxResults(50)

	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}
