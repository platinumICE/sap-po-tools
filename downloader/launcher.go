package main

import (
	"fmt"
	"os"
	"sync"
)

const (
	ToolVersion string = "1.0.0 (2023-12-10)"
	ToolRepo           = "https://github.com/platinumICE/sap-po-tools"
	ToolAuthor         = "Marat Bareev"
)

var wgDownloaders, wgUnpackers, wgWriters sync.WaitGroup

func main() {
	fmt.Println(`------------------------------------------`)
	fmt.Printf("SAP PO Tools : Downloader v%s\n", ToolVersion)
	fmt.Printf("Public Repo  : %s\n", ToolRepo)
	fmt.Printf("Author       : %s\n", ToolAuthor)
	fmt.Println(`------------------------------------------`)

	runtime_config, err := ParseLaunchOptions()
	if err != nil {
		fmt.Printf("Error parsing command-line: %s\n", err)
		os.Exit(1)
	}

	connection_config, err := GetConnectionConfig(runtime_config)
	if err != nil {
		fmt.Printf("Error reading connection file: %s\n", err)
		os.Exit(2)
	}

	idList, err := prepareMessageList(runtime_config)
	if err != nil {
		fmt.Println("Error processing Message ID list:", err)
		os.Exit(3)
	}

	if len(idList) == 0 {
		fmt.Println("Error processing Message ID list: list is empty")
		os.Exit(3)
	}

	err = prepareFileWriter(runtime_config, connection_config)
	if err != nil {
		fmt.Println("Error preparing output directory:", err)
		os.Exit(4)
	}

	initiateHTTPClient(runtime_config)

	statsTicker := runStatistics()

	messageChannel := make(chan XIAdapterMessage, 10000)
	err = searchMessages(runtime_config, connection_config, idList, messageChannel)
	if err != nil {
		fmt.Println("Error processing Message ID list:", err)
		os.Exit(6)
	}

	if runtime_config.StatisticsOnly {
		statsTicker.Stop()
		generateStatistics(messageChannel)
	} else {
		wgWriters.Add(1)
		payloadChannel := make(chan XIMessagePayloads, 100)
		go FileWriter(runtime_config, payloadChannel)

		wgUnpackers.Add(1)
		versionChannel := make(chan XIMessageVersion, 100)
		go Unpacker(runtime_config, versionChannel, payloadChannel)

		for i := 0; i < runtime_config.DownloadThreads; i++ {
			wgDownloaders.Add(1)
			go Downloader(runtime_config, connection_config, messageChannel, versionChannel)
		}

		wgDownloaders.Wait()
		close(versionChannel)

		wgUnpackers.Wait()
		close(payloadChannel)

		wgWriters.Wait()

		statsTicker.Stop()
		showEndCredits()
		openTargetDirectory(runtime_config)
	}

}
