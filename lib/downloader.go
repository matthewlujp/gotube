package gotube

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	logger                         *errorLogger
	youtubeWatchURLPattern         = regexp.MustCompile(`https?://www.youtube.com/watch\?v=(\w{11})`)
	jsURLRegex                     = regexp.MustCompile(`.\"js\":\"(.+?)\"`)
	ageRestrictedJsURLRegex        = regexp.MustCompile(`;yt\.setConfig\(\{\'PLAYER_CONFIG\':\s*{.+?"js":"(.+?)"}.+?}(,\'EXPERIMENT_FLAGS\'|;)`)
	titleRegex                     = regexp.MustCompile(`"title":"(.+?)","`)
	adaptiveFmtsRegex              = regexp.MustCompile(`"adaptive_fmts":"(.+?)"`)
	urlFmtsRegex                   = regexp.MustCompile(`"url_encoded_fmt_stream_map":"(.+?)"`)
	ageRestrictedAdaptiveFmtsRegex = regexp.MustCompile(`adaptive_fmts=(.+?)&`)
	ageRestrictedUrlFmtsRegex      = regexp.MustCompile(`url_encoded_fmt_stream_map=(.+?)&`)
	stsRegex                       = regexp.MustCompile(`"sts"\s*:\s*(\d+)`)
	videoIDRegex                   = regexp.MustCompile(`v=(.+?)&`)
)

func init() {
	logger = newLogger(true)
}

// YoutubeDownloader collects information of a Youtube video and fetches streams of it.
type YoutubeDownloader struct {
	client  client
	Streams []*Stream // accessable
	url     string
	title   string
}

// NewDownloader returns a instance which implements YoutubeDownloader according to a given url
func NewDownloader(url string) (*YoutubeDownloader, error) {
	bURL := []byte(url)
	if youtubeWatchURLPattern.Match(bURL) {
		return &YoutubeDownloader{client: &youtubeClient{}, url: url}, nil
	}
	return nil, fmt.Errorf("unexpected URL format %s", url)
}

// FetchStreamManifests build Stream instances based on information collected
// By using one of obtained Stream instances, video can be downloaded.
func (dl *YoutubeDownloader) FetchStreamManifests() ([]*Stream, error) {
	res, errGet := dl.client.Get(dl.url)
	if errGet != nil {
		return nil, fmt.Errorf("request to %s failed, %s", dl.url, errGet)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s got status %s", dl.url, res.Status)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, fmt.Errorf("read html failed, %s", err)
	}
	html := buf.Bytes()

	var videoData map[string]string
	if restricted := isAgeRestricted(html); restricted {
		resEmb, errGetEmb := dl.client.Get("https://www.youtube.com/embed/6LZM3_wp2ps")
		if errGetEmb != nil {
			return nil, fmt.Errorf("request for embed html failed, %s", errGetEmb)
		}
		defer resEmb.Body.Close()
		bufEmb := new(bytes.Buffer)
		if _, err := bufEmb.ReadFrom(resEmb.Body); err != nil {
			return nil, fmt.Errorf("read embed html failed, %s", err)
		}
		embedHTML := bufEmb.Bytes()

		if data, err := extractInfo(embedHTML, restricted); err != nil {
			return nil, err
		} else {
			videoData = data
		}

		videoInfoURL := extractVideoInfoURL(embedHTML, dl.url)
		resVideoInfo, errGetVideoInfo := dl.client.Get(videoInfoURL)
		if errGetVideoInfo != nil {
			return nil, fmt.Errorf("request for video info failed, %s", errGetVideoInfo)
		}
		defer resVideoInfo.Body.Close()
		bufVideoInfo := new(bytes.Buffer)
		if _, err := bufVideoInfo.ReadFrom(resVideoInfo.Body); err != nil {
			return nil, fmt.Errorf("read video info failed, %s", err)
		}
		videoInfo := bufVideoInfo.Bytes()

		streams := extractStreamsFromVideoInfo(videoInfo)
		if len(streams) == 0 {
			return nil, errors.New("no streams extracted")
		}
		videoData["streams"] = strings.Join(streams, ",")
	} else {
		if info, err := extractInfo(html, restricted); err != nil {
			return nil, err
		} else {
			videoData = info
		}

		streams := extractStreamsFromHTML(html)
		if len(streams) == 0 {
			return nil, errors.New("no streams extracted")
		}
		videoData["streams"] = strings.Join(streams, ",")
	}

	// download js script for signature decipher
	resJs, errGetJs := dl.client.Get(videoData["jsURL"])
	var d *youtubeDecipherer
	if errGetJs != nil {
		// do not exit because stream urls may include deciphered signature from the first place
		logger.printf("failed to download js script from %s, %s", videoData["jsURL"], errGetJs)
	} else if resJs.StatusCode != http.StatusOK {
		logger.printf("js script download from %s got status %s", videoData["jsURL"], res.Status)
		resJs.Body.Close()
	} else {
		defer resJs.Body.Close()
		bufJs := new(bytes.Buffer)
		if _, err := bufJs.ReadFrom(resJs.Body); err != nil {
			logger.printf("failed in reading js script, %s", err)
		}
		var errNewDecipherer error
		d, errNewDecipherer = newDecipherer(bufJs.Bytes())
		if errNewDecipherer != nil {
			logger.printf("failed in building decipherer, %s", errNewDecipherer)
		}
	}

	stringStreams := strings.Split(videoData["streams"], ",")
	dl.Streams = make([]*Stream, 0, len(stringStreams))
	for _, ss := range stringStreams {
		streamInfo, errInflate := inflateStringStream(ss)
		if errInflate != nil {
			logger.printf("%s", errInflate)
			continue
		}
		stream, errBuildStream := newStream(streamInfo, dl.client, d)
		if errBuildStream != nil {
			logger.printf("%s", errBuildStream)
			continue
		}
		dl.Streams = append(dl.Streams, stream)
	}

	dl.title = videoData["title"]

	if len(dl.Streams) < 1 {
		return dl.Streams, errors.New("no stream is obtained")
	}
	return dl.Streams, nil
}

