package main

import (
	"fmt"
	"io"
	"log"
	"os"

	gotube "github.com/matthewlujp/gotube/lib"
)

func main() {
	usageText := "Usage: gotube [Youtube video url] [file path to save a downloaded video]"
	if len(os.Args) < 3 {
		panic(usageText)
	}

	url := os.Args[1]
	saveFilePath := os.Args[2]

	downloader, errNewPlayer := gotube.NewDownloader(url)
	if errNewPlayer != nil {
		panic(errNewPlayer)
	}

	if err := downloader.FetchStreams(); err != nil {
		panic(err)
	}
	log.Printf("streams fetched, %v", downloader.Streams)

	for _, s := range downloader.Streams {
		log.Println("strem", s.MediaType, s.Format, s.Resolution, s.Abr, s.VideoCodec, s.AudioCodec)
		if s.MediaType == "video" && s.Format == "mp4" && s.AudioCodec != "" {
			dataReader, errDownload := s.Download()
			if errDownload != nil {
				panic(errDownload)
			}
			defer dataReader.Close()
			log.Println("download succeeded")

			f, errOpen := os.Create(saveFilePath)
			if errOpen != nil {
				panic(errOpen)
			}
			if _, err := io.Copy(f, dataReader); err != nil {
				panic(err)
			}
			fmt.Printf("video contents of %s is saved in %s.\nbitrate %s, fps %s, resolution %s\n", url, saveFilePath, s.Abr, s.Fps, s.Resolution)
			return
		}
	}

	fmt.Printf("no mp4 stream is not found for %s", url)
}
