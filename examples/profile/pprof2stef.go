package profile

import (
	"fmt"
	"io"

	"github.com/google/pprof/profile"

	"github.com/splunk/stef/go/pkg"

	stefprofile "github.com/splunk/stef/examples/profile/internal/profile"
)

type pprof2stef struct {
	mappings  []stefprofile.Mapping
	functions []stefprofile.Function
	locations []stefprofile.Location
}

func (c *pprof2stef) convertMappings(prof *profile.Profile) {
	c.mappings = make([]stefprofile.Mapping, len(prof.Mapping)+1)
	for _, src := range prof.Mapping {
		dst := &c.mappings[src.ID]
		dst.Init()
		dst.SetMemoryStart(src.Start)
		dst.SetMemoryLimit(src.Limit)
		dst.SetFileOffset(src.Offset)
		dst.SetFilename(src.File)
		dst.SetBuildId(src.BuildID)
		dst.SetHasFunctions(src.HasFunctions)
		dst.SetHasFilenames(src.HasFilenames)
		dst.SetHasLineNumbers(src.HasLineNumbers)
		dst.SetHasInlineFrames(src.HasInlineFrames)
		dst.Freeze()
	}
}

func (c *pprof2stef) convertFunctions(prof *profile.Profile) {
	c.functions = make([]stefprofile.Function, len(prof.Function)+1)
	for _, src := range prof.Function {
		dst := &c.functions[src.ID]
		dst.Init()
		dst.SetName(src.Name)
		dst.SetSystemName(src.SystemName)
		dst.SetFilename(src.Filename)
		dst.SetStartLine(uint64(src.StartLine))
		dst.Freeze()
	}
}

func (c *pprof2stef) convertLocations(prof *profile.Profile) {
	c.locations = make([]stefprofile.Location, len(prof.Location)+1)
	for _, src := range prof.Location {
		dst := &c.locations[src.ID]
		dst.Init()

		dst.SetAddress(src.Address)
		dst.SetIsFolded(src.IsFolded)

		// Convert mapping
		if src.Mapping != nil {
			dst.SetMapping(&c.mappings[src.Mapping.ID])
		}

		// Convert lines
		lines := dst.Lines()
		lines.EnsureLen(len(src.Line))
		for i, line := range src.Line {
			stefLine := lines.At(i)
			stefLine.SetLine(uint64(line.Line))
			stefLine.SetColumn(uint64(line.Column))

			// Convert function
			if line.Function != nil {
				stefLine.SetFunction(&c.functions[line.Function.ID])
			}
		}
		dst.Freeze()
	}
}

func (c *pprof2stef) convertSample(srcSample *profile.Sample, srcProf *profile.Profile, dst *stefprofile.Sample) {
	// Convert locations
	locations := dst.Locations()
	locations.EnsureLen(0)
	for _, loc := range srcSample.Location {
		locations.Append(&c.locations[loc.ID])
	}

	// Convert values
	values := dst.Values()
	values.EnsureLen(len(srcSample.Value))
	for i, value := range srcSample.Value {
		stefValue := values.At(i)
		stefValue.SetVal(value)

		// Set value type if available
		if i < len(srcProf.SampleType) {
			valueType := &stefprofile.SampleValueType{}
			valueType.Init()
			valueType.SetType(srcProf.SampleType[i].Type)
			valueType.SetUnit(srcProf.SampleType[i].Unit)
			valueType.Freeze()
			stefValue.SetType(valueType)
		}
	}

	// Convert labels
	dstLabels := dst.Labels()
	labelIndex := 0

	// Count total number of labels first
	totalLabels := 0
	for _, values := range srcSample.Label {
		totalLabels += len(values)
	}
	for _, values := range srcSample.NumLabel {
		totalLabels += len(values)
	}

	// Ensure the labels multimap has enough space
	dstLabels.EnsureLen(totalLabels)

	// Add string labels
	for key, values := range srcSample.Label {
		for _, value := range values {
			dstLabels.SetKey(labelIndex, key)
			dstLabelValue := dstLabels.At(labelIndex).Value()
			dstLabelValue.SetStr(value)
			labelIndex++
		}
	}

	// Add numeric labels
	for key, values := range srcSample.NumLabel {
		for _, value := range values {
			dstLabels.SetKey(labelIndex, key)
			dstLabelValue := dstLabels.At(labelIndex).Value()
			dstLabelValue.SetType(stefprofile.LabelValueTypeNum)
			numValue := dstLabelValue.Num()
			numValue.SetVal(value)
			if len(srcSample.NumUnit) > 0 {
				// NumUnit is a map in pprof, find the unit for this key
				if unit, exists := srcSample.NumUnit[key]; exists && len(unit) > 0 {
					numValue.SetUnit(unit[0])
				}
			}
			labelIndex++
		}
	}
}

// convertPprofToStef converts a pprof Profile to STEF format using SampleWriter
func convertPprofToStef(prof *profile.Profile, dst io.Writer) error {
	// Create a chunk writer for the destination
	chunkWriter := pkg.NewWrapChunkWriter(dst)

	// Create sample writer
	writer, err := stefprofile.NewSampleWriter(chunkWriter, pkg.WriterOptions{})
	if err != nil {
		return fmt.Errorf("failed to create srcSample writer: %w", err)
	}

	// Convert profile metadata once
	convertProfileMetadata(prof, writer.Record.Metadata())

	c := pprof2stef{}
	c.convertMappings(prof)
	c.convertFunctions(prof)
	c.convertLocations(prof)

	// Convert each sample
	for _, srcSample := range prof.Sample {
		c.convertSample(srcSample, prof, &writer.Record)
		if err := writer.Write(); err != nil {
			return fmt.Errorf("failed to write srcSample: %w", err)
		}
	}

	// Flush the writer to ensure all data is written
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// convertProfileMetadata converts pprof profile metadata to STEF ProfileMetadata
func convertProfileMetadata(src *profile.Profile, dst *stefprofile.ProfileMetadata) {
	dst.SetDropFrames(src.DropFrames)
	dst.SetKeepFrames(src.KeepFrames)

	dst.SetTimeNanos(src.TimeNanos)
	dst.SetDurationNanos(src.DurationNanos)
	dst.SetPeriod(src.Period)

	// Convert period type - access the actual SampleValueType pointer
	if src.PeriodType != nil {
		periodType := dst.PeriodType()
		periodType.SetType(src.PeriodType.Type)
		periodType.SetUnit(src.PeriodType.Unit)
	}

	// Convert default sample type
	if src.DefaultSampleType != "" {
		// Find the matching sample type in src.SampleType
		for _, st := range src.SampleType {
			if st.Type == src.DefaultSampleType {
				defaultSampleType := dst.DefaultSampleType()
				defaultSampleType.SetType(st.Type)
				defaultSampleType.SetUnit(st.Unit)
				break
			}
		}
	}

	// Convert comments
	if len(src.Comments) > 0 {
		comments := dst.Comments()
		comments.EnsureLen(0)
		for _, comment := range src.Comments {
			comments.Append(comment)
		}
	}
}
