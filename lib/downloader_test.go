package gotube

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	mock "github.com/matthewlujp/gotube/lib/mocks"
)

var (
	validURL                  = "https://www.youtube.com/watch?v=iEPTlhBmwRg"
	dummyURL                  = "https://www.mytube.com/watch?v=iEPTlhBmwRg"
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
		c.EXPECT().Get(jsURL).Return(getMockScript()),
	)

	downloader := YoutubeDownloader{
		url:    validURL,
		client: c,
	}

	streams, errFetch := downloader.FetchStreamManifests()
	if errFetch != nil {
		t.Errorf("error while fetching stream manifests, %s", errFetch)
	}
	if len(streams) != streamNumbers {
		t.Errorf("got %d streams, %d expected", len(streams), streamNumbers)
	}
}

func TestFetchStreamManifestsForAgeRestricted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)

	gomock.InOrder(
		c.EXPECT().Get(ageRestrictedURL).Return(getMockAgeRestrictedPage()),
		c.EXPECT().Get(ageRestrictedEmbedURL).Return(getContent(mockAgeRestrictedEmbedPagePath)),
		c.EXPECT().Get(ageRestrictedVideoInfoURL).Return(getContent(mockAgeRestrictedVideoInfoPath)),
		c.EXPECT().Get(ageRestrictedJsURL).Return(getMockAgeRestrictedScript()),
	)

	downloader := YoutubeDownloader{
		url:    ageRestrictedURL,
		client: c,
	}

	streams, errFetch := downloader.FetchStreamManifests()
	if errFetch != nil {
		t.Errorf("error while fetching stream manifests, %s", errFetch)
	}
	if len(streams) != restrictedStreamNumbers {
		t.Errorf("got %d streams, %d expected", len(streams), restrictedStreamNumbers)
	}
}

func TestExtractInfo(t *testing.T) {
	html, err := readCompressedFile(mockPagePath)
	if err != nil {
		t.Fatal(err)
	}

	extractData, err := extractInfo(html, false)
	if err != nil {
		t.Errorf("extract video information failed, %s", err)
	}

	// jsURL
	if extractedJsURL, ok := extractData["jsURL"]; !ok {
		t.Error("failed to extract jsURL")
	} else if extractedJsURL != jsURL {
		t.Errorf("wrong js url, got %s, expected %s", extractedJsURL, jsURL)
	}

	// title
	if extractedTitle, ok := extractData["title"]; !ok {
		t.Error("failed to extract title")
	} else if extractedTitle != title {
		t.Errorf("wrong title, got %s, expected %s", extractedTitle, title)
	}

	// ToDo(matthewlujp): move to TestFetch...
	// // streams
	// rawURLStream, errURLStream := getStrStream(rawURLEncodedStreamPath)
	// // log.Printf("url: %s\n", rawURLStream)
	// if errURLStream != nil {
	// 	t.Fatal(errURLStream)
	// }
	// rawAdaptiveStream, errAdaptiveStream := getStrStream(rawAgeRestrictedAdaptiveStreamPath)
	// // log.Printf("adaptive: %s\n", rawAdaptiveStream)
	// if errAdaptiveStream != nil {
	// 	t.Fatal(errAdaptiveStream)
	// }
	// if extractedRawStreams, ok := extractData["streams"]; !ok {
	// 	t.Error("failed to extract streams")
	// } else if extractedRawStreams != fmt.Sprintf("%s,%s", rawURLStream, rawAdaptiveStream) && extractedRawStreams != fmt.Sprintf("%s,%s", rawAdaptiveStream, rawURLStream) {
	// 	t.Error("wrong raw streams")
	// }
}
func TestExtractInfoAgeRestricted(t *testing.T) {
	html, err := readCompressedFile(mockAgeRestrictedEmbedPagePath)
	if err != nil {
		t.Fatal(err)
	}

	extractData, err := extractInfo(html, true)
	if err != nil {
		t.Errorf("extract video information failed, %s", err)
	}

	// jsURL
	if extractedJsURL, ok := extractData["jsURL"]; !ok {
		t.Error("failed to extract jsURL")
	} else if extractedJsURL != ageRestrictedJsURL {
		t.Errorf("wrong js url, got %s, expected %s", extractedJsURL, jsURL)
	}

	// title
	if extractedTitle, ok := extractData["title"]; !ok {
		t.Error("failed to extract title")
	} else if extractedTitle != ageRestrictedTitle {
		t.Errorf("wrong title, got %s, expected %s", extractedTitle, title)
	}

	// ToDo(matthewlujp): move to TestFetch...AgeRestricted
	// // streams
	// rawURLStream, errURLStream := getStrStream(rawAgeRestrictedURLEncodedStreamPath)
	// // log.Printf("url: %s\n", rawURLStream)
	// if errURLStream != nil {
	// 	t.Fatal(errURLStream)
	// }
	// rawAdaptiveStream, errAdaptiveStream := getStrStream(rawAgeRestrictedAdaptiveStreamPath)
	// // log.Printf("adaptive: %s\n", rawAdaptiveStream)
	// if errAdaptiveStream != nil {
	// 	t.Fatal(errAdaptiveStream)
	// }
	// if extractedRawStreams, ok := extractData["streams"]; !ok {
	// 	t.Error("failed to extract streams")
	// } else if extractedRawStreams != fmt.Sprintf("%s,%s", rawURLStream, rawAdaptiveStream) && extractedRawStreams != fmt.Sprintf("%s,%s", rawAdaptiveStream, rawURLStream) {
	// 	t.Error("wrong raw streams")
	// }
}

func TestInflateStringStream(t *testing.T) {
	// case 1 (for noramal video)
	rawStream := "type=video%2Fmp4%3B+codecs%3D%22avc1.64001F%2C+mp4a.40.2%22\\u0026itag=22\\u0026url=https%3A%2F%2Fr1---sn-ogul7n7s.googlevideo.com%2Fvideoplayback%3Fdur%3D278.778%26pl%3D17%26itag%3D22%26key%3Dyt6%26ip%3D126.2.187.172%26ms%3Dau%252Conr%26source%3Dyoutube%26mv%3Dm%26id%3Do-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR%26expire%3D1521061926%26mm%3D31%252C26%26mn%3Dsn-ogul7n7s%252Csn-3pm7snez%26mime%3Dvideo%252Fmp4%26lmt%3D1518448989107051%26ratebypass%3Dyes%26ei%3DxjupWtr7NIifqQG8vp_gBA%26fvip%3D1%26c%3DWEB%26mt%3D1521040211%26ipbits%3D0%26requiressl%3Dyes%26sparams%3Ddur%252Cei%252Cid%252Cinitcwndbps%252Cip%252Cipbits%252Citag%252Clmt%252Cmime%252Cmm%252Cmn%252Cms%252Cmv%252Cpl%252Cratebypass%252Crequiressl%252Csource%252Cexpire%26initcwndbps%3D772500\\u0026quality=hd720\\u0026sp=signature\\u0026s=55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77"
	expected := map[string]string{
		"itag":    "22",
		"quality": "hd720",
		"s":       "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
		"sp":      "signature",
		"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
		"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=22&key=yt6&ip=126.2.187.172&ms=au%2Conr&source=youtube&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&expire=1521061926&mm=31%2C26&mn=sn-ogul7n7s%2Csn-3pm7snez&mime=video%2Fmp4&lmt=1518448989107051&ratebypass=yes&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&c=WEB&mt=1521040211&ipbits=0&requiressl=yes&sparams=dur%2Cei%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Cratebypass%2Crequiressl%2Csource%2Cexpire&initcwndbps=772500",
	}
	if stream, err := inflateStringStream(rawStream); err != nil {
		t.Error("failed to inflate raw stream", err)
	} else if !reflect.DeepEqual(stream, expected) {
		t.Errorf("wrong inflation")
		for k, v := range expected {
			if val, ok := stream[k]; !ok {
				t.Errorf("%s not in inflated map", k)
			} else if val != v {
				t.Errorf("%s in stream %s, while in expected %s", k, val, v)
			}
		}
	}

	// case 2 (for age restricted video)
	rawStream = "itag=22&quality=hd720&type=video%2Fmp4%3B+codecs%3D%22avc1.64001F%2C+mp4a.40.2%22&url=https%3A%2F%2Fr1---sn-3pm7sn7r.googlevideo.com%2Fvideoplayback%3Fmt%3D1523922201%26mm%3D31%252C29%26itag%3D22%26requiressl%3Dyes%26ipbits%3D0%26ei%3DdTXVWoC1GNPWqAHZ2JOgCg%26sparams%3Ddur%252Cei%252Cid%252Cinitcwndbps%252Cip%252Cipbits%252Citag%252Clmt%252Cmime%252Cmm%252Cmn%252Cms%252Cmv%252Cpl%252Cratebypass%252Crequiressl%252Csource%252Cexpire%26dur%3D282.192%26ratebypass%3Dyes%26pl%3D17%26fvip%3D5%26ms%3Dau%252Crdu%26source%3Dyoutube%26mv%3Dm%26beids%3D%255B9466593%255D%26ip%3D126.225.83.8%26key%3Dyt6%26lmt%3D1507668791912906%26c%3DWEB%26initcwndbps%3D802500%26id%3Do-AB3avPHYxCHm1GOOiFZiFYPmAKUF-Vr1fUP6xCxawj_X%26mime%3Dvideo%252Fmp4%26signature%3D7C4CF38F629A9E79B784B7F0C5763BBE9EE7E6AB.DC7917FB4F2A270FA8B4FB075F94857ED11B310C%26expire%3D1523943893%26mn%3Dsn-3pm7sn7r%252Csn-3pm76n7s"
	expected = map[string]string{
		"itag":    "22",
		"quality": "hd720",
		"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
		"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mt=1523922201&mm=31%2C29&itag=22&requiressl=yes&ipbits=0&ei=dTXVWoC1GNPWqAHZ2JOgCg&sparams=dur%2Cei%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Cratebypass%2Crequiressl%2Csource%2Cexpire&dur=282.192&ratebypass=yes&pl=17&fvip=5&ms=au%2Crdu&source=youtube&mv=m&beids=%5B9466593%5D&ip=126.225.83.8&key=yt6&lmt=1507668791912906&c=WEB&initcwndbps=802500&id=o-AB3avPHYxCHm1GOOiFZiFYPmAKUF-Vr1fUP6xCxawj_X&mime=video%2Fmp4&signature=7C4CF38F629A9E79B784B7F0C5763BBE9EE7E6AB.DC7917FB4F2A270FA8B4FB075F94857ED11B310C&expire=1523943893&mn=sn-3pm7sn7r%2Csn-3pm76n7s",
	}
	if stream, err := inflateStringStream(rawStream); err != nil {
		t.Error("failed to inflate raw stream", err)
	} else if !reflect.DeepEqual(stream, expected) {
		t.Errorf("wrong inflation")
		for k, v := range expected {
			if val, ok := stream[k]; !ok {
				t.Errorf("%s not in inflated map", k)
			} else if val != v {
				t.Errorf("%s in stream %s, while in expected %s", k, val, v)
			}
		}
	}

}
