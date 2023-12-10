package main

import (
	"fmt"
	"os"
	"sync/atomic"
)

///////// MODE: NONE /////////////////

func FileWriterModeNone(options RuntimeConfiguration, version <-chan XIMessagePayloads) {

	for entry := range version {

		path := createPath(entry.Folder)

		for _, item := range entry.Parts {
			newFilename := fmt.Sprintf("%s/%s", path, item.Filename)
			err := os.WriteFile(newFilename, item.Contents, 0666)
			if err != nil {
				fmt.Printf("Error writing file [%s] to disk: %s", item.Filename, err)
				continue
			}

			atomic.AddInt32(&statistics.FilesWrittenToDisk, 1)
			atomic.AddInt64(&statistics.DiskBytesWritten, int64(len(item.Contents)))
		}
	}

}
