package gotube

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	mock "github.com/matthewlujp/gotube/lib/mocks"
)

const (
	dummyURL                  = "https://www.mytube.com/watch?v=iEPTlhBmwRg"
	validURL                  = "https://www.youtube.com/watch?v=iEPTlhBmwRg"
	jsURL                     = "https://youtube.com/yts/jsbin/player-vfllqtOs7/ja_JP/base.js"
	title                     = "Maroon 5 - Moves Like Jagger ft. Christina Aguilera"
	ageRestrictedURL          = "https://www.youtube.com/watch?v=6LZM3_wp2ps&has_verified=1"
	ageRestrictedEmbedURL     = "https://www.youtube.com/embed/6LZM3_wp2ps"
	ageRestrictedJsURL        = "https://youtube.com/yts/jsbin/player-vflX7BSrP/ja_JP/base.js"
	ageRestrictedTitle        = "Watch_Dogs: Open World Gameplay Premiere Commented [North America]"
	ageRestrictedVideoInfoURL = "https://youtube.com/get_video_info?video_id=6LZM3_wp2ps&eurl=https%3A%2F%2Fyoutube.googleapis.com%2Fv%2F6LZM3_wp2ps&sts=17632"
)

type fakeClient struct {
	client
	fakeGet func(url string) (*http.Response, error)
}

func (c *fakeClient) Get(url string) (*http.Response, error) {
	return c.fakeGet(url)
}

func TestNewDownloader(t *testing.T) {
	// invalid or non youtube url
	if _, err := NewDownloader(dummyURL); err == nil {
		t.Errorf("invalid or non youtube url should be rejected, but didn't for %s", dummyURL)
	}

	// watch youtube url
	if _, err := NewDownloader(validURL); err != nil {
		t.Errorf("failed to instantiate player from %s", validURL)
	}
}

func TestFetchStreamManifests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)

	gomock.InOrder(
		c.EXPECT().Get(validURL).Return(getMockPage()),
		c.EXPECT().Get(jsURL).Return(getContent(mockScriptPath)),
	)

	downloader := YoutubeDownloader{
		url:    validURL,
		client: c,
	}

	if err := downloader.FetchStreams(); err != nil {
		t.Errorf("error while fetching stream manifests, %s", err)
	}

	// check title
	if downloader.title != title {
		t.Errorf("wrong title, exepected %s, got %s", title, downloader.title)
	}

	// check streams
	if len(downloader.Streams) != len(videoStreams) {
		t.Errorf("got %d streams, %d expected", len(downloader.Streams), len(videoStreams))
	} else {
		sort.Slice(
			downloader.Streams,
			func(i, j int) bool { return downloader.Streams[i].itag < downloader.Streams[j].itag },
		)
		sort.Slice(
			videoStreams,
			func(i, j int) bool { return videoStreams[i].itag < videoStreams[j].itag },
		)
		for i, s := range downloader.Streams {
			if !s.equal(videoStreams[i]) {
				t.Errorf("\n%v\nis different from\n%v\n", *s, *videoStreams[i])
			}
		}
	}
}

func TestFetchStreamManifestsForAgeRestricted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)

	gomock.InOrder(
		c.EXPECT().Get(ageRestrictedURL).Return(getContent(mockAgeRestrictedPagePath)),
		c.EXPECT().Get(ageRestrictedEmbedURL).Return(getContent(mockAgeRestrictedEmbedPagePath)),
		c.EXPECT().Get(ageRestrictedVideoInfoURL).Return(getContent(mockAgeRestrictedVideoInfoPath)),
		c.EXPECT().Get(ageRestrictedJsURL).Return(getContent(mockAgeRestrictedScriptPath)),
	)

	downloader := YoutubeDownloader{
		url:    ageRestrictedURL,
		client: c,
	}

	if err := downloader.FetchStreams(); err != nil {
		t.Errorf("error while fetching stream manifests, %s", err)
	}

	// check title
	if downloader.title != ageRestrictedTitle {
		t.Errorf("wrong title, exepected %s, got %s", ageRestrictedTitle, downloader.title)
	}

	// check streams
	if len(downloader.Streams) != len(restrictedVideoStreams) {
		t.Errorf("got %d streams, %d expected", len(downloader.Streams), len(restrictedVideoStreams))
	} else {
		sort.Slice(
			downloader.Streams,
			func(i, j int) bool { return downloader.Streams[i].itag < downloader.Streams[j].itag },
		)
		sort.Slice(
			restrictedVideoStreams,
			func(i, j int) bool { return restrictedVideoStreams[i].itag < restrictedVideoStreams[j].itag },
		)
		for i, s := range downloader.Streams {
			if !s.equal(restrictedVideoStreams[i]) {
				t.Errorf("\n%v\nis different from\n%v\n", *s, *restrictedVideoStreams[i])
			}
		}
	}

}

