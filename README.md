# gotube
Youtube downloader implemented in Golang

# How to
## Use as a library

```go:sample.go
import gotube "github.com/matthewlujp/gotube/lib"

player, _ := gotube.NewPlayer("https://youtube.com/watch?v=iEPTlhBmwRg")
streams, _ := player.FetchStreamManifests()
r, _ := streams[0].Download()
// r is a io.ReadCloser
```

## Command line usage

```sh
cd $Go/src/github.com/matthewlujp/gotube
make build

./downloader "https://youtube.com/watch?v=iEPTlhBmwRg" $HOME/Desktop/download.mp4
```


## Run test

```sh
cd $Go/src/github.com/matthewlujp/gotube
make test
```

