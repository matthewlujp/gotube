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
	ageRestrictedURLFmtsRegex      = regexp.MustCompile(`url_encoded_fmt_stream_map=(.+?)&`)
	stsRegex                       = regexp.MustCompile(`"sts"\s*:\s*(\d+)`)
	videoIDRegex                   = regexp.MustCompile(`watch\?v=([\w-]{11})`)
)

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

// FetchStreams build Stream instances based on information collected
// By using one of obtained Stream instances, video can be downloaded.
func (dl *YoutubeDownloader) FetchStreams() error {
	videoData, errExtractData := dl.extractData() // extract title, jsURL, and string form streams
	if errExtractData != nil {
		return errExtractData
	}
	dl.title = videoData["title"]

	// download js script and build a decipherer instance
	var dcph decipherer
	if js, err := dl.getResource(videoData["jsURL"]); err == nil {
		if d, err := newDecipherer(js); err == nil {
			dcph = d
		}
	}

	// create stream instances
	stringStreams := strings.Split(videoData["streams"], ",")
	dl.Streams = make([]*Stream, 0, len(stringStreams))
	for _, ss := range stringStreams {
		stream, errBuildStream := inflateStream(ss, dl.client, dcph)
		if errBuildStream != nil {
			logger.printf("%s", errBuildStream)
			continue
		}
		dl.Streams = append(dl.Streams, stream)
	}

	if len(dl.Streams) < 1 {
		return errors.New("no stream is obtained")
	}
	return nil
}

func (dl *YoutubeDownloader) getResource(url string) ([]byte, error) {
	res, errGet := dl.client.Get(url)
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
	return buf.Bytes(), nil
}

func (dl *YoutubeDownloader) extractData() (map[string]string, error) {
	html, errGetHTML := dl.getResource(dl.url)
	if errGetHTML != nil {
		return nil, errGetHTML
	}

	videoData := make(map[string]string)

	videoID, errVideoID := dl.extractVideoID()
	if errVideoID != nil {
		return nil, errVideoID
	}
	ageRestricted := isAgeRestricted(html)

	var embedHTML []byte
	var videoInfo []byte
	if ageRestricted {
		if html, err := dl.getResource(embedURL(videoID)); err != nil {
			return nil, err
		} else {
			embedHTML = html
		}

		if info, err := dl.getAuxiliaryInfo(embedHTML, videoID); err != nil {
			return nil, err
		} else {
			videoInfo = info
		}
	}

	videoData["title"] = extractTitle(html, embedHTML, ageRestricted)
	videoData["jsURL"] = extractJsURL(html, embedHTML, ageRestricted)
	streams, errExtractStreams := extractStreams(html, videoInfo, ageRestricted)
	if errExtractStreams != nil {
		return nil, errExtractStreams
	}
	videoData["streams"] = strings.Join(streams, ",")
	return videoData, nil
}

func (dl *YoutubeDownloader) extractVideoID() (string, error) {
	res := videoIDRegex.FindStringSubmatch(dl.url)
	if res == nil {
		return "", fmt.Errorf("no id found in %s", dl.url)
	}
	return res[1], nil
}

func (dl *YoutubeDownloader) getAuxiliaryInfo(embedHTML []byte, videoID string) ([]byte, error) {
	sts := stsRegex.FindSubmatch(embedHTML)
	if len(sts) < 2 {
		return nil, errors.New("failed to obtain sts for video info url")
	}
	videoInfoURL := auxiliaryInfoURL(videoID, string(sts[1][:]))
	videoInfo, err := dl.getResource(videoInfoURL)
	if err != nil {
		return nil, err
	}
	return videoInfo, nil
}

func embedURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
}

func auxiliaryInfoURL(videoID, sts string) string {
	return fmt.Sprintf(
		"https://youtube.com/get_video_info?video_id=%s&eurl=%s&sts=%s",
		videoID,
		url.QueryEscape(fmt.Sprintf("https://youtube.googleapis.com/v/%s", videoID)),
		sts,
	)
}

func extractTitle(html, embedHTML []byte, ageRestricted bool) string {
	var title [][]byte
	if ageRestricted {
		title = titleRegex.FindSubmatch(embedHTML)
	} else {
		title = titleRegex.FindSubmatch(html)
	}
	if title == nil {
		logger.print("no title is extracted\n")
		return ""
	}
	return string(title[1][:])
}

func extractJsURL(html, embedHTML []byte, ageRestricted bool) string {
	var jsURL [][]byte
	if ageRestricted {
		jsURL = ageRestrictedJsURLRegex.FindSubmatch(embedHTML)
	} else {
		jsURL = jsURLRegex.FindSubmatch(html)
	}
	if jsURL == nil {
		logger.print("no js url is extracted\n")
		return ""
	}
	return "https://youtube.com" + strings.Replace(string(jsURL[1][:]), `\/`, "/", -1)
}

func extractStreams(html, videoInfo []byte, ageRestricted bool) ([]string, error) {
	streams := make([]string, 0, 2)

	if ageRestricted {
		for _, rgx := range []*regexp.Regexp{ageRestrictedAdaptiveFmtsRegex, ageRestrictedURLFmtsRegex} {
			if strms := rgx.FindSubmatch(videoInfo); strms != nil {
				if decoded, err := url.QueryUnescape(string(strms[1][:])); err == nil {
					streams = append(streams, decoded)
				}
			}
		}
	} else {
		for _, rgx := range []*regexp.Regexp{adaptiveFmtsRegex, urlFmtsRegex} {
			if strms := rgx.FindSubmatch(html); strms != nil {
				streams = append(streams, string(strms[1][:]))
			}
		}
	}

	if streams == nil || len(streams) == 0 {
		return nil, errors.New("no streams found")
	}
	return streams, nil
}

func isAgeRestricted(html []byte) bool {
	return strings.Contains(string(html[:]), "og:restrictions:age")
}

func inflateStream(rawStream string, c client, d decipherer) (*Stream, error) {
	// split into key=val pairs
	var items []string
	if strings.Contains(rawStream, "\\u0026") {
		items = strings.Split(rawStream, "\\u0026")
	} else {
		items = strings.Split(rawStream, "&")
	}

	// create a map which is used for building a stream instance
	values := make(map[string]string, len(items))
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
		values[vals[0]] = unescaped
	}

	if stream, err := newStream(values, c, d); err != nil {
		return nil, err
	} else {
		return stream, nil
	}
}
