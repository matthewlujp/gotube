package gotube

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock "github.com/matthewlujp/gotube/mocks"
)

func TestNewStream(t *testing.T) {
	invalidStreamInfo := make(map[string]string)
	stream, err := newStream(invalidStreamInfo, nil, nil)
	if err == nil {
		t.Errorf("invalid streamInfo %v should be rejected, but didn't", invalidStreamInfo)
	}

	streamInfo := map[string]string{
		"itag":      "22",
		"quality":   "hd720",
		"s":         "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
		"sp":        "signature",
		"type":      "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
		"signature": "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
		"url":       "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=22&key=yt6&ip=126.2.187.172&ms=au%2Conr&source=youtube&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&expire=1521061926&mm=31%2C26&mn=sn-ogul7n7s%2Csn-3pm7snez&mime=video%2Fmp4&lmt=1518448989107051&ratebypass=yes&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&c=WEB&mt=1521040211&ipbits=0&requiressl=yes&sparams=dur%2Cei%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Cratebypass%2Crequiressl%2Csource%2Cexpire&initcwndbps=772500",
	}
	stream, err = newStream(streamInfo, nil, nil)
	expected := Stream{
		itag:       22,
		Abr:        "192kbps",
		Fps:        "30",
		Resolution: "720p",
		Quality:    "hd720",
		MediaType:  "video",
		Format:     "mp4",
		VideoCodec: "avc1.64001F",
		AudioCodec: "mp4a.40.2",
		Is3D:       false,
		IsLive:     false,
		signature:  "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
		url:        "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=22&key=yt6&ip=126.2.187.172&ms=au%2Conr&source=youtube&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&expire=1521061926&mm=31%2C26&mn=sn-ogul7n7s%2Csn-3pm7snez&mime=video%2Fmp4&lmt=1518448989107051&ratebypass=yes&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&c=WEB&mt=1521040211&ipbits=0&requiressl=yes&sparams=dur%2Cei%2Cid%2Cinitcwndbps%2Cip%2Cipbits%2Citag%2Clmt%2Cmime%2Cmm%2Cmn%2Cms%2Cmv%2Cpl%2Cratebypass%2Crequiressl%2Csource%2Cexpire&initcwndbps=772500",
		client:     nil,
		decipherer: nil,
	}
	if err != nil {
		t.Errorf("failed to build stream, %s", err)
	} else if !reflect.DeepEqual(*stream, expected) {
		t.Errorf("wrong stream is built,\ngot: %v,\nexpected: %v", *stream, expected)
	}
}

func TestDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)
	d := mock.NewMockdecipherer(ctrl)

	stream := Stream{
		signature:  "hoge",
		url:        "https://foobar?itag=22",
		client:     c,
		decipherer: d,
	}

	content := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	gomock.InOrder(
		d.EXPECT().Decipher("hoge").Return("geho", nil),
		c.EXPECT().Get("https://foobar?itag=22&signature=geho").Return(
			&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(content)),
			},
			nil,
		),
	)

	data, errDownload := stream.Download()
	if errDownload != nil {
		t.Fatalf("stream donwload failed, %s", errDownload)
	}
	if bytes.Compare(data, content) != 0 {
		t.Errorf("got stream data %v, expected %v", data, content)
	}
}

func TestParallelDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)
	d := mock.NewMockdecipherer(ctrl)

	stream := Stream{
		signature:  "hoge",
		url:        "https://foobar?itag=22",
		duration:   time.Second * time.Duration(20*7.5),
		client:     c,
		decipherer: d,
	}

	content := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
	}
	// 13 bytes for 20 * 7.5 = 150 seconds
	// Since one request is for 20 seconds, 8 requests are supposed to make.
	// One request is floor(17 / 8) = 2 bytes
	// In order to keep request chunk sie 20 seconds, additional request is conducted.
	d.EXPECT().Decipher("hoge").Return("geho", nil)
	header := make(http.Header)
	header.Set("Content-Length", "17")
	c.EXPECT().Head("https://foobar?itag=22&signature=geho").Return(
		&http.Response{
			Header:     header,
			StatusCode: 200,
		},
		nil,
	)
	for d := 0; d < 8; {
		requestURL := fmt.Sprintf("https://foobar?itag=22&signature=geho&range=%d-%d", 2*d, 2*d+1)
		c.EXPECT().Get(requestURL).Return(
			&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(content[2*d : 2*d+2])),
			},
			nil,
		)
		d++
	}
	c.EXPECT().Get("https://foobar?itag=22&signature=geho&range=16-16").Return(
		&http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(content[16:17])),
		},
		nil,
	)

	data, errDownload := stream.ParallelDownload()
	if errDownload != nil {
		t.Fatalf("stream donwload failed, %s", errDownload)
	}
	if bytes.Compare(data, content) != 0 {
		t.Errorf("got stream data %v, expected %v", data, content)
	}
}
func TestSequentialChunkDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)
	d := mock.NewMockdecipherer(ctrl)

	stream := Stream{
		signature:  "hoge",
		url:        "https://foobar?itag=22",
		duration:   time.Second * time.Duration(20*7.5),
		client:     c,
		decipherer: d,
	}

	content := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
	}
	// 13 bytes for 20 * 7.5 = 150 seconds
	// Since one request is for 20 seconds, 8 requests are supposed to make.
	// One request is floor(17 / 8) = 2 bytes
	// In order to keep request chunk sie 20 seconds, additional request is conducted.
	d.EXPECT().Decipher("hoge").Return("geho", nil)
	header := make(http.Header)
	header.Set("Content-Length", "17")
	exploreCall := c.EXPECT().Head("https://foobar?itag=22&signature=geho").Return(
		&http.Response{
			Header:     header,
			StatusCode: 200,
		},
		nil,
	)

	for d := 0; d < 8; {
		requestURL := fmt.Sprintf("https://foobar?itag=22&signature=geho&range=%d-%d", 2*d, 2*d+1)
		c.EXPECT().Get(requestURL).Return(
			&http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(content[2*d : 2*d+2])),
			},
			nil,
		).After(exploreCall)
		d++
	}
	c.EXPECT().Get("https://foobar?itag=22&signature=geho&range=16-16").Return(
		&http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(content[16:17])),
		},
		nil,
	).After(exploreCall)

	log.Printf("start chunk download\n")
	dataChan, errDownload := stream.SequentialChunkDownload(20 * time.Second)
	if errDownload != nil {
		t.Fatalf("stream donwload failed, %s", errDownload)
	}

	expectedDataChunks := [][]byte{
		[]byte{0x00, 0x01},
		[]byte{0x02, 0x03},
		[]byte{0x04, 0x05},
		[]byte{0x06, 0x07},
		[]byte{0x08, 0x09},
		[]byte{0x0A, 0x0B},
		[]byte{0x0C, 0x0D},
		[]byte{0x0E, 0x0F},
		[]byte{0x10},
	}
	i := 0
	for d := range dataChan {
		log.Printf("data arrived %v\n", d)
		if bytes.Compare(d, expectedDataChunks[i]) != 0 {
			t.Errorf("got stream chunk data %v, expected %v", d, expectedDataChunks[i])
		}
		i++
	}

}

func TestGetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockclient(ctrl)
	d := mock.NewMockdecipherer(ctrl)

	stream := Stream{
		signature:  "hoge",
		url:        "https://foobar?itag=22",
		client:     c,
		decipherer: d,
	}
	expected := 50970333
	header := make(http.Header)
	header.Set("Content-Length", strconv.Itoa(expected))

	gomock.InOrder(
		d.EXPECT().Decipher("hoge").Return("geho", nil),
		c.EXPECT().Head("https://foobar?itag=22&signature=geho").Return(
			&http.Response{
				StatusCode: 200,
				Header:     header,
				Body:       nil,
			},
			nil,
		),
	)

	if size, err := stream.GetSize(); err != nil {
		t.Fatalf("get stream size failed, %s", err)
	} else if size != expected {
		t.Errorf("stream size expected %d, got %d", expected, size)
	}
}

func TestBuildDownloadURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := mock.NewMockdecipherer(ctrl)

	signature := "456ABCDEFG123"
	decipheredSignature := "ABCDEFG123456"
	d.EXPECT().Decipher(signature).Return(decipheredSignature, nil).Times(1)

	streamURL := "https://youtube.com?foo=bar"

	t.Run("signature and decipherer is missing", func(t *testing.T) {
		// Decipher not called
		stream := Stream{url: streamURL}
		if _, err := stream.getDownloadURL(); err == nil {
			t.Error("error should be raised when signature and decipherer are not given")
		}
	})

	t.Run("both signature and decipherer are valid", func(t *testing.T) {
		stream := Stream{
			url:        streamURL,
			signature:  signature,
			decipherer: d,
		}
		expected := fmt.Sprintf("%s&signature=%s", streamURL, decipheredSignature)
		if builtURL, err := stream.getDownloadURL(); err != nil {
			t.Errorf("failed to build download url, %s", err)
		} else if builtURL != expected {
			t.Errorf("wrong donwload URL is built %s, expected %s", builtURL, expected)
		}
	})

	t.Run("deciphererd signature has already been included in a url", func(t *testing.T) {
		// Decipher not called
		expected := fmt.Sprintf("%s&signature=%s", streamURL, decipheredSignature)
		stream := Stream{
			url: expected,
		}
		if builtURL, err := stream.getDownloadURL(); err != nil {
			t.Error("no error should be raised when deciphered signature is included in url")
		} else if builtURL != expected {
			t.Errorf("wrong donwload URL is built %s, expected %s", builtURL, expected)
		}
	})

}
