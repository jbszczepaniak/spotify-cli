# Spotify CLI
[![Build Status](https://travis-ci.org/jedruniu/spotify-cli.svg?branch=master)](https://travis-ci.org/jedruniu/spotify-cli)
[![codecov](https://codecov.io/gh/jedruniu/spotify-cli/branch/master/graph/badge.svg)](https://codecov.io/gh/jedruniu/spotify-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/jedruniu/spotify-cli)](https://goreportcard.com/report/github.com/jedruniu/spotify-cli)

Spotify Client which runs in the terminal.

![screenshot](screen_shot.png)

## Getting Started

### Prerequisites
1. Linux/MacOS operating system
2. Google Chrome browser installed
3. Go language installed 
4. Premium Spotify Account
5. Created Spotify Application under https://beta.developer.spotify.com/dashboard/applications (set redirect URI to http://localhost:8888/spotify-cli)

### Installing
1. Go to https://beta.developer.spotify.com/dashboard/applications, find created earlier Spotify Application, find Client ID and Client Secret, and put them in environment variables
```
export SPOTIFY_CLIENT_ID=xxxxxxxxxxxxx
export SPOTIFY_SECRET=yyyyyyyyyyyyyyyy
```

#### With `git clone` and `go build`
2. Clone this repostitory
```
git clone https://github.com/jedruniu/spotify-cli.git
``` 
3. Build application
```
go build
```
4. Run application
```
./spotify-cli
```

#### With downloading released package
2. Go to https://github.com/jedruniu/spotify-cli/releases and download newest package release for your architecture
3. Unpack downloaded package to some directory and cd to it. 
4. Run application
```
./spotify-cli
```

## Running tests

```
go test -v
```
## Built With
* [tui](https://github.com/marcusolsson/tui-go) - Terminal User Interface framework
* [Spotify](https://github.com/zmb3/spotify) - Spotify Web API Wrapper
* [dep](https://github.com/golang/dep) - Go dependency management tool
* [goreleaser](https://goreleaser.com) - Releasing tool for go
