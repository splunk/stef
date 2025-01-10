package tests

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"modernc.org/b/v2"

	"github.com/splunk/stef/benchmarks/testutils"
	"github.com/splunk/stef/stef-go/pkg"
	"github.com/splunk/stef/stef-otel/oteltef"
	otlpconvert "github.com/splunk/stef/stef-pdata/metrics"
)

/*
func TestTreeConverter(t *testing.T) {
	source := generators.File{
		FilePath: "testdata/hipstershop.pb.zst",
	}
	otlpData := source.Generate()

	converter := otlpconvert.NewOtlpToSortedTree()
	sorted, err := converter.FromOtlp(otlpData.ResourceMetrics())
	require.NoError(t, err)

	srcDataPointCount := 0
	srcResources := b.TreeNew[*oteltef.Resource, bool](oteltef.CmpResource)
	srcScopes := b.TreeNew[*oteltef.Scope, bool](oteltef.CmpScope)

	encoder := &anyvalue.Encoder{}
	for i := 0; i < otlpData.ResourceMetrics().Len(); i++ {
		rm := otlpData.ResourceMetrics().At(i)
		srcResources.Set(otlptools.ResourceToStef(rm.Resource(), rm.SchemaUrl(), encoder), true)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sms := rm.ScopeMetrics().At(j)
			srcScopes.Set(otlptools.ScopeToStef(sms.Scope(), sms.SchemaUrl(), encoder), true)
			for k := 0; k < sms.Metrics().Len(); k++ {
				metric := sms.Metrics().At(k)
				if metric.Type() == pmetric.MetricTypeGauge {
					srcDataPointCount += metric.Gauge().DataPoints().Len()
				} else if metric.Type() == pmetric.MetricTypeSum {
					srcDataPointCount += metric.Sum().DataPoints().Len()
				} else {
					log.Fatalf("Unsupported metric type: %v\n", metric.Name())
				}
			}
		}
	}

	destDataPointCount := 0
	destResources := b.TreeNew[*types.Resource, bool](types.CmpResource)
	destScopes := b.TreeNew[*types.Scope, bool](types.CmpScope)

	err = sorted.Iter(
		func(metric *types.Metric, byMetric *sortedbymetric.ByMetric) error {
			err := byMetric.Iter(
				func(resource *types.Resource, byResource *sortedbymetric.ByResource) error {
					destResources.Set(resource, true)
					err := byResource.Iter(
						func(scope *types.Scope, byScope *sortedbymetric.ByScope) error {
							destScopes.Set(scope, true)
							err := byScope.Iter(
								func(attrs types.AttrList, values *types.TimedPoints) error {
									destDataPointCount += len(values.Values())
									return nil
								},
							)
							return err
						},
					)
					return err
				},
			)
			return err
		},
	)
	require.NoError(t, err)

	require.True(t, equalMaps(srcResources, destResources))
	require.True(t, equalMaps(srcScopes, destScopes))
	require.EqualValues(t, srcDataPointCount, destDataPointCount)
}
*/

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
	otlpData, err := testutils.ReadOTLPFile("testdata/oteldemo-with-histogram.otlp.zst", true)
	require.NoError(t, err)

	converter := otlpconvert.NewOtlpToSortedTree()
	sorted, err := converter.FromOtlp(otlpData.ResourceMetrics())
	require.NoError(t, err)

	outputBuf := &pkg.MemChunkWriter{}
	writer, err := oteltef.NewMetricsWriter(outputBuf, pkg.WriterOptions{})
	require.NoError(t, err)

	err = sorted.ToTef(writer)
	require.NoError(t, err)

	sorted = nil

	err = writer.Flush()
	require.NoError(t, err)
}
