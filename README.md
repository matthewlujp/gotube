# gotube
Youtube downloader implemented in Golang.
This project is based on [nficano/pytube](https://github.com/nficano/pytube).

# How to
## Use as a library

```go:sample.go
import gotube "github.com/matthewlujp/gotube/lib"

downloader, _ := gotube.NewPlayer("https://youtube.com/watch?v=iEPTlhBmwRg")
downloader.FetchStreamManifests()
r, _ := downloader.Streams[0].Download()
// r is a io.ReadCloser
```

## Command line usage
After building the source, execute the following.

```sh
$ gotube "https://www.youtube.com/watch?v=09R8_2nJtjg"                                                           
Fetched streams:
ID    Stream info
0 --- Stream<MediaType:video Quality:1080p Format:mp4 Resolution:1080p>
1 --- Stream<MediaType:video Quality:1080p Format:webm Resolution:1080p>
2 --- Stream<MediaType:video Quality:720p Format:mp4 Resolution:720p>
3 --- Stream<MediaType:video Quality:720p Format:webm Resolution:720p>
4 --- Stream<MediaType:video Quality:480p Format:mp4 Resolution:480p>
5 --- Stream<MediaType:video Quality:480p Format:webm Resolution:480p>
6 --- Stream<MediaType:video Quality:360p Format:mp4 Resolution:360p>
7 --- Stream<MediaType:video Quality:360p Format:webm Resolution:360p>
8 --- Stream<MediaType:video Quality:240p Format:mp4 Resolution:240p>
9 --- Stream<MediaType:video Quality:240p Format:webm Resolution:240p>
10 --- Stream<MediaType:video Quality:144p Format:mp4 Resolution:144p>
11 --- Stream<MediaType:video Quality:144p Format:webm Resolution:144p>
12 --- Stream<MediaType:audio Quality: Format:mp4 Resolution:>
13 --- Stream<MediaType:audio Quality: Format:webm Resolution:>
14 --- Stream<MediaType:audio Quality: Format:webm Resolution:>
15 --- Stream<MediaType:audio Quality: Format:webm Resolution:>
16 --- Stream<MediaType:audio Quality: Format:webm Resolution:>
17 --- Stream<MediaType:video Quality:hd720 Format:mp4 Resolution:720p>
18 --- Stream<MediaType:video Quality:medium Format:webm Resolution:360p>
19 --- Stream<MediaType:video Quality:medium Format:mp4 Resolution:360p>
20 --- Stream<MediaType:video Quality:small Format:3gpp Resolution:240p>
21 --- Stream<MediaType:video Quality:small Format:3gpp Resolution:144p>
Choose stream ID> 19

Downloading 19 th stream, Stream<MediaType:video Quality:medium Format:mp4 Resolution:360p> ......Downloaded
Where to save the video?> youtube_video.mp4
Saving on youtube_video.mp4......Download completed!
Written on youtube_video.mp4.
Bitrate 96kbps, FPS 30, Resolution 360p
```

Propmt to 1) choose which stream to download, and 2) designate file path to save the video, appears.
You can also designate save file path with option -s as follows.

```sh
$ gotube "https://www.youtube.com/watch?v=09R8_2nJtjg" -s youtube_video.mp4
```

There are pre-buit binaries for OSX, Linxus, and Windos (all of them are for amd64, i.e., x86_64).
You can pick one from bins.


## Run test
gomock and mockgen are used (https://github.com/golang/mock).
Makefile downloads those libraries when executing test.

```sh
cd $Go/src/github.com/matthewlujp/gotube
make test
```

