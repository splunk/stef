package profile

import (
	"fmt"
	"io"

	"github.com/google/pprof/profile"
	"modernc.org/b/v2"

	stefprofile "github.com/splunk/stef/examples/profile/internal/profile"
	"github.com/splunk/stef/go/pkg"
)

// convertStefMetadata converts STEF ProfileMetadata to pprof Profile metadata fields
func convertStefMetadata(stefMetadata *stefprofile.ProfileMetadata, prof *profile.Profile) {
	prof.DropFrames = stefMetadata.DropFrames()
	prof.KeepFrames = stefMetadata.KeepFrames()
	prof.TimeNanos = stefMetadata.TimeNanos()
	prof.DurationNanos = stefMetadata.DurationNanos()
	prof.Period = stefMetadata.Period()

	// Convert period type
	stefPeriodType := stefMetadata.PeriodType()
	if stefPeriodType != nil {
		prof.PeriodType = &profile.ValueType{
			Type: stefPeriodType.Type(),
			Unit: stefPeriodType.Unit(),
		}
	}

	// Convert default sample type
	stefDefaultType := stefMetadata.DefaultSampleType()
	if stefDefaultType != nil {
		prof.DefaultSampleType = stefDefaultType.Type()
	}

	// Convert comments
	stefComments := stefMetadata.Comments()
	for i := 0; i < stefComments.Len(); i++ {
		prof.Comments = append(prof.Comments, stefComments.At(i))
	}
}

// convertStefToPprof converts a STEF format profile to pprof format
func convertStefToPprof(src io.Reader) (*profile.Profile, error) {
	reader, err := stefprofile.NewSampleReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to create sample reader: %w", err)
	}

	prof := &profile.Profile{}

	// Trees to track unique elements and assign IDs
	mappingMap := b.TreeNew[*stefprofile.Mapping, *profile.Mapping](stefprofile.CmpMapping)
	functionMap := b.TreeNew[*stefprofile.Function, *profile.Function](stefprofile.CmpFunction)
	locationMap := b.TreeNew[*stefprofile.Location, *profile.Location](stefprofile.CmpLocation)

	var mappingID, functionID, locationID uint64 = 1, 1, 1
	var metadata *stefprofile.ProfileMetadata

	// Read all samples and build the profile
	for {
		err := reader.Read(pkg.ReadOptions{})
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read sample: %w", err)
		}

		// Store metadata from first sample
		if metadata == nil {
			metadata = reader.Record.Metadata()
			convertStefMetadata(metadata, prof)
		}

		// Create pprof sample
		pprofSample := &profile.Sample{}

		// Process locations
		stefLocations := reader.Record.Locations()
		for i := 0; i < stefLocations.Len(); i++ {
			stefLocation := stefLocations.At(i)

			var pprofLocation *profile.Location
			if existing, exists := locationMap.Get(stefLocation); exists {
				pprofLocation = existing
			} else {
				// Create new location
				pprofLocation = &profile.Location{
					ID:       locationID,
					Address:  stefLocation.Address(),
					IsFolded: stefLocation.IsFolded(),
				}
				locationID++

				// Process mapping
				if stefMapping := stefLocation.Mapping(); stefMapping != nil {
					if existing, exists := mappingMap.Get(stefMapping); exists {
						pprofLocation.Mapping = existing
					} else {
						pprofMapping := &profile.Mapping{
							ID:              mappingID,
							Start:           stefMapping.MemoryStart(),
							Limit:           stefMapping.MemoryLimit(),
							Offset:          stefMapping.FileOffset(),
							File:            stefMapping.Filename(),
							BuildID:         stefMapping.BuildId(),
							HasFunctions:    stefMapping.HasFunctions(),
							HasFilenames:    stefMapping.HasFilenames(),
							HasLineNumbers:  stefMapping.HasLineNumbers(),
							HasInlineFrames: stefMapping.HasInlineFrames(),
						}
						mappingID++
						prof.Mapping = append(prof.Mapping, pprofMapping)
						mappingMap.Set(stefMapping, pprofMapping)
						pprofLocation.Mapping = pprofMapping
					}
				}

				// Process lines
				stefLines := stefLocation.Lines()
				for j := 0; j < stefLines.Len(); j++ {
					stefLine := stefLines.At(j)
					pprofLine := profile.Line{
						Line:   int64(stefLine.Line()),
						Column: int64(stefLine.Column()),
					}

					// Process function
					if stefFunction := stefLine.Function(); stefFunction != nil {
						if existing, exists := functionMap.Get(stefFunction); exists {
							pprofLine.Function = existing
						} else {
							pprofFunction := &profile.Function{
								ID:         functionID,
								Name:       stefFunction.Name(),
								SystemName: stefFunction.SystemName(),
								Filename:   stefFunction.Filename(),
								StartLine:  int64(stefFunction.StartLine()),
							}
							functionID++
							prof.Function = append(prof.Function, pprofFunction)
							functionMap.Set(stefFunction, pprofFunction)
							pprofLine.Function = pprofFunction
						}
					}

					pprofLocation.Line = append(pprofLocation.Line, pprofLine)
				}

				prof.Location = append(prof.Location, pprofLocation)
				locationMap.Set(stefLocation, pprofLocation)
			}

			pprofSample.Location = append(pprofSample.Location, pprofLocation)
		}

		// Process values
		stefValues := reader.Record.Values()
		for i := 0; i < stefValues.Len(); i++ {
			stefValue := stefValues.At(i)
			pprofSample.Value = append(pprofSample.Value, stefValue.Val())
		}

		// Process labels
		pprofSample.Label = make(map[string][]string)
		pprofSample.NumLabel = make(map[string][]int64)
		pprofSample.NumUnit = make(map[string][]string)

		stefLabels := reader.Record.Labels()
		for i := 0; i < stefLabels.Len(); i++ {
			key := stefLabels.At(i).Key()
			stefLabelValue := stefLabels.At(i).Value()

			if stefLabelValue.Type() == stefprofile.LabelValueTypeStr {
				value := stefLabelValue.Str()
				pprofSample.Label[key] = append(pprofSample.Label[key], value)
			} else if stefLabelValue.Type() == stefprofile.LabelValueTypeNum {
				stefNumValue := stefLabelValue.Num()
				value := stefNumValue.Val()
				unit := stefNumValue.Unit()

				pprofSample.NumLabel[key] = append(pprofSample.NumLabel[key], value)
				// Always add the unit, even if it's empty string, to match pprof behavior
				pprofSample.NumUnit[key] = append(pprofSample.NumUnit[key], unit)
			}
		}

		prof.Sample = append(prof.Sample, pprofSample)

		// Build sample types from first sample's values if not already done
		if len(prof.SampleType) == 0 && stefValues.Len() > 0 {
			for i := 0; i < stefValues.Len(); i++ {
				stefValue := stefValues.At(i)
				stefType := stefValue.Type()
				prof.SampleType = append(
					prof.SampleType, &profile.ValueType{
						Type: stefType.Type(),
						Unit: stefType.Unit(),
					},
				)
			}
		}
	}

	return prof, nil
}
