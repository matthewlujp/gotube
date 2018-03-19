package gotube

import (
	"testing"
)

var (
	watchYoutubeURL = "https://www.youtube.com/watch?v=iEPTlhBmwRg" // Music PV "Moves like a jagger" by Maroon5
	embedYoutubeURL = "https://www.youtube.com/embed/iEPTlhBmwRg"
	dummyURL        = "https://www.mytube.com/watch?v=iEPTlhBmwRg"
)

func TestNewPlayer(t *testing.T) {
	// invalid or non youtube url
	if _, err := NewPlayer(dummyURL); err == nil {
		t.Errorf("invalid or non youtube url should be rejected, but didn't for %s", dummyURL)
	}

	// watch youtube url
	player, err := NewPlayer(watchYoutubeURL)
	if err != nil {
		t.Errorf("failed to instantiate player from %s", watchYoutubeURL)
	} else if player.IsEmbed() {
		t.Errorf("embed player was instantiated from %s, watch player expected", watchYoutubeURL)
	}

	// embed youtube url
	player, err = NewPlayer(embedYoutubeURL)
	if err != nil {
		t.Errorf("failed to instantiate player from %s", embedYoutubeURL)
	} else if !player.IsEmbed() {
		t.Errorf("watch player was instantiated from %s, embed player expected", embedYoutubeURL)
	}
}
