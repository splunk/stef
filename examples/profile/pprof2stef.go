package profile

import (
	"fmt"
	"io"

	"github.com/google/pprof/profile"

	"github.com/splunk/stef/go/pkg"

	stefprofile "github.com/splunk/stef/examples/profile/internal/profile"
)

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

	// Convert each sample
	for _, srcSample := range prof.Sample {
		convertSample(srcSample, prof, &writer.Record)
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

// convertSample converts a pprof sample to STEF Sample
func convertSample(
	srcSample *profile.Sample, srcProf *profile.Profile,
	dst *stefprofile.Sample,
) {
	// Convert locations
	locations := dst.Locations()
	locations.EnsureLen(0)
	for _, loc := range srcSample.Location {
		stefLoc := convertLocation(loc)
		locations.Append(stefLoc)
	}

	// Convert values
	values := dst.Values()
	values.EnsureLen(0)
	for i, value := range srcSample.Value {
		stefValue := stefprofile.NewSampleValue()
		stefValue.SetVal(value)

		// Set value type if available
		if i < len(srcProf.SampleType) {
			valueType := stefValue.Type()
			valueType.SetType(srcProf.SampleType[i].Type)
			valueType.SetUnit(srcProf.SampleType[i].Unit)
		}

		values.Append(stefValue)
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
			numValue := dstLabelValue.Num()
			numValue.SetVal(value)
			if len(srcSample.NumUnit) > 0 {
				// NumUnit is a map in pprof, find the unit for this key
				if unit, exists := srcSample.NumUnit[key]; exists && len(unit) > 0 {
					numValue.SetUnit(unit[0])
				}
			}
			dstLabelValue.SetType(stefprofile.LabelValueTypeNum)
			labelIndex++
		}
	}
}

// convertLocation converts a pprof location to STEF Location
func convertLocation(loc *profile.Location) *stefprofile.Location {
	stefLoc := stefprofile.NewLocation()

	stefLoc.SetAddress(loc.Address)
	stefLoc.SetIsFolded(loc.IsFolded)

	// Convert mapping
	if loc.Mapping != nil {
		mapping := stefLoc.Mapping()
		mapping.SetMemoryStart(loc.Mapping.Start)
		mapping.SetMemoryLimit(loc.Mapping.Limit)
		mapping.SetFileOffset(loc.Mapping.Offset)
		mapping.SetFilename(loc.Mapping.File)
		mapping.SetBuildId(loc.Mapping.BuildID)
		mapping.SetHasFunctions(loc.Mapping.HasFunctions)
		mapping.SetHasFilenames(loc.Mapping.HasFilenames)
		mapping.SetHasLineNumbers(loc.Mapping.HasLineNumbers)
		mapping.SetHasInlineFrames(loc.Mapping.HasInlineFrames)
	}

	// Convert lines
	lines := stefLoc.Lines()
	for _, line := range loc.Line {
		stefLine := stefprofile.NewLine()
		stefLine.SetLine(uint64(line.Line))
		stefLine.SetColumn(uint64(line.Column))

		// Convert function
		if line.Function != nil {
			function := stefLine.Function()
			function.SetName(line.Function.Name)
			function.SetSystemName(line.Function.SystemName)
			function.SetFilename(line.Function.Filename)
			function.SetStartLine(uint64(line.Function.StartLine))
		}

		lines.Append(stefLine)
	}

	return stefLoc
}