func isAgeRestricted(html []byte) bool {
	return strings.Contains(string(html[:]), "og:restrictions:age")
}

func extractInfo(html []byte, ageRestricted bool) (map[string]string, error) {
	info := make(map[string]string, 3)

	var jsURL [][]byte
	if ageRestricted {
		jsURL = ageRestrictedJsURLRegex.FindSubmatch(html)
		if jsURL == nil || len(jsURL) < 2 {
			return nil, errors.New("failed to extract js url")
		}
	} else {
		jsURL = jsURLRegex.FindSubmatch(html)
		if jsURL == nil || len(jsURL) < 2 {
			return nil, errors.New("failed to extract js url")
		}
	}
	info["jsURL"] = "https://youtube.com" + strings.Replace(string(jsURL[1][:]), `\/`, "/", -1)

	title := titleRegex.FindSubmatch(html)
	if title == nil || len(title) < 2 {
		return nil, errors.New("failed to extract title")
	}
	info["title"] = string(title[1][:])
	return info, nil
}

func extractStreamsFromHTML(html []byte) []string {
	streams := make([]string, 0, 2)
	adaptiveStreams := adaptiveFmtsRegex.FindSubmatch(html)
	if adaptiveStreams != nil && len(adaptiveStreams) >= 2 {
		streams = append(streams, string(adaptiveStreams[1][:]))
	}
	urlStreams := urlFmtsRegex.FindSubmatch(html)
	if urlStreams != nil && len(urlStreams) >= 2 {
		streams = append(streams, string(urlStreams[1][:]))
	}
	return streams
}

func extractStreamsFromVideoInfo(strInfo []byte) []string {
	streams := make([]string, 0, 2)
	adaptiveStreams := ageRestrictedAdaptiveFmtsRegex.FindSubmatch(strInfo)
	if adaptiveStreams != nil && len(adaptiveStreams) >= 2 {
		if decoded, err := url.QueryUnescape(string(adaptiveStreams[1][:])); err == nil {
			streams = append(streams, decoded)
		}
	}
	urlStreams := ageRestrictedUrlFmtsRegex.FindSubmatch(strInfo)
	if urlStreams != nil && len(urlStreams) >= 2 {
		if decoded, err := url.QueryUnescape(string(urlStreams[1][:])); err == nil {
			streams = append(streams, decoded)
		}
	}
	// fmt.Println(streams)
	return streams
}

func inflateStringStream(rawStream string) (map[string]string, error) {
	var items []string
	if strings.Contains(rawStream, "\\u0026") {
		items = strings.Split(rawStream, "\\u0026")
	} else {
		items = strings.Split(rawStream, "&")
	}
	stream := make(map[string]string, len(items))
	for _, item := range items {
		vals := strings.Split(item, "=")
		unescaped, err := url.QueryUnescape(vals[1])
		if err != nil {
			return nil, err
		}
		unescaped, err = url.QueryUnescape(unescaped) // apply twice to eliminate %
		if err != nil {
			return nil, err
		}
		stream[vals[0]] = unescaped
	}
	return stream, nil
}

func extractVideoInfoURL(embedHTML []byte, videoURL string) string {
	sts := stsRegex.FindSubmatch(embedHTML)
	if len(sts) < 2 {
		return ""
	}
	videoID := videoIDRegex.FindStringSubmatch(videoURL)
	fmt.Println(videoURL, videoID)
	if len(videoID) < 2 {
		return ""
	}
	return fmt.Sprintf(
		"https://youtube.com/get_video_info?video_id=%s&eurl=%s&sts=%s",
		videoID[1],
		url.QueryEscape(fmt.Sprintf("https://youtube.googleapis.com/v/%s", videoID[1])),
		sts[1],
	)
}
