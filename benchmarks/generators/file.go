package generators

import (
	"fmt"
	"log"
	"path"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/splunk/stef/benchmarks/testutils"
)

type File struct {
	FilePath  string
	BatchSize int
}

func (f *File) GetName() string {
	if f.BatchSize > 0 {
		return fmt.Sprintf("%s/%d", path.Base(f.FilePath), f.BatchSize)
	}
	return path.Base(f.FilePath)
}

func (f *File) Generate() pmetric.Metrics {
	all, err := testutils.ReadOTLPFile(f.FilePath)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", f.FilePath, err)
	}

	batchSize := f.BatchSize
	if batchSize == 0 {
		batchSize = all.ResourceMetrics().Len()
	}

	i := 0
	all.ResourceMetrics().RemoveIf(
		func(metrics pmetric.ResourceMetrics) bool {
			i++
			return i > batchSize
		},
	)

	return all
}
