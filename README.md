# go-youtube-crawler
* Concurrently grab YouTube videos metadata base on provided keywords
* Rate limit API calls, configurable to N request/min
* Save data to a file for query

## Config File
Require config.[toml](https://github.com/toml-lang/toml) file in the executable directory

```toml
title = "GO YouTube Crawler configuration"
log_level = "info"         #debug, info, warn, error, fatal, panic

[crawler]
max_videos = 1000           # how many videos to grab at most
max_videos_per_call = 50    # max video per each API call, NO MORE THAN 50 videos
calls_per_minute = 20       # API calls rate limit: x api calls/minute
concurrent_calls = 2        # Number of concurrent API calls

[database]
file = "videos.db"          # save result to the filename at the current directory
```

### How much time does it take to grab 1000 videos
According to [APIs Explorer](https://developers.google.com/apis-explorer/?hl=en_US#p/youtube/v3/youtube.search.list), 50 is the maximum number of items returns from video search.  It takes 2 API calls, search and list, to grab the set of 50 videos.  

> runtime_minute = max_video * 2 / max_videos_per_call / calls_per_minute

When it is set to `80 calls per minute`, the expected runtime will be around `0.5 minutes` or `30 seconds`.
   
## Enviornment Variables
YOUTUBE_API_KEY=`Your Google Cloud Platform API key credential authorized for YouTube Data API v3`

Log on to [GCP Console](https://console.cloud.google.com/apis/credentials) to create or retrieve your API key credential, authorized for **YouTube Data API v3** 

## 3rd Party Dependency
* go get -u [github.com/rs/zerolog/log](https://github.com/rs/zerolog) `for logging`
* go get -u [github.com/spf13/viper](https://github.com/spf13/viper) `for configuration`
* go get -u [google.golang.org/api/youtube/v3](https://godoc.org/google.golang.org/api/youtube/v3) `for youtube`
* go get -u [google.golang.org/api/googleapi/transport](https://google.golang.org/api/googleapi/transport) `for youtube`
* go get [github.com/google/uuid](https://github.com/google/uuid) `for uuid`
* go get -u [github.com/tidwall/buntdb](https://github.com/tidwall/buntdb) `for database`

## Build
``` bash
> go build
```

## Run
#### Grab Mode
Grab video metadata from YouTube Data API, results will be saved to database file.
``` bash
> go-youtube-crawler --keywords term1 --keywords "term 2"
```

`--keywords` 0 or more keywords can be specified for searching video from YouTube, use double quote to surround the term with space.

#### Find Mode
Find videos from the database file.
``` bash
> go-youtube-crawler --find --title *trump* --desc *google*  
```

`--find` Enable find mode, return all videos when both `--title` and `--desc` are not provided 

`--title` 0 or 1 title field filter can be specified, filter value accepts only [go regular expression syntax](https://golang.org/pkg/regexp/syntax/).

`--desc` 0 or 1 description field filter can be specified, filter value accepts only [go regular expression syntax](https://golang.org/pkg/regexp/syntax/).

## Reference
* [YouTube API Reference - Search by keyword](https://developers.google.com/youtube/v3/code_samples/go#search_by_keyword)
* YouTube Data API Overview - [Quota usage](https://developers.google.com/youtube/v3/getting-started#quota) 