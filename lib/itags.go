package gotube

type formatProfile struct {
	itag       int
	resolution string
	bitrate    string
	is3D       bool
	isLive     bool
	is60fps    bool
}

var itags = map[int]formatProfile{
	5:   formatProfile{itag: 5, resolution: "240p", bitrate: "64kbps"},
	6:   formatProfile{itag: 6, resolution: "270p", bitrate: "64kbps"},
	13:  formatProfile{itag: 13, resolution: "144p", bitrate: ""},
	17:  formatProfile{itag: 17, resolution: "144p", bitrate: "24kbps"},
	18:  formatProfile{itag: 18, resolution: "360p", bitrate: "96kbps"},
	22:  formatProfile{itag: 22, resolution: "720p", bitrate: "192kbps"},
	34:  formatProfile{itag: 34, resolution: "360p", bitrate: "128kbps"},
	35:  formatProfile{itag: 35, resolution: "480p", bitrate: "128kbps"},
	36:  formatProfile{itag: 36, resolution: "240p", bitrate: ""},
	37:  formatProfile{itag: 37, resolution: "1080p", bitrate: "192kbps"},
	38:  formatProfile{itag: 38, resolution: "3072p", bitrate: "192kbps"},
	43:  formatProfile{itag: 43, resolution: "360p", bitrate: "128kbps"},
	44:  formatProfile{itag: 44, resolution: "480p", bitrate: "128kbps"},
	45:  formatProfile{itag: 45, resolution: "720p", bitrate: "192kbps"},
	46:  formatProfile{itag: 46, resolution: "1080p", bitrate: "192kbps"},
	59:  formatProfile{itag: 59, resolution: "480p", bitrate: "128kbps"},
	78:  formatProfile{itag: 78, resolution: "480p", bitrate: "128kbps"},
	82:  formatProfile{itag: 82, resolution: "360p", bitrate: "128kbps"},
	83:  formatProfile{itag: 83, resolution: "480p", bitrate: "128kbps"},
	84:  formatProfile{itag: 84, resolution: "720p", bitrate: "192kbps"},
	85:  formatProfile{itag: 85, resolution: "1080p", bitrate: "192kbps"},
	91:  formatProfile{itag: 91, resolution: "144p", bitrate: "48kbps"},
	92:  formatProfile{itag: 92, resolution: "240p", bitrate: "48kbps"},
	93:  formatProfile{itag: 93, resolution: "360p", bitrate: "128kbps"},
	94:  formatProfile{itag: 94, resolution: "480p", bitrate: "128kbps"},
	95:  formatProfile{itag: 95, resolution: "720p", bitrate: "256kbps"},
	96:  formatProfile{itag: 96, resolution: "1080p", bitrate: "256kbps"},
	100: formatProfile{itag: 100, resolution: "360p", bitrate: "128kbps"},
	101: formatProfile{itag: 101, resolution: "480p", bitrate: "192kbps"},
	102: formatProfile{itag: 102, resolution: "720p", bitrate: "192kbps"},
	132: formatProfile{itag: 132, resolution: "240p", bitrate: "48kbps"},
	151: formatProfile{itag: 151, resolution: "720p", bitrate: "24kbps"},
	// DASH Video
	133: formatProfile{itag: 133, resolution: "240p", bitrate: ""},
	134: formatProfile{itag: 134, resolution: "360p", bitrate: ""},
	135: formatProfile{itag: 135, resolution: "480p", bitrate: ""},
	136: formatProfile{itag: 136, resolution: "720p", bitrate: ""},
	137: formatProfile{itag: 137, resolution: "1080p", bitrate: ""},
	138: formatProfile{itag: 138, resolution: "2160p", bitrate: ""},
	160: formatProfile{itag: 160, resolution: "144p", bitrate: ""},
	167: formatProfile{itag: 167, resolution: "360p", bitrate: ""},
	168: formatProfile{itag: 168, resolution: "480p", bitrate: ""},
	169: formatProfile{itag: 169, resolution: "720p", bitrate: ""},
	170: formatProfile{itag: 170, resolution: "1080p", bitrate: ""},
	212: formatProfile{itag: 212, resolution: "480p", bitrate: ""},
	218: formatProfile{itag: 218, resolution: "480p", bitrate: ""},
	219: formatProfile{itag: 219, resolution: "480p", bitrate: ""},
	242: formatProfile{itag: 242, resolution: "240p", bitrate: ""},
	243: formatProfile{itag: 243, resolution: "360p", bitrate: ""},
	244: formatProfile{itag: 244, resolution: "480p", bitrate: ""},
	245: formatProfile{itag: 245, resolution: "480p", bitrate: ""},
	246: formatProfile{itag: 246, resolution: "480p", bitrate: ""},
	247: formatProfile{itag: 247, resolution: "720p", bitrate: ""},
	248: formatProfile{itag: 248, resolution: "1080p", bitrate: ""},
	264: formatProfile{itag: 264, resolution: "144p", bitrate: ""},
	266: formatProfile{itag: 266, resolution: "2160p", bitrate: ""},
	271: formatProfile{itag: 271, resolution: "144p", bitrate: ""},
	272: formatProfile{itag: 272, resolution: "2160p", bitrate: ""},
	278: formatProfile{itag: 278, resolution: "144p", bitrate: ""},
	298: formatProfile{itag: 298, resolution: "720p", bitrate: ""},
	299: formatProfile{itag: 299, resolution: "1080p", bitrate: ""},
	302: formatProfile{itag: 302, resolution: "720p", bitrate: ""},
	303: formatProfile{itag: 303, resolution: "1080p", bitrate: ""},
	308: formatProfile{itag: 308, resolution: "1440p", bitrate: ""},
	313: formatProfile{itag: 313, resolution: "2160p", bitrate: ""},
	315: formatProfile{itag: 315, resolution: "2160p", bitrate: ""},
	// DASH Audio
	139: formatProfile{itag: 139, resolution: "", bitrate: "48kbps"},
	140: formatProfile{itag: 140, resolution: "", bitrate: "128kbps"},
	141: formatProfile{itag: 141, resolution: "", bitrate: "256kbps"},
	171: formatProfile{itag: 171, resolution: "", bitrate: "128kbps"},
	172: formatProfile{itag: 172, resolution: "", bitrate: "256kbps"},
	249: formatProfile{itag: 249, resolution: "", bitrate: "50kbps"},
	250: formatProfile{itag: 250, resolution: "", bitrate: "70kbps"},
	251: formatProfile{itag: 251, resolution: "", bitrate: "160kbps"},
	256: formatProfile{itag: 256, resolution: "", bitrate: ""},
	258: formatProfile{itag: 258, resolution: "", bitrate: ""},
	325: formatProfile{itag: 325, resolution: "", bitrate: ""},
	328: formatProfile{itag: 328, resolution: "", bitrate: ""},
}

var itags60FPS = [6]int{298, 299, 302, 303, 308, 315}
var itags3D = [7]int{82, 83, 84, 85, 100, 101, 102}
var itagsLive = [8]int{91, 92, 93, 94, 95, 96, 132, 151}

func getFormatProfile(itag int) *formatProfile {
	format := formatProfile{itag: itag}
	if f, ok := itags[itag]; ok {
		format.resolution = f.resolution
		format.bitrate = f.bitrate
	}
	for _, i := range itags60FPS {
		if i == itag {
			format.is60fps = true
		}
	}
	for _, i := range itags3D {
		if i == itag {
			format.is3D = true
		}
	}
	for _, i := range itagsLive {
		if i == itag {
			format.isLive = true
		}
	}
	return &format
}
