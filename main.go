package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	gotube "github.com/matthewlujp/gotube/lib"
)

func main() {
	var saveFilePath string

	usageText := "Usage: gotube [Youtube video url] [file path to save a downloaded video]"
	flag.StringVar(&saveFilePath, "s", "", "save file path")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println(usageText)
		os.Exit(-1)
	}
	url := flag.Arg(0)

	downloader, errNewPlayer := gotube.NewDownloader(url)
	if errNewPlayer != nil {
		panic(errNewPlayer)
	}
	if err := downloader.FetchStreams(); err != nil {
		panic(err)
	}

	fmt.Println("Fetched streams:\nID    Stream info")
	for i, s := range downloader.Streams {
		fmt.Printf("%d --- %s\n", i, s)
	}

	fmt.Print("Choose stream ID> ")
	var streamID int
	fmt.Scan(&streamID)
	fmt.Println("")

	// download a designated stream
	stream := downloader.Streams[streamID]
	fmt.Printf("Downloading %d th stream, %s ......", streamID, stream)
	dataReader, errDownload := stream.Download()
	if errDownload != nil {
		fmt.Printf("failed to download stream %s, %s", stream, errDownload)
		os.Exit(-1)
	}
	defer dataReader.Close()
	fmt.Println("Downloaded")

	if saveFilePath == "" {
		fmt.Print("Where to save the video?> ")
		fmt.Scan(&saveFilePath)
	}
	fmt.Printf("Saving on %s......", saveFilePath)
	f, errOpen := os.Create(saveFilePath)
	if errOpen != nil {
		fmt.Printf("Failed to create or open %s, %s\n", saveFilePath, errOpen)
		os.Exit(-1)
	}
	if _, err := io.Copy(f, dataReader); err != nil {
		fmt.Printf("Error while writing the downloader video to the file, %s", err)
		os.Exit(-1)
	}
	fmt.Printf("Download completed!\nWritten on %s.\nBitrate %s, FPS %s, Resolution %s\n", saveFilePath, stream.Abr, stream.Fps, stream.Resolution)
}
