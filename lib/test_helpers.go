// This methods in this src are only used in test.
// Do not use for implementation of functionalities.

package gotube

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	mockPagePath            = "./mocks/data/page.html.zip"
	mockScriptPath          = "./mocks/data/script.js.zip"
	rawURLEncodedStreamPath = "./mocks/data/raw_url_streams.txt"
	rawAdaptiveStreamPath   = "./mocks/data/raw_adaptive_streams.txt"
)

type StreamType int

const (
	URLEncoded = iota
	Adaptive
)

var (
	compressedFileNameRegex = regexp.MustCompile(`(.+).zip$`)
	jsURL                   = "https://youtube.com/yts/jsbin/player-vfllqtOs7/ja_JP/base.js"
	title                   = "Maroon 5 - Moves Like Jagger ft. Christina Aguilera"
	streamNumbers           = 22
	urlEmbedSignatures      = []string{
		"55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
		"5BD6ED7C27FD64795240C063804623C724EA1146.C8E7C24E94A7826704380EFBD4322E2CD22421E99",
		"4055B39977B419C9388A9FB7B96151F84DAAE97A.33099D716E75A1B378B338CAA0E25A85DBDC1EB99",
		"DD71E200D3B3FA0B7A61E4B1190A6788EF37E829.D57B8FE16ABCE0FA068DAA7EE779F39EF55644D99",
		"4151B9AAE45D0E237689C39495A83A92803A2242.19685C96AC0407995B7D3589720AA258FF1EAE222",
	}
	adaptiveSignatures = []string{
		"28CC0A36652CAD187D627E16636997E601D4B8F1.AA4899BAD0E70AB17CE344B7AF61FFF6033A41D88",
		"13C56BC945D559A010AE708D3A9FB7E2650DE7A2.1B18F6CBB5497D4F1440DF6CAE5C8964F90F0B577",
		"8594FDB60B20F293E4DAA696057E092596A93A78.2364425B43A595169059F3A1B7D492653242820FF",
		"8ADB6C2A7B1495D5D185D36E9DD7D551DA31144E.378E540B4D43BFCD55D3F0FDFE9917A25AF5F2533",
		"8D6C9C3B3120D9255D84CB41321CB2BDF7DFB509.7F95C0F4A520800AEC669E42D15468B9CB8997ABB",
		"C2C7BC594CA1F4E418E7DA94BCB7CA735FF2CFEE.0964F3E744CA0E5AFAB2878313CC8C31D72BEF211",
		"F70DC518AA116B720F68F788E46DFDFEE48150C0.55169044E9BB31BD76A9740D037ACC17F00B5C944",
		"B164216AC1F0537301E589F7E7928A6199E6A7DF.BFE7D1BD8855371FD5F06E72C882FE7A4345572CC",
		"5957E567EE4542FF58B76F03F31852C545937872.AAFA1205E28978EB1098F4105ADBE667E213E1D11",
		"FB815B445D791C4D86FAAF09CDA481ED7376D03E.3E9F2B0348A5CCCCCF20E033531D1B29804041544",
		"7C231A919F112805152924B7D38D22A8B96CA898.A500E0B53B2D3485315648357B4E668E92D67EA88",
		"50E302DCB14CE59CFB3C7633D60C9756882D58B1.263F06C5DCF635DC64F45668A1177B58EA1C196FF",
		"004B5D0AEFF60F4A4954E7502CA1A0B270916BC0.DF48B24DD36B89FAB7AF5B6815EE448C75F1EF5DD",
		"6FB02005AC59045CF0BA995CA96C8C7986C4AB3F.3064355DF67DB2589DF99F9FED382F0B63A80AB22",
		"F327D0728D7872EC9195D16F14D57AE547CA5639.5FD1BC8990C6D6A31A87B061E65A93E9E5D96F077",
		"995A55948FA2F10B8BA3CBFC4FDFEB172CC33659.98E94FD530DF08344674937354C6C86DCA2500466",
		"16B55DA94087076C8A4C6F4AF3DDD11D9686CFB0.D73F8D2134ABE26E757E2E4708E038964074DC22",
	}
)

func readCompressedFile(filePath string) ([]byte, error) {
	var filename string
	if res := compressedFileNameRegex.FindStringSubmatch(filepath.Base(filePath)); res == nil || len(res) < 2 {
		return nil, fmt.Errorf("%s is not compressed file", filePath)
	} else {
		filename = res[1]
	}

	zf, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer zf.Close()

	for _, file := range zf.File {
		if file.Name == filename {
			fc, errOpen := file.Open()
			if errOpen != nil {
				return nil, errOpen
			}
			buf := new(bytes.Buffer)
			_, errRead := buf.ReadFrom(fc)
			if errRead != nil {
				return nil, errRead
			}
			return buf.Bytes(), nil
		}
	}
	return nil, fmt.Errorf("failed to open %s", filePath)
}

func getContent(filePath string) (*http.Response, error) {
	contents, err := readCompressedFile(filePath)
	if err != nil {
		return nil, err
	}

	res := http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(contents)),
	}
	return &res, nil
}

func getMockPage() (*http.Response, error) {
	return getContent(mockPagePath)
}

func getMockScript() (*http.Response, error) {
	return getContent(mockScriptPath)
}

func getStrStream(st StreamType) (string, error) {
	var streamPath string
	switch st {
	case URLEncoded:
		streamPath = rawURLEncodedStreamPath
	case Adaptive:
		streamPath = rawAdaptiveStreamPath
	default:
		panic("got invalid type specification")
	}

	f, err := os.Open(streamPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", err
	}
	strStream := strings.Replace(string(buf.Bytes()[:]), "\n", "", -1)
	return strStream, nil
}
