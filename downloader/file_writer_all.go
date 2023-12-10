package main

import (
	"archive/zip"
	"fmt"
	"os"
	"sync/atomic"
)

///////////////// MODE ALL ///////////////

func FileWriterModeAll(options RuntimeConfiguration, version <-chan XIMessagePayloads) {

	newFilename := "export.zip"
	if options.MessageListFilename != "" {
		newFilename = options.MessageListFilename + ".zip"
	}

	file, err := os.Create(newFilename)
	if err != nil {
		fmt.Printf("Failed creating file [%s]: %s\n", newFilename, err)
		os.Exit(7)
	}
	defer file.Close()

	// Create a new zip archive.
	w := zip.NewWriter(file)
	defer w.Close()

	for entry := range version {

		for _, item := range entry.Parts {
			fullpath := fmt.Sprintf("%s/%s", entry.Folder, item.Filename)
			f, err := w.Create(fullpath)
			if err != nil {
				fmt.Printf("Failed writing file [%s] to ZIP: %s\n", fullpath, err)
				continue
			}

			_, err = f.Write(item.Contents)
			if err != nil {
				fmt.Printf("Failed writing file [%s] to ZIP: %s\n", fullpath, err)
				continue
			}
		}
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		fmt.Printf("Fail on ZIP file close [%s]: %s\n", newFilename, err)
		return
	}

	stats, _ := file.Stat()
	atomic.AddInt64(&statistics.DiskBytesWritten, stats.Size())
	atomic.AddInt32(&statistics.FilesWrittenToDisk, 1)
}
