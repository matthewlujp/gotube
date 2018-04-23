package gotube

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	mimeTypeCodecRegex = regexp.MustCompile(`([a-z]+?)/([a-z0-9]+?);\s*codecs=\"([\w\s.,]+?)\"`)
)

type Stream struct {
	itag       int
	Abr        string
	Fps        string
	Resolution string
	MediaType  string // video, audio
	Quality    string // hd720
	Format     string // mp4
	VideoCodec string
	AudioCodec string
	Is3D       bool
	IsLive     bool
	signature  string
	url        string
	client     client
	decipherer decipherer
}

// Download returns a reader for downloaded video stream
func (s *Stream) Download() ([]byte, error) {
	downloadURL, errURLBuild := s.buildDownloadURL()
	if errURLBuild != nil {
		return nil, errURLBuild
	}
	logger.printf("download url prepared: %s", downloadURL)

	res, err := s.client.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s got status %d", downloadURL, res.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, fmt.Errorf("error while reading video content, %s", err)
	}
	return buf.Bytes(), nil
}

func (s *Stream) String() string {
	base := fmt.Sprintf("Stream<MediaType:%s Quality:%s Format:%s Resolution:%s", s.MediaType, s.Quality, s.Format, s.Resolution)
	if s.Is3D {
		base += " 3D"
	}
	if s.IsLive {
		base += " Live"
	}
	return base + ">"
}

func newStream(streamInfo map[string]string, c client, d decipherer) (*Stream, error) {
	s := Stream{}

	streamURL, urlOK := streamInfo["url"]
	if !urlOK {
		return nil, errors.New("no stream url found")
	}
	s.url = streamURL

	if v, ok := streamInfo["s"]; ok {
		s.signature = v
	} else {
		logger.print("no signature found")
	}

	if v, ok := streamInfo["quality"]; ok {
		s.Quality = v
	} else if v, ok = streamInfo["quality_label"]; ok {
		s.Quality = v
	} else {
		logger.print("no quality found")
	}

	if v, ok := streamInfo["type"]; ok {
		typeAndCodecs := mimeTypeCodecRegex.FindStringSubmatch(v)
		if typeAndCodecs == nil || len(typeAndCodecs) < 4 {
			logger.printf("failed extract video/audio info, such as codecs, %v is extracted from %s", typeAndCodecs, v)
		} else {
			s.MediaType = typeAndCodecs[1]
			s.Format = typeAndCodecs[2]
			codecs := strings.Split(typeAndCodecs[3], ", ")
			if len(codecs) < 2 {
				if s.MediaType == "audio" {
					s.VideoCodec = ""
					s.AudioCodec = codecs[0]
				} else {
					s.VideoCodec = codecs[0]
					s.AudioCodec = ""
				}
			} else {
				s.VideoCodec = codecs[0]
				s.AudioCodec = codecs[1]
			}
		}
	} else {
		logger.print("no type found")
	}

	if v, ok := streamInfo["itag"]; ok {
		var errConvert error
		s.itag, errConvert = strconv.Atoi(v)
		if errConvert != nil {
			logger.printf("failed to convert itag from string to int, %s", errConvert)
		} else {
			fp := getFormatProfile(s.itag)
			if fp.is60fps {
				s.Fps = "60"
			} else {
				s.Fps = "30"
			}
			s.Abr = fp.bitrate
			s.Resolution = fp.resolution
			s.Is3D = fp.is3D
			s.IsLive = fp.isLive
		}
	} else {
		logger.print("no itag found")
	}

	s.client = c
	s.decipherer = d
	return &s, nil
}

func (s *Stream) buildDownloadURL() (string, error) {
	if strings.Contains(s.url, "&signature=") {
		return s.url, nil
	}
	if s.signature == "" {
		return "", errors.New("stream url does not contain a signature and signature is not in this Stream structure either")
	} else if s.decipherer == nil {
		return "", errors.New("stream url does not contain a signature and decipherer is missin (nil) either")
	}
	decipheredSignature, err := s.decipherer.Decipher(s.signature)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s&signature=%s", s.url, decipheredSignature), nil
}

func (s *Stream) equal(other *Stream) bool {
	return s.itag == other.itag &&
		s.Abr == other.Abr &&
		s.Fps == other.Fps &&
		s.Resolution == other.Resolution &&
		s.MediaType == other.MediaType &&
		s.Quality == other.Quality &&
		s.Format == other.Format &&
		s.VideoCodec == other.VideoCodec &&
		s.AudioCodec == other.AudioCodec &&
		s.Is3D == other.Is3D &&
		s.IsLive == other.IsLive &&
		s.signature == other.signature &&
		s.url == other.url

}
