package main

//
//import (
//	"github.com/rs/zerolog/log"
//	"google.golang.org/api/googleapi/transport"
//	"google.golang.org/api/youtube/v3"
//	"net/http"
//)
//
//type YouTubeClient struct {
//	apikey    string
//	apiClient *http.Client
//}
//
//func New(apiKey string) (YouTubeClient, error) {
//	client := YouTubeClient{
//		apikey: apiKey,
//	}
//
//	return client, nil
//}
//
//func SearchVideoByKeywords(apiKey string, keywords []string) {
//	client := &http.Client{
//		Transport: &transport.APIKey{Key: apiKey},
//	}
//	service, err := youtube.New(client)
//	if err != nil {
//		log.Fatal().Err(err).Msg("Fatal Error creating new YouTube client")
//	}
//}
