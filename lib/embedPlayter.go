package gotube

import (
	"regexp"
)

var (
	youtubeEmbedURLPattern = regexp.MustCompile(`https?://www.youtube.com/embed/(\w{11})`)
)

type embedPlayer struct {
	client  client
	streams []*Stream
	url     string
}

func (p *embedPlayer) IsEmbed() bool {
	return true
}

func (p *embedPlayer) FetchStreamManifests() ([]*Stream, error) {
	return make([]*Stream, 0), nil
}

func (p *embedPlayer) GetStreams() []*Stream {
	return make([]*Stream, 0)
}
