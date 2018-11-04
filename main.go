package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/api/youtube/v3"
	"sync"
	"time"
)

var (
	keywords    *[]string
	client      *YouTubeClient
	taskCtrl    *TaskController
	maxVideos   int64
	searchCount int64
	fetchCount  int64
	nextCh      chan string
	saveCh      chan []*youtube.Video
	wg          sync.WaitGroup
)

func initEnv() {
	// init environment variables
	const youtubeAPIKey = "YOUTUBE_API_KEY"
	if err := viper.BindEnv(youtubeAPIKey); err != nil || viper.Get(youtubeAPIKey) == nil {
		if err == nil {
			err = fmt.Errorf("fatal error missing environment variable: %s.\nPlease reference to README.md Environment Variable section", youtubeAPIKey)
		}
		panic(err)
	}

	// init config file
	viper.SetConfigFile("config.toml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s.\nPlease reference to README.md Config File section", err))
	}

	// init command line flags
	keywords = pflag.StringArray("keywords", nil, "")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil || keywords == nil || len(*keywords) == 0 {
		if err == nil {
			err = fmt.Errorf("missing keywords flag.\n Usage: go-youtube-crawler --keywords term1 --keywords term2 ... ")
		}
		panic(err)
	}

	// init logging
	zerolog.TimeFieldFormat = ""
	logLevel := viper.GetString("log_level")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		panic(fmt.Errorf("Unknown log_level: %s", logLevel))
	}
	zerolog.SetGlobalLevel(level)

	log.Debug().Interface("viper", viper.AllSettings()).Msg("show all config settings")

	// init main package vars
	// TODO: validate config values
	apiKey := viper.GetString("youtube_api_key")
	maxVideos = viper.GetInt64("crawler.max_videos")
	videosPerPage := viper.GetInt64("crawler.max_videos_per_call")
	callsPerMinute := viper.GetInt("crawler.calls_per_minute")
	concurrency := viper.GetInt("crawler.concurrent_calls")

	taskCtrl = NewController(callsPerMinute, concurrency)
	client, err = NewClient(apiKey, videosPerPage)
	if err != nil {
		panic(fmt.Errorf("fatal error create new YouTubeClient, error: %s", err))
	}

	nextCh = make(chan string)
	saveCh = make(chan []*youtube.Video)
}

func search(nextPageToken string) Task {
	return func(ctx context.Context, trace TaskTrace) {
		endOfSearch := false
		log.Info().
			Str("id", trace.TraceID).
			Str("nextPageToken", nextPageToken).
			Msg("search started")

		trace.Started = time.Now()
		ids, nextToken, err := client.SearchVideoByKeywords(*keywords, nextPageToken)
		if err != nil {
			log.Error().
				Str("id", trace.TraceID).
				Interface("trace", trace).
				Err(err).
				Msg("search error")
			return
		}

		log.Info().
			Str("id", trace.TraceID).
			Int64("searchCount", searchCount).
			Int("newVidoes", len(ids)).
			Int64("maxVideos", maxVideos).
			Str("nextToken", nextToken).
			Msg("new search result")
		searchCount += int64(len(ids))

		// search next page
		if searchCount < maxVideos && nextToken != "" {
			nextCh <- nextToken
		} else {
			endOfSearch = true
			// found enough results OR no more results
			// but still need to fetch the remaining videos
			log.Info().
				Str("id", trace.TraceID).
				Int64("searchCount", searchCount).
				Int64("maxVideos", maxVideos).
				Msg("ALL search completed")
		}

		// queue to fetch video metadata
		if fetchCount >= maxVideos {
			// fetch enough videos, do not fetch more
			log.Info().
				Str("id", trace.TraceID).
				Int64("fetchCount", fetchCount).
				Int64("maxVideos", maxVideos).
				Msg("fulfilled fetch videos quota")
		} else if len(ids) > 0 {
			taskCtrl.taskCh <- fetch(ids)
		} else {
			// no results
			log.Info().
				Str("id", trace.TraceID).
				Bool("endOfSearch", endOfSearch).
				Msg("NO search results")

			if endOfSearch {
				taskCtrl.End()
			}
		}

		trace.Completed = time.Now()
		log.Info().
			Str("id", trace.TraceID).
			Int("results", len(ids)).
			Interface("trace", trace).
			Msg("search completed")
	}
}

func fetch(ids []string) Task {
	return func(ctx context.Context, trace TaskTrace) {
		log.Info().
			Str("id", trace.TraceID).
			Strs("ids", ids).
			Int("videos", len(ids)).
			Msg("fetch videos started")

		trace.Started = time.Now()
		videos, err := client.ListVideosByIds(ids)
		if err != nil {
			log.Error().
				Str("id", trace.TraceID).
				Interface("trace", trace).
				Err(err).
				Msg("fetch videos error")
		}

		saveCh <- videos

		trace.Completed = time.Now()
		log.Info().
			Str("id", trace.TraceID).
			Int("videos", len(videos)).
			Interface("trace", trace).
			Msg("fetch videos completed")
	}
}

func main() {
	initEnv()
	ctx, cancel := context.WithCancel(context.Background())
	taskCtrl.Start(ctx)
	defer func() {
		close(nextCh)
		close(saveCh)
		taskCtrl.End()
	}()

	done := false

	wg.Add(1)
	go func() {
		defer wg.Done()

		taskCtrl.taskCh <- search("")
		for i := int64(1); !done; i++ {
			select {
			case <-ctx.Done():
				done = true
				return
			case nextToken := <-nextCh:
				taskCtrl.taskCh <- search(nextToken)
				break
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		count := int64(0)

		for !done {
			select {
			case <-ctx.Done():
				done = true
				break
			case videos := <-saveCh:
				for _, video := range videos {
					count++
					fmt.Println(video.Id)
					fmt.Println(video.Snippet.Title)
					fmt.Println("===========================")
					//fmt.Println(video.Snippet.Description)
				}

				if count >= maxVideos {
					log.Info().Msg("all videos saved, exit gracefully")
					done = true
					cancel()
				}

				break
			case <-taskCtrl.doneCh:
				log.Info().Msg("all task workers exited")
				done = true
				cancel()
				break
			}
		}
	}()

	// wait for exit
	wg.Wait()
	log.Info().Msg("all go routines exited, terminate")
}
