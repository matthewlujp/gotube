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
	logger                 *errorLogger
	youtubeWatchURLPattern = regexp.MustCompile(`https?://www.youtube.com/watch\?v=(\w{11})`)
	jsURLRegex             = regexp.MustCompile(`.\"js\":\"(.+?)\"`)
	titleRegex             = regexp.MustCompile(`"title":"(.+?)","`)
	adaptiveFmtsRegex      = regexp.MustCompile(`"adaptive_fmts":"(.+?)"`)
	urlFmtsRegex           = regexp.MustCompile(`"url_encoded_fmt_stream_map":"(.+?)"`)
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
func (p *YoutubeDownloader) FetchStreamManifests() ([]*Stream, error) {
	res, errGet := p.client.Get(p.url)
	if errGet != nil {
		return nil, fmt.Errorf("request to %s failed, %s", p.url, errGet)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s got status %s", p.url, res.Status)
	}

	buf := new(bytes.Buffer)
	_, errRead := buf.ReadFrom(res.Body)
	if errRead != nil {
		return nil, fmt.Errorf("read html failed, %s", errRead)
	}
	videoInfo, errInfo := extractInfo(buf.Bytes())
	if errInfo != nil {
		return nil, errInfo
	}

	// download js script for signature decipher
	resJs, errGetJs := p.client.Get(videoInfo["jsURL"])
	var d *youtubeDecipherer
	if errGetJs != nil {
		// do not exit because stream urls may include deciphered signature from the first place
		logger.printf("failed to download js script from %s, %s", videoInfo["jsURL"], errGet)
	} else if resJs.StatusCode != http.StatusOK {
		logger.printf("js script download from %s got status %s", videoInfo["jsURL"], res.Status)
		resJs.Body.Close()
	} else {
		defer resJs.Body.Close()
		bufJs := new(bytes.Buffer)
		_, errRead = bufJs.ReadFrom(resJs.Body)
		if errRead != nil {
			logger.printf("failed in reading js script, %s", errRead)
		}
		var errNewDecipherer error
		d, errNewDecipherer = newDecipherer(bufJs.Bytes())
		if errNewDecipherer != nil {
			logger.printf("failed in building decipherer, %s", errNewDecipherer)
		}
	}

	stringStreams := strings.Split(videoInfo["streams"], ",")
	p.Streams = make([]*Stream, 0, len(stringStreams))
	for _, ss := range stringStreams {
		streamInfo, errInflate := inflateStringStream(ss)
		if errInflate != nil {
			logger.printf("%s", errInflate)
			continue
		}
		stream, errBuildStream := newStream(streamInfo, p.client, d)
		if errBuildStream != nil {
			logger.printf("%s", errBuildStream)
			continue
		}
		p.Streams = append(p.Streams, stream)
	}

	if len(p.Streams) < 1 {
		return p.Streams, errors.New("no stream is obtained")
	}
	return p.Streams, nil
}

func extractInfo(html []byte) (map[string]string, error) {
	info := make(map[string]string, 3)

	jsURL := jsURLRegex.FindSubmatch(html)
	if jsURL == nil || len(jsURL) < 2 {
		return nil, errors.New("failed to extract js url")
	}
	info["jsURL"] = "https://youtube.com" + strings.Replace(string(jsURL[1][:]), `\/`, "/", -1)

	title := titleRegex.FindSubmatch(html)
	if title == nil || len(title) < 2 {
		return nil, errors.New("failed to extract title")
	}
	info["title"] = string(title[1][:])

	streams := make([]string, 0, 2)
	adaptiveStreams := adaptiveFmtsRegex.FindSubmatch(html)
	if adaptiveStreams != nil && len(adaptiveStreams) >= 2 {
		streams = append(streams, string(adaptiveStreams[1][:]))
	}
	urlStreams := urlFmtsRegex.FindSubmatch(html)
	if urlStreams != nil && len(urlStreams) >= 2 {
		streams = append(streams, string(urlStreams[1][:]))
	}
	if len(streams) == 0 {
		return nil, errors.New("no streams extracted")
	}
	info["streams"] = strings.Join(streams, ",")

	return info, nil
}

func inflateStringStream(rawStream string) (map[string]string, error) {
	items := strings.Split(rawStream, "\\u0026")
	stream := make(map[string]string, len(items))
	for _, item := range items {
		vals := strings.Split(item, "=")
		unescaped, err := url.QueryUnescape(vals[1])
		if err != nil {
			return nil, err
		}
		stream[vals[0]] = unescaped
	}
	return stream, nil
}
