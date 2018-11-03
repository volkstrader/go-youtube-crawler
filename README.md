# go-youtube-crawler
* Concurrently grab YouTube videos metadata base on provided keywords
* Rate limit API calls, configurable to N request/min
* Save data to a file for query

## Config File
Require config.[toml](https://github.com/toml-lang/toml) file in the executable directory

```toml
title = "GO YouTube Crawler configuration"

[crawler]
calls_per_minutes = 60  # API calls rate limit: x api calls/minute
concurrent_calls = 2    # Number of concurrent API calls
```
## Enviornment Variables
YOUTUBE_API_KEY=`Your Google Cloud Platform API key credential authorized for YouTube Data API v3`

Log on to [GCP Console](https://console.cloud.google.com/apis/credentials) to create or retrieve your API key credential, authorized for **YouTube Data API v3** 

## 3rd Party Dependency
* go get -u [github.com/rs/zerolog/log](https://github.com/rs/zerolog)
* go get -u [github.com/spf13/viper](https://github.com/spf13/viper)
* go get -u [google.golang.org/api/youtube/v3](https://godoc.org/google.golang.org/api/youtube/v3)
* go get -u [google.golang.org/api/googleapi/transport](https://google.golang.org/api/googleapi/transport)

## Build
``` bash
> go build
```

## Run
```bash
> go-youtube-crawler --keywords term1 --keywords "term 2"
```

## Reference
[YouTube API Reference - Search by keyword](https://developers.google.com/youtube/v3/code_samples/go#search_by_keyword)