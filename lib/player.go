package gotube

import (
	"fmt"
)

var (
	logger *errorLogger
)

type Player interface {
	IsEmbed() bool
	FetchStreamManifests() ([]*Stream, error)
	GetStreams() []*Stream
}

func init() {
	logger = newLogger(true)
}

// NewPlayer returns a instance which implements Player according to a given url
func NewPlayer(url string) (Player, error) {
	bURL := []byte(url)
	if youtubeWatchURLPattern.Match(bURL) {
		return &watchPlayer{client: &youtubeClient{}, url: url}, nil
	} else if youtubeEmbedURLPattern.Match(bURL) {
		return &embedPlayer{client: &youtubeClient{}, url: url}, nil
	}
	return nil, fmt.Errorf("unexpected URL format %s", url)
}
