package tests

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"modernc.org/b/v2"

	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/go/otel/otelstef"
	"github.com/splunk/stef/go/pdata/metrics/sortedbymetric"
	"github.com/splunk/stef/go/pkg"
)

func equalMaps[K any](t1, t2 *b.Tree[K, bool]) bool {
	return mapContains(t1, t2) && mapContains(t2, t1)
}

func mapContains[K any](searchFor, inTree *b.Tree[K, bool]) bool {
	iter, err := searchFor.SeekFirst()
	if err != nil {
		return true
	}
	for {
		k, _, err := iter.Next()
		if err == io.EOF {
			break
		}
		_, exists := inTree.Get(k)
		if !exists {
			return false
		}
	}
	return true
}

func TestConvertFromOTLP(t *testing.T) {
	otlpData, err := testutils.ReadOTLPFile("testdata/hipstershop-otelmetrics.zst")
	require.NoError(t, err)

	sorted, err := sortedbymetric.OtlpToSortedTree(otlpData)
	require.NoError(t, err)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := otelstef.NewMetricsWriter(outputBuf, pkg.WriterOptions{})
	require.NoError(t, err)

	err = sorted.ToStef(writer)
	require.NoError(t, err)

	sorted = nil

	err = writer.Flush()
	require.NoError(t, err)
}
