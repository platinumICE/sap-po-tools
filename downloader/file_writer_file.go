package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"sync/atomic"
)

///////// MODE: FILE /////////////////

func FileWriterModeFile(options RuntimeConfiguration, version <-chan XIMessagePayloads) {

	for entry := range version {
		path := createPath(entry.Folder)
		for _, item := range entry.Parts {

			bytesDisk, ok := FileWriterWriteGZIP(options, item, path)

			atomic.AddInt32(&statistics.FilesWrittenToDisk, ok)
			atomic.AddInt64(&statistics.DiskBytesWritten, bytesDisk)
		}
	}

}

// return bytes written to disk, payload size and number of files
func FileWriterWriteGZIP(options RuntimeConfiguration, item XIPayload, path string) (int64, int32) {

	newFilename := fmt.Sprintf("%s/%s.gz", path, item.Filename)

	file, err := os.Create(newFilename)
	if err != nil {
		fmt.Printf("Failed creating file [%s]: %s\n", newFilename, err)
		return 0, 0
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)

	_, err = gzipWriter.Write(item.Contents)
	if err != nil {
		fmt.Printf("Failed writing to file [%s]: %s\n", newFilename, err)
		return 0, 0
	}

	err = gzipWriter.Close()
	if err != nil {
		fmt.Printf("Failed writing to file [%s]: %s\n", newFilename, err)
		return 0, 0
	}

	stats, _ := file.Stat()
	return int64(stats.Size()), 1
}
