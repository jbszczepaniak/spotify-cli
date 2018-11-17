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
4. Premium Spotify Account
5. Created Spotify Application under https://beta.developer.spotify.com/dashboard/applications (set redirect URI to http://localhost:8888/spotify-cli)

Go to https://beta.developer.spotify.com/dashboard/applications, find created earlier Spotify Application, find Client ID and Client Secret, and put them in environment variables
```
export SPOTIFY_CLIENT_ID=xxxxxxxxxxxxx
export SPOTIFY_SECRET=yyyyyyyyyyyyyyyy
```

### Running from release

1. Download release for your OS/architecture under https://github.com/jedruniu/spotify-cli/releases
2. Unpack it (i.e. with `tar -xvf spotify-cli_1.0.1_Darwin_x86_64.tar spotify`)
3. Run it (`./spotify-cli`)

### Building from sources

#### Additional prerequisities
1. Go language installed 

#### Steps

1. Get repository
```
go get github.com/jedruniu/spotify-cli
```
2. Go to directory with spotify-cli
```
cd $GOPATH/src/github.com/jedruniu/spotify-cli
```
3. Run dep
```
dep ensure
```
4. Install application
```
go install
```
5. Run application
```
spotify-cli
```

## Running tests

```
go test -v
```
## Built With
* [tui](https://github.com/marcusolsson/tui-go) - Terminal User Interface framework
* [Spotify](https://github.com/zmb3/spotify) - Spotify Web API Wrapper
* [dep](https://github.com/golang/dep) - Go dependency management tool 

## TODO 
* Unit test playback.go
* Unit test main.go
