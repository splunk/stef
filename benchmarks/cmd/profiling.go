package cmd

import (
	"log"
	"os"
	"runtime/pprof"
)

var CpuProfileFileName = ""

func SetupProfiling() (doneFunc func()) {
	if CpuProfileFileName != "" {
		f, err := os.Create(CpuProfileFileName)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}

		return func() {
			pprof.StopCPUProfile()
			if err := f.Close(); err != nil {
				log.Fatalf("Cannot close file %s: %v", CpuProfileFileName, err)
			}
		}
	}
	return func() {}
}
