package cmd

import (
	"log"
	"os"
	"path"
	"strings"
)

type FileEntry struct {
	FullPath string
	RelPath  string
	Size     int64
}

func ListInputFiles(inputPath string, stats *CmdStats) []FileEntry {
	inputPath = path.Clean(inputPath)
	return listInputFiles(inputPath, "", stats)
}

func listInputFiles(inputPath string, rootDir string, stats *CmdStats) []FileEntry {
	fstat, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("Cannot read file or directory %s: %v", inputPath, err)
	}
	var files []FileEntry
	if fstat.IsDir() {
		if rootDir == "" {
			rootDir = inputPath + "/"
		}
		dirEntries, err := os.ReadDir(inputPath)
		if err != nil {
			log.Fatalf("Cannot read directory %s: %v", inputPath, err)
		}
		for _, e := range dirEntries {
			filePath := path.Join(inputPath, e.Name())
			files = append(files, listInputFiles(filePath, rootDir, stats)...)
		}
	} else {
		var relPath string
		if rootDir == "" {
			relPath = path.Base(inputPath)
		} else {
			var found bool
			relPath, found = strings.CutPrefix(inputPath, rootDir)
			if !found {
				log.Fatal("Cannot list input files")
			}
		}
		files = append(
			files, FileEntry{
				FullPath: inputPath,
				RelPath:  relPath,
				Size:     fstat.Size(),
			},
		)
		stats.InputFileCount += 1
		stats.InputBytes += uint64(fstat.Size())
	}

	return files
}

func NameNoExt(fileName string) string {
	return fileName[:len(fileName)-len(path.Ext(fileName))]
}
