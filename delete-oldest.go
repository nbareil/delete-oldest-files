package main

import (
	"fmt"
	"log"
	"os"
	"math"
	"sort"
	"syscall"
)

const captureDirectory = "/tmp/removal"
const FREE_THRESHOLD = 21


type SortedFileInfo struct {
	entries []os.FileInfo
}

// Methods required by sort.Interface.
func (s SortedFileInfo) Len() int {
	return len(s.entries)
}

func (s SortedFileInfo) Less(i, j int) bool {
	return s.entries[i].ModTime().Unix() < s.entries[j].ModTime().Unix()
}

func (s SortedFileInfo) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
}


func getFreeSpace(dir string) float64 {
	var buf syscall.Statfs_t

	err := syscall.Statfs(dir, &buf)
	if err != nil {
		panic("Cannot get information about FS")
	}

	if buf.Blocks == 0 {
		return 0
	}

	percent := math.Float64frombits(buf.Bfree) / math.Float64frombits(buf.Blocks) * 100

	return percent
}

func removeOldestFiles(dir string) {
	directoryfd, err := os.Open(dir)
	if err != nil {
		panic("Cannot open directory")
	}
	defer directoryfd.Close()

	entries, err := directoryfd.Readdir(-1)
	sortedEntries := SortedFileInfo{entries: entries}
	sort.Sort(sortedEntries)

	for _, entry := range sortedEntries.entries {
		path := captureDirectory + "/" + entry.Name()
		fmt.Printf("unlink %s\n", path)
		err := os.Remove(path)
		if err != nil {
			log.Printf("Cannot remove %s: %s\n", path, err)
			continue
		}
		if getFreeSpace(dir) > FREE_THRESHOLD {
			break
		}
	}
}

func main() {
	free := getFreeSpace(captureDirectory)
	if free < FREE_THRESHOLD {
		removeOldestFiles(captureDirectory)
	} else {
		fmt.Printf("Doing nothing [%.2f]\n", free)
	}

}
