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
	mockPagePath                   = "./mocks/data/page.html.zip"
	mockScriptPath                 = "./mocks/data/script.js.zip"
	mockAgeRestrictedPagePath      = "./mocks/data/age_restricted_page.html.zip"
	mockAgeRestrictedEmbedPagePath = "./mocks/data/age_restricted_embed_page.html.zip"
	mockAgeRestrictedVideoInfoPath = "./mocks/data/age_restricted_video_info.txt.zip"
	mockAgeRestrictedScriptPath    = "./mocks/data/age_restricted_script.js.zip"
)

var (
	compressedFileNameRegex = regexp.MustCompile(`(.+).zip$`)
)

var videoStreams []*Stream
var restrictedVideoStreams []*Stream

func init() {
	args := []map[string]string{
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=22&key=yt6&ip=126.2.187.172&ms=au,onr&source=youtube&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&expire=1521061926&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&mime=video/mp4&lmt=1518448989107051&ratebypass=yes&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&c=WEB&mt=1521040211&ipbits=0&requiressl=yes&sparams=dur,ei,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&initcwndbps=772500",
			"s":       "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77",
			"quality": "hd720",
			"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
			"itag":    "22",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=0.000&pl=17&itag=43&source=youtube&expire=1521061926&c=WEB&mime=video/webm&ratebypass=yes&ipbits=0&clen=28050507&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450467893651&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire",
			"s":       "5BD6ED7C27FD64795240C063804623C724EA1146.C8E7C24E94A7826704380EFBD4322E2CD22421E99",
			"quality": "medium",
			"type":    "video/webm; codecs=\"vp8.0, vorbis\"",
			"itag":    "43",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=18&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&ratebypass=yes&ipbits=0&clen=19799961&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448034625421&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire",
			"s":       "4055B39977B419C9388A9FB7B96151F84DAAE97A.33099D716E75A1B378B338CAA0E25A85DBDC1EB99",
			"quality": "medium",
			"type":    "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"",
			"itag":    "18",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.825&pl=17&itag=36&source=youtube&expire=1521061926&c=WEB&mime=video/3gpp&ipbits=0&clen=7859105&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448004377270&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "DD71E200D3B3FA0B7A61E4B1190A6788EF37E829.D57B8FE16ABCE0FA068DAA7EE779F39EF55644D99",
			"quality": "small",
			"type":    "video/3gpp; codecs=\"mp4v.20.3, mp4a.40.2\"",
			"itag":    "36",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.825&pl=17&itag=17&source=youtube&expire=1521061926&c=WEB&mime=video/3gpp&ipbits=0&clen=2839609&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448027106230&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "4151B9AAE45D0E237689C39495A83A92803A2242.19685C96AC0407995B7D3589720AA258FF1EAE222",
			"quality": "small",
			"type":    "video/3gpp; codecs=\"mp4v.20.3, mp4a.40.2\"",
			"itag":    "17",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=137&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=59035281&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448904618557&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "28CC0A36652CAD187D627E16636997E601D4B8F1.AA4899BAD0E70AB17CE344B7AF61FFF6033A41D88",
			"quality": "1080p",
			"type":    "video/mp4; codecs=\"avc1.640028\"",
			"itag":    "137",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=248&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=56688957&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450960637068&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "13C56BC945D559A010AE708D3A9FB7E2650DE7A2.1B18F6CBB5497D4F1440DF6CAE5C8964F90F0B577",
			"quality": "1080p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "248",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=136&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=46557774&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448868078020&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "8594FDB60B20F293E4DAA696057E092596A93A78.2364425B43A595169059F3A1B7D492653242820FF",
			"quality": "720p",
			"type":    "video/mp4; codecs=\"avc1.4d401f\"",
			"itag":    "136",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=247&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=33606092&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450887315122&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "8ADB6C2A7B1495D5D185D36E9DD7D551DA31144E.378E540B4D43BFCD55D3F0FDFE9917A25AF5F2533",
			"quality": "720p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "247",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=135&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=27954678&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448850968833&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "8D6C9C3B3120D9255D84CB41321CB2BDF7DFB509.7F95C0F4A520800AEC669E42D15468B9CB8997ABB",
			"quality": "480p",
			"type":    "video/mp4; codecs=\"avc1.4d401e\"",
			"itag":    "135",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=244&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=17347971&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450882558526&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "C2C7BC594CA1F4E418E7DA94BCB7CA735FF2CFEE.0964F3E744CA0E5AFAB2878313CC8C31D72BEF211",
			"quality": "480p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "244",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=134&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=14778536&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448844863372&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "F70DC518AA116B720F68F788E46DFDFEE48150C0.55169044E9BB31BD76A9740D037ACC17F00B5C944",
			"quality": "360p",
			"type":    "video/mp4; codecs=\"avc1.4d401e\"",
			"itag":    "134",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=243&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=9705599&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450794575285&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "B164216AC1F0537301E589F7E7928A6199E6A7DF.BFE7D1BD8855371FD5F06E72C882FE7A4345572CC",
			"quality": "360p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "243",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=133&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=6811724&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448841560979&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "5957E567EE4542FF58B76F03F31852C545937872.AAFA1205E28978EB1098F4105ADBE667E213E1D11",
			"quality": "240p",
			"type":    "video/mp4; codecs=\"avc1.4d4015\"",
			"itag":    "133",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=242&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=5104802&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450802468932&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "FB815B445D791C4D86FAAF09CDA481ED7376D03E.3E9F2B0348A5CCCCCF20E033531D1B29804041544",
			"quality": "240p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "242",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=160&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/mp4&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=2945134&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448840259190&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "7C231A919F112805152924B7D38D22A8B96CA898.A500E0B53B2D3485315648357B4E668E92D67EA88",
			"quality": "144p",
			"type":    "video/mp4; codecs=\"avc1.4d400c\"",
			"itag":    "160",
		},
		map[string]string{
			"url":     "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.653&pl=17&itag=278&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=video/webm&aitags=133,134,135,136,137,160,242,243,244,247,248,278&ipbits=0&clen=3161565&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450831390303&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":       "50E302DCB14CE59CFB3C7633D60C9756882D58B1.263F06C5DCF635DC64F45668A1177B58EA1C196FF",
			"quality": "144p",
			"type":    "video/webm; codecs=\"vp9\"",
			"itag":    "278",
		},
		map[string]string{
			"url":  "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.778&pl=17&itag=140&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=audio/mp4&ipbits=0&clen=4428314&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518448574293099&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":    "004B5D0AEFF60F4A4954E7502CA1A0B270916BC0.DF48B24DD36B89FAB7AF5B6815EE448C75F1EF5DD",
			"type": "audio/mp4; codecs=\"mp4a.40.2\"",
			"itag": "140",
		},
		map[string]string{
			"url":  "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.718&pl=17&itag=171&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=audio/webm&ipbits=0&clen=3825867&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450499610158&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":    "6FB02005AC59045CF0BA995CA96C8C7986C4AB3F.3064355DF67DB2589DF99F9FED382F0B63A80AB22",
			"type": "audio/webm; codecs=\"vorbis\"",
			"itag": "171",
		},
		map[string]string{
			"url":  "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.741&pl=17&itag=249&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=audio/webm&ipbits=0&clen=1717626&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450461531768&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":    "F327D0728D7872EC9195D16F14D57AE547CA5639.5FD1BC8990C6D6A31A87B061E65A93E9E5D96F077",
			"type": "audio/webm; codecs=\"opus\"",
			"itag": "249",
		},
		map[string]string{
			"url":  "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.741&pl=17&itag=250&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=audio/webm&ipbits=0&clen=2219765&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450472211655&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":    "995A55948FA2F10B8BA3CBFC4FDFEB172CC33659.98E94FD530DF08344674937354C6C86DCA2500466",
			"type": "audio/webm; codecs=\"opus\"",
			"itag": "250",
		},
		map[string]string{
			"url":  "https://r1---sn-ogul7n7s.googlevideo.com/videoplayback?dur=278.741&pl=17&itag=251&keepalive=yes&source=youtube&expire=1521061926&c=WEB&mime=audio/webm&ipbits=0&clen=4386238&initcwndbps=772500&key=yt6&ip=126.2.187.172&ms=au,onr&mt=1521040211&mv=m&id=o-AIMLylRYsEhoQdkeKYgwTOxaBWNe5BVIMqnuOI5fUZyR&mm=31,26&mn=sn-ogul7n7s,sn-3pm7snez&gir=yes&lmt=1518450482584196&ei=xjupWtr7NIifqQG8vp_gBA&fvip=1&requiressl=yes&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire",
			"s":    "16B55DA94087076C8A4C6F4AF3DDD11D9686CFB0.D73F8D2134ABE26E757E2E4708E038964074DC222",
			"type": "audio/webm; codecs=\"opus\"",
			"itag": "251",
		},
	}

	videoStreams = make([]*Stream, 0, len(args))
	for _, a := range args {
		s, _ := newStream(a, nil, nil)
		videoStreams = append(videoStreams, s)
	}

	args = []map[string]string{
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&ip=126.225.83.8&c=WEB&dur=282.192&requiressl=yes&initcwndbps=1176250&pl=17&mt=1523929552&ratebypass=yes&source=youtube&mv=m&ms=au,rdu&signature=14E7732E9DED9E67E16E3FE08E2CAE254258E860.0FA43416C13303EF0320E3C370FB06AAD94F7509&ipbits=0&lmt=1507668791912906&sparams=dur,ei,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&itag=22&fvip=5&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&expire=1523951333&mime=video/mp4&key=yt6",
			"quality": "hd720",
			"type":    "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
			"itag":    "22",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=14130CAA3CE17F3F5CC13F80F9D5ACAF8CE14841.6B17B1DBCA26C66AEA61BB4984F494D2376EF510&ipbits=0&lmt=1361409963312387&itag=43&fvip=5&expire=1523951333&mime=video/webm&key=yt6&requiressl=yes&gir=yes&dur=0.000&initcwndbps=1176250&pl=17&ratebypass=yes&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&clen=29851711",
			"quality": "medium",
			"type":    "video/webm; codecs=\"vp8.0, vorbis\"",
			"itag":    "43",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=D050C6CA8B2567D874BA88698D0E305B585AF843.8E5C3304435958825B4CE9516279DF9BF5DDF8ED&ipbits=0&lmt=1446644551421167&itag=18&fvip=5&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.192&initcwndbps=1176250&pl=17&ratebypass=yes&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,ratebypass,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&clen=24397217",
			"quality": "medium",
			"type":    "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"",
			"itag":    "18",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=32630DBCE931C335112C8E41202D423003A28E71.8F02C30BD08D57430A9303B50CA32FBA45A8A856&ipbits=0&lmt=1386887845620058&itag=36&fvip=5&expire=1523951333&mime=video/3gpp&key=yt6&requiressl=yes&gir=yes&dur=282.354&initcwndbps=1176250&pl=17&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&clen=8065709",
			"quality": "small",
			"type":    "video/3gpp; codecs=\"mp4v.20.3, mp4a.40.2\"",
			"itag":    "36",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=8E9A4340ED12AE9E7DB96BA092413E28AAE76CB7.D4AD90C02D670456F17CE4630C10F81A06F8E87A&ipbits=0&lmt=1386887774806263&itag=17&fvip=5&expire=1523951333&mime=video/3gpp&key=yt6&requiressl=yes&gir=yes&dur=282.586&initcwndbps=1176250&pl=17&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&clen=2893716",
			"quality": "small",
			"type":    "video/3gpp; codecs=\"mp4v.20.3, mp4a.40.2\"",
			"itag":    "17",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=6B2EBAA1C8BB5DE98FC725515148472A6F74E39C.2FA2CA4BCF878B893E104C779453742C3C06154A&ipbits=0&lmt=1507668511230467&itag=137&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=124634963",
			"quality": "1080p",
			"type":    "video/mp4; codecs=\"avc1.640028\"",
			"itag":    "137",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=5EAD9AF00CF84B3BB17228ECD580B7B60785CAEE.56E25A59F2198A96B67D6D635B62B934A0B0CB3A&ipbits=0&lmt=1507668526298009&itag=136&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=64136030",
			"quality": "720p",
			"type":    "video/mp4; codecs=\"avc1.4d401f\"",
			"itag":    "136",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=94C75F74B566805BA52947993E6C147584E13B0D.98A16801F8B6A484F775AF76F0ECB2A51C1D090E&ipbits=0&lmt=1507668526284102&itag=135&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=32766677",
			"quality": "480p",
			"type":    "video/mp4; codecs=\"avc1.4d401f\"",
			"itag":    "135",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=5BC062D479646FADB82641D8A46C40A83E0AD055.BCAF86384F7197FC6C59460076F03FE524D09E60&ipbits=0&lmt=1507668526481649&itag=134&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=16964932",
			"quality": "360p",
			"type":    "video/mp4; codecs=\"avc1.4d401e\"",
			"itag":    "134",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=0B5D5FE1D19A33BB6B3B4F81D4A59373F9FBE11E.C993D63CA86DB4B3CFDD8606054723780D419050&ipbits=0&lmt=1507668526074111&itag=133&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=6340756",
			"quality": "240p",
			"type":    "video/mp4; codecs=\"avc1.4d4015\"",
			"itag":    "133",
		},
		map[string]string{
			"url":     "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=51AEC9DE9CA17A6D212346CF1EC1969566AE88A1.214C279A3A5890463FDC8E48816CE9DC5D986AEC&ipbits=0&lmt=1507668526072795&itag=160&aitags=133,134,135,136,137,160&expire=1523951333&mime=video/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.133&initcwndbps=1176250&fvip=5&pl=17&source=youtube&ip=126.225.83.8&sparams=aitags,clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=3058777",
			"quality": "144p",
			"type":    "video/mp4; codecs=\"avc1.4d400c\"",
			"itag":    "160",
		},
		map[string]string{
			"url":  "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=55E02BAC7A14827B4ACD99CEC38A3EBA84BE47D1.5389D4A5C412C0EAB902733431FBD40CC5E10410&ipbits=0&lmt=1507668635300298&itag=140&fvip=5&expire=1523951333&mime=audio/mp4&key=yt6&requiressl=yes&gir=yes&dur=282.192&initcwndbps=1176250&pl=17&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=4482638",
			"type": "audio/mp4; codecs=\"mp4a.40.2\"",
			"itag": "140",
		},
		map[string]string{
			"url":  "https://r1---sn-3pm7sn7r.googlevideo.com/videoplayback?mm=31,29&mn=sn-3pm7sn7r,sn-3pm76n7s&c=WEB&mt=1523929552&mv=m&ms=au,rdu&signature=35D8DAB2F526D222F0EAE2D1C03015F4CA85FB45.4AD1B0A2F09C24B1301B11352AC5A8C373DBD8AA&ipbits=0&lmt=1393566609280233&itag=171&fvip=5&expire=1523951333&mime=audio/webm&key=yt6&requiressl=yes&gir=yes&dur=282.162&initcwndbps=1176250&pl=17&source=youtube&ip=126.225.83.8&sparams=clen,dur,ei,gir,id,initcwndbps,ip,ipbits,itag,keepalive,lmt,mime,mm,mn,ms,mv,pl,requiressl,source,expire&ei=hVLVWpGbJYS64gLKjoPoBg&id=o-ADVGi3m71YbktDcznNDu6_PehGXwDu1EQqwAQpvkRfCs&keepalive=yes&clen=3705144",
			"type": "audio/webm; codecs=\"vorbis\"",
			"itag": "171",
		},
	}

	restrictedVideoStreams = make([]*Stream, 0, len(args))
	for _, a := range args {
		s, _ := newStream(a, nil, nil)
		restrictedVideoStreams = append(restrictedVideoStreams, s)
	}

}

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

func getStrStream(streamPath string) (string, error) {
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
