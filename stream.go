package gotube

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	maxSimultaneousRequests = 20
)

var (
	mimeTypeCodecRegex = regexp.MustCompile(`([a-z]+?)/([a-z0-9]+?);\s*codecs=\"([\w\s.,]+?)\"`)
)

// Stream represents a video data of a specific format.
// This structure is responsible for downloading video.
type Stream struct {
	itag         int
	Abr          string
	Fps          string
	Resolution   string
	MediaType    string // video, audio
	Quality      string // hd720
	Format       string // mp4
	VideoCodec   string
	AudioCodec   string
	Is3D         bool
	IsLive       bool
	Duration     time.Duration
	signature    string
	url          string
	downloadURL  string
	buildURLOnce sync.Once
	client       client
	decipherer   decipherer
}

// Download returns a byte slice of video content
func (s *Stream) Download() ([]byte, error) {
	downloadURL, errURLBuild := s.getDownloadURL()
	if errURLBuild != nil {
		return nil, errURLBuild
	}
	logger.printf("download url prepared: %s", downloadURL)
	return s.download(downloadURL)
}

// ParallelDownload returns a byte slice of video content
// It conducts download in parallel
// A video is separated every 20 seconds and they are requested in parallel
// Bytes for 20 seconds are calculated based on video duration and the bytes length.
// Bytes length are checked before the get request by sending head rewquest.
func (s *Stream) ParallelDownload() ([]byte, error) {
	ranges, errRanges := s.byteRanges(time.Second * 20) // chunkSize of 20 seconds
	if errRanges != nil {
		return nil, errRanges
	}

	collectedData := make([][]byte, len(ranges)-1)

	// decipher and get basic url
	downloadURL, errURLBuild := s.getDownloadURL()
	if errURLBuild != nil {
		return nil, errURLBuild
	}
	logger.printf("download url prepared: %s", downloadURL)

	// parallel download
	eg := errgroup.Group{}
	for i := range ranges[:len(ranges)-1] {
		idx := i
		eg.Go(func() error {
			data, err := s.download(fmt.Sprintf("%s&range=%d-%d", downloadURL, ranges[idx], ranges[idx+1]-1))
			if err != nil {
				return err
			}

			collectedData[idx] = data
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	logger.print("download completed")

	// concatenate downloaded data
	for _, data := range collectedData[1:] {
		collectedData[0] = append(collectedData[0], data...)
	}
	return collectedData[0], nil
}

// SequentialChunkDownload returns a channel from which you can receive data chunks successively.
// Download is conducted in a goroutine (parallel) and []byte is sent through the returned channel in order.
//
// A video is separated every {chunkDuration} seconds and they are requested in parallel
// Bytes for {chunkDuration} seconds are calculated based on video duration and the bytes length.
// Bytes length are checked before the get request by sending head rewquest.
// TODO: return a channel for error as well
// TODO: receive signal and stop goroutines
func (s *Stream) SequentialChunkDownload(chunkDuration time.Duration) (<-chan []byte, error) {
	ranges, errRanges := s.byteRanges(chunkDuration) // byte size of chunkDuration
	if errRanges != nil {
		return nil, errRanges
	}

	// decipher and get basic url
	downloadURL, errURLBuild := s.getDownloadURL()
	if errURLBuild != nil {
		return nil, errURLBuild
	}
	logger.printf("download url prepared: %s", downloadURL)

	// create a slice of channels to notify completion of data fetch by sending empty struct
	doneChans := []chan struct{}{}
	for i := 0; i < len(ranges)-1; i++ {
		doneChans = append(doneChans, make(chan struct{}))
	}
	// slice of []byte to save arriced data
	collectedData := make([][]byte, len(ranges)-1)

	// create another data channel for output
	outputChan := make(chan []byte)
	// reorder arrived data and resend to outputChan in a goroutine
	go func() {
		for i, ch := range doneChans {
			<-ch // wait data fetch completion notification
			outputChan <- collectedData[i]
		}
		close(outputChan)
	}()

	// use semaphore to limit simultaneous request to YouTube
	semaphore := make(chan struct{}, maxSimultaneousRequests)

	// parallel download
	// TODO: use worker & dispatcher model to reuse goroutines
	go func() {
		for i := range ranges[:len(ranges)-1] {
			semaphore <- struct{}{} // block if requests count exceeds the limit
			idx := i
			go func() {
				defer func() {
					<-semaphore // release one slot
				}()
				var errDL error
				collectedData[idx], errDL = s.download(fmt.Sprintf("%s&range=%d-%d", downloadURL, ranges[idx], ranges[idx+1]-1))
				if errDL != nil {
					// TODO: retry 3 times
					if strings.Contains(errDL.Error(), "operation timed out") {
						logger.printf("range %d-%d, timeout -> retry", ranges[idx], ranges[idx+1]-1)
						// retry for timeout
						collectedData[idx], errDL = s.download(fmt.Sprintf("%s&range=%d-%d", downloadURL, ranges[idx], ranges[idx+1]-1))
						if errDL != nil {
							logger.printf("range %d-%d, retry failed, %s", ranges[idx], ranges[idx+1]-1, errDL)
						}
					} else {
						logger.printf("range %d-%d, %s", ranges[idx], ranges[idx+1]-1, errDL)
					}
				}
				doneChans[idx] <- struct{}{}
			}()
		}
	}()

	// a caller of this method receives data via outputChan
	return outputChan, nil
}

// download get resource and return byte slice
func (s *Stream) download(url string) ([]byte, error) {
	res, errGet := s.client.Get(url)
	if errGet != nil {
		return nil, errGet
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get request to %s got status %s", url, res.Status)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, fmt.Errorf("failed to read content downloaded from %s, %s", url, err)
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

// byteRanges returns a slice of indexes that split the video data evenly except for the last chunk.
// One chunk is less or equivalent to a given duration.
// [0, 1*chunkSize, 2*chunkSize, ..., size], is used to designate start and end of a stream
func (s *Stream) byteRanges(duration time.Duration) ([]int, error) {
	totalSize, err := s.GetSize()
	if err != nil {
		return nil, fmt.Errorf("failed to split video data, %s", err)
	}

	chunkSize := s.bytesForDuration(totalSize, duration)
	ranges := make([]int, 0, totalSize/chunkSize+1)
	for i := 0; i*chunkSize < totalSize; i++ {
		ranges = append(ranges, i*chunkSize)
	}
	ranges = append(ranges, totalSize)
	return ranges, nil
}

// bytesForDuration estimates byte size for a given duration (seconds)
func (s *Stream) bytesForDuration(dataSize int, fetchDuration time.Duration) int {
	if s.Duration == 0 {
		return dataSize
	}
	return int(math.Floor(float64(dataSize) / float64(s.Duration) * float64(fetchDuration)))
}

// GetSize returns content size of this stream
func (s *Stream) GetSize() (int, error) {
	sURL, errURL := s.getDownloadURL()
	if errURL != nil {
		return -1, errURL
	}
	res, err := s.client.Head(sURL)
	if err != nil {
		return -1, err
	}
	defer func() {
		if res.Body != nil {
			res.Body.Close()
		}
	}()

	if res.Header.Get("Content-Length") == "" {
		return -1, errors.New("GetSize failed, header does not contain Content-Length")
	}
	size, errConvert := strconv.Atoi(res.Header.Get("Content-Length"))
	if errConvert != nil {
		return -1, fmt.Errorf("GetSize failed, invalid size %s, %s", res.Header.Get("Content-Length"), errConvert)
	}
	return size, nil
}

// newStream returns a Stream instance
// The first argument is a map[string]string whose keys and values  are
// url:
// s: signature
// quality: hd720, medium, etc.
// type: "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"", "audio/webm; codecs=\"opus\""
// itag:
// duration: video duration in seconds
func newStream(streamInfo map[string]string, c client, d decipherer) (*Stream, error) {
	s := Stream{}

	streamURL, urlOK := streamInfo["url"]
	if !urlOK {
		return nil, errors.New("no stream url found")
	}
	s.url = streamURL

	if v, ok := streamInfo["s"]; ok {
		s.signature = v
	}

	if v, ok := streamInfo["quality"]; ok {
		s.Quality = v
	} else if v, ok = streamInfo["quality_label"]; ok {
		s.Quality = v
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
	}

	if duration, ok := streamInfo["duration"]; ok {
		if seconds, err := strconv.Atoi(duration); err == nil {
			s.Duration = time.Duration(seconds) * time.Second
		}
	}

	s.client = c
	s.decipherer = d
	return &s, nil
}

// getDownloadURL returns a url with a deciphered signature added
// Building url is only conducted once.
func (s *Stream) getDownloadURL() (string, error) {
	// decipher signature and build download url only once
	s.buildURLOnce.Do(
		func() {
			if strings.Contains(s.url, "&signature=") {
				// signature has already been included
				s.downloadURL = s.url
				return
			}

			if s.signature != "" && s.decipherer != nil {
				if decipheredSig, err := s.decipherer.Decipher(s.signature); err == nil {
					s.downloadURL = fmt.Sprintf("%s&signature=%s", s.url, decipheredSig)
				}
			}
		},
	)

	if s.downloadURL == "" {
		return "", errors.New("failed to decipher signature")
	}
	return s.downloadURL, nil
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
		s.url == other.url &&
		s.Duration == other.Duration
}
