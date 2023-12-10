package main

import (
	"fmt"
	"slices"
	"strconv"
	"time"
)

type Statistics struct {
	MessagesInFile     int32 // number of lines in list file
	MessagesFound      int32 // number of messages returned from PO search call
	MessagesDownloaded int32 // number of messages processed

	PayloadsExtracted      int32 // number of individual files written including inside archives
	FilesWrittenToDisk     int32 // number of files written to disk
	NetworkBytesDownloaded int64 // number of raw bytes (HTTP)
	PayloadSize            int64 // number of raw bytes (payload except RAW)
	DiskBytesWritten       int64 // number of resulting bytes (files after compression)

}

var statistics Statistics

func runStatistics() *time.Ticker {
	statisticsTicker := time.NewTicker(time.Second)
	go UpdateStatistics(statisticsTicker)
	return statisticsTicker
}

func UpdateStatistics(ticker *time.Ticker) {
	for range ticker.C {
		if statistics.MessagesFound > 0 {
			fmt.Printf("Downloading messages... [%d / %d]\n", statistics.MessagesDownloaded, statistics.MessagesFound)
		} else {
			fmt.Printf("Searching for messages by message IDs [%d IDs]...\n", statistics.MessagesInFile)
		}
	}
}

func showEndCredits() {
	// to ensure you see "[100 / 100]" messages done
	fmt.Printf("Downloading messages... [%d / %d]\n", statistics.MessagesDownloaded, statistics.MessagesFound)
	fmt.Println("--------------")
	fmt.Println("---- DONE ----")
	fmt.Println("--------------")
	fmt.Printf("Processed messages    : %d / %d [%d Kb]\n", statistics.MessagesDownloaded, statistics.MessagesFound, statistics.NetworkBytesDownloaded/1024)
	fmt.Printf("Payloads extracted    : %d [%d Kb]\n", statistics.PayloadsExtracted, statistics.PayloadSize/1024)
	fmt.Printf("Files written to disk : %d [%d Kb]\n", statistics.FilesWrittenToDisk, statistics.DiskBytesWritten/1024)
}

func generateStatistics(c <-chan XIAdapterMessage) {
	// processor for -statsonly mode
	stats := make(map[string]int)
	keys := []string{}

	for msg := range c {

		// LOG
		for _, log := range msg.LogLocations.String {
			key := "LOG." + log
			v, ok := stats[key]
			if !ok {
				stats[key] = 1
				keys = append(keys, key)
			} else {
				stats[key] = v + 1
			}
		}

		// STAGING
		vers, _ := strconv.Atoi(msg.Version)
		for i := 0; i <= vers; i++ {
			key := fmt.Sprintf("STAGE.%d", i)
			v, ok := stats[key]
			if !ok {
				stats[key] = 1
				keys = append(keys, key)
			} else {
				stats[key] = v + 1
			}
		}

	}

	slices.Sort(keys)

	println(`------------------------------------------`)
	println(`--------- Message versions found ---------`)
	for _, key := range keys {
		fmt.Printf("%30s:   %d\n", key, stats[key])
	}
	println(`------------------------------------------`)
}