func TestInflateStream(t *testing.T) {
	// case 1 (for noramal video)
	rawStream := "type=video%2Fmp4%3B+codecs%3D%22avc1.64001F%2C+mp4a.40.2%22\\u0026itag=22\\u0026url=https%3A%2F%2Fr1---sn-ogul7n7s.googlevideo.com%2Fvideoplayback%3Fdur%3D278.778%26pl%3D17%26itag%3D22%26key%3Dyt6%26ip%3D126.2.187.172%26ms%3Dau%252Conr%26source%3Dyoutube%26mv%3Dm%26id%3Do-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR%26expire%3D1521061926%26mm%3D31%252C26%26mn%3Dsn-ogul7n7s%252Csn-3pm7snez%26mime%3Dvideo%252Fmp4%26lmt%3D1518448989107051%26ratebypass%3Dyes%26ei%3DxjupWtr7NIifqQG8vp_gBA%26fvip%3D1%26c%3DWEB%26mt%3D1521040211%26ipbits%3D0%26requiressl%3Dyes%26sparams%3Ddur%252Cei%252Cid%252Cinitcwndbps%252Cip%252Cipbits%252Citag%252Clmt%252Cmime%252Cmm%252Cmn%252Cms%252Cmv%252Cpl%252Cratebypass%252Crequiressl%252Csource%252Cexpire%26initcwndbps%3D772500\\u0026quality=hd720\\u0026sp=signature\\u0026s=55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77"
	expected, _ := newStream(
		map[string]string{
			"itag":    "22",
			"quality": "hd720",
			"s":       "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
			"sp":      "signature",
			"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=22&key=yt6&ip=126.2.187.172&ms=au,onr&source=youtube&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&expire=1521061926&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&mime=video/mp4&lmt=1518448989107051&ratebypass=yes&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&c=WEB&mt=1521040211&ipbits=0&requiressl=yes&sparams=dur,ei,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&initcwndbps=772500",
		}, nil, nil)
	if stream, err := inflateStream(rawStream, nil, nil); err != nil {
		t.Error("failed to inflate raw stream", err)
	} else if !reflect.DeepEqual(stream, expected) {
		t.Errorf("wrong inflation expected %v, got %v", expected, stream)
	}

	// case 2 (for age restricted video)
	rawStream = "itag=22&quality=hd720&type=video%2Fmp4%3B+codecs%3D%22avc1.64001F%2C+mp4a.40.2%22&url=https%3A%2F%2Fr1---sn-3pm7sn7r.googlevideo.com%2Fvideoplayback%3Fmt%3D1523922201%26mm%3D31%252C29%26itag%3D22%26requiressl%3Dyes%26ipbits%3D0%26ei%3DdTXVWoC1GNPWqAHZ2JOgCg%26sparams%3Ddur%252Cei%252Cid%252Cinitcwndbps%252Cip%252Cipbits%252Citag%252Clmt%252Cmime%252Cmm%252Cmn%252Cms%252Cmv%252Cpl%252Cratebypass%252Crequiressl%252Csource%252Cexpire%26dur%3D282.192%26ratebypass%3Dyes%26pl%3D17%26fvip%3D5%26ms%3Dau%252Crdu%26source%3Dyoutube%26mv%3Dm%26beids%3D%255B9466593%255D%26ip%3D126.225.83.8%26key%3Dyt6%26lmt%3D1507668791912906%26c%3DWEB%26initcwndbps%3D802500%26id%3Do-AB3avPHYxCHm1GOOiFZiFYPmAKUF-Vr1fUP6xCxawj_X%26mime%3Dvideo%252Fmp4%26signature%3D7C4CF38F629A9E79B784B7F0C5763BBE9EE7E6AB.DC7917FB4F2A270FA8B4FB075F94857ED11B310C%26expire%3D1523943893%26mn%3Dsn-3pm7sn7r%252Csn-3pm76n7s"
	expected, _ = newStream(
		map[string]string{
			"itag":    "22",
			"quality": "hd720",
			"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mt=1523922201&mm=31,29&itag=22&requiressl=yes&ipbits=0&ei=dTXVWoC1GNPWqAHZ2JOgCg&sparams=dur,ei,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&dur=282.192&ratebypass=yes&pl=17&fvip=5&ms=au,rdu&source=youtube&mv=m&beids=[9466593]&ip=126.225.83.8&key=yt6&lmt=1507668791912906&c=WEB&initcwndbps=802500&id=o-AB3avPHYxCHm1GOOiFZiFYPmAKUF-Vr1fUP6xCxawj_X&mime=video/mp4&signature=7C4CF38F629A9E79B784B7F0C5763BBE9EE7E6AB.DC7917FB4F2A270FA8B4FB075F94857ED11B310C&expire=1523943893&mn=sn-3pm7sn7r,sn-3pm76n7s",
		}, nil, nil)
	if stream, err := inflateStream(rawStream, nil, nil); err != nil {
		t.Error("failed to inflate raw stream", err)
	} else if !reflect.DeepEqual(stream, expected) {
		t.Errorf("wrong inflation")
	}

}
