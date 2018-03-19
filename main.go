package main

import (
	"fmt"
	"io"
	"log"
	"os"

	gotube "github.com/matthewlujp/gotube/lib"
)

func main() {
	url := os.Args[1]
	saveFilePath := os.Args[2]

	p, errNewPlayer := gotube.NewPlayer(url)
	if errNewPlayer != nil {
		panic(errNewPlayer)
	}
	log.Printf("new player created")
	streams, errFetch := p.FetchStreamManifests()
	if errFetch != nil {
		panic(errFetch)
	}
	log.Printf("streams fetched, %v", streams)

	for _, s := range streams {
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
			_, errCopy := io.Copy(f, dataReader)
			if errCopy != nil {
				panic(errCopy)
			}
			// errWrite := ioutil.WriteFile(saveFilePath, data, 0777)
			// if errWrite != nil {
			// 	panic(errWrite)
			// }
			fmt.Printf("video contents of %s is saved in %s.\nbitrate %s, fps %s, resolution %s\n", url, saveFilePath, s.Abr, s.Fps, s.Resolution)
			return
		}
	}

	fmt.Printf("no mp4 stream is not found for %s", url)
}
