package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	gotube "github.com/matthewlujp/gotube/lib"
	"github.com/pkg/profile"
)

var (
	saveFilePath *string
	url          string
	cpuProfile   *bool
)

func init() {
	usageText := "Usage: gotube [Youtube video url] [file path to save a downloaded video]"
	saveFilePath = flag.String("s", "", "save file path")
	cpuProfile = flag.Bool("p", false, "write cpu profile to a file under /var")
	flag.Parse()

	if !*cpuProfile && flag.NArg() < 1 {
		log.Fatalln(usageText)
	}
	url = flag.Arg(0)
}

func main() {
	if *cpuProfile {
		runProfile()
	} else {
		run()
	}
}

func runProfile() {
	// code for cpu profiling
	fmt.Println("cpu profile")
	defer profile.Start().Stop()

	streams, errFetch := getStreams("https://www.youtube.com/watch?v=09R8_2nJtjg")
	if errFetch != nil {
		log.Fatalln(errFetch)
	}
	// data, errDownload := streams[19].Download()
	data, errDownload := streams[19].ParallelDownload()
	if errDownload != nil {
		log.Fatalln(errDownload)
	}
	if err := save("test.mp4", data); err != nil {
		log.Fatalln(err)
	}
}

func run() {
	// code for command line usage
	streams, err := getStreams(url)
	if err != nil {
		log.Fatal(err)
	}
	streamID := printStreamsAndPrompt(streams) // make user choose a stream

	// download a designated stream
	stream := streams[streamID]
	fmt.Printf("Downloading %d th stream, %s ......", streamID, stream)
	// data, errDownload := stream.Download()
	data, errDownload := stream.ParallelDownload()
	if errDownload != nil {
		log.Fatalf("failed to download stream %s, %s", stream, errDownload)
	}
	fmt.Println("Downloaded")

	// make user to input save file path if not designated as a commandline flag
	if *saveFilePath == "" {
		fmt.Print("Where to save the video?> ")
		fmt.Scan(saveFilePath)
	}
	fmt.Printf("Saving on %s......", *saveFilePath)

	if err := save(*saveFilePath, data); err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Download completed!\nWritten on %s.\nBitrate %s, FPS %s, Resolution %s\n", *saveFilePath, stream.Abr, stream.Fps, stream.Resolution)
}

func getStreams(url string) ([]*gotube.Stream, error) {
	downloader, errNewPlayer := gotube.NewDownloader(url)
	if errNewPlayer != nil {
		return nil, errNewPlayer
	}
	if err := downloader.FetchStreams(); err != nil {
		return nil, err
	}
	return downloader.Streams, nil
}

func printStreamsAndPrompt(streams []*gotube.Stream) int {
	fmt.Println("Fetched streams:\nID    Stream info")
	for i, s := range streams {
		fmt.Printf("%d --- %s\n", i, s)
	}

	fmt.Print("Choose stream ID> ")
	var streamID int
	fmt.Scan(&streamID)
	fmt.Println("")
	return streamID
}

func save(path string, data []byte) error {
	f, errOpen := os.Create(path)
	if errOpen != nil {
		return fmt.Errorf("failed to create or open %s, %s", path, errOpen)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("error while writing the downloader video to the file, %s", err)
	}
	return nil
}
