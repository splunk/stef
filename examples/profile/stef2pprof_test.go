package profile

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	pprof "github.com/google/pprof/profile"
	"github.com/stretchr/testify/require"
)

// isEqualPprof compares two pprof profiles for equality
func isEqualPprof(prof1, prof2 *pprof.Profile) (bool, string) {

	if prof1 == nil && prof2 == nil {
		return true, ""
	}
	if prof1 == nil || prof2 == nil {
		return false, "one profile is nil"
	}

	// Compare basic metadata
	if prof1.DropFrames != prof2.DropFrames {
		return false, fmt.Sprintf("DropFrames differ: %q vs %q", prof1.DropFrames, prof2.DropFrames)
	}
	if prof1.KeepFrames != prof2.KeepFrames {
		return false, fmt.Sprintf("KeepFrames differ: %q vs %q", prof1.KeepFrames, prof2.KeepFrames)
	}
	if prof1.TimeNanos != prof2.TimeNanos {
		return false, fmt.Sprintf("TimeNanos differ: %d vs %d", prof1.TimeNanos, prof2.TimeNanos)
	}
	if prof1.DurationNanos != prof2.DurationNanos {
		return false, fmt.Sprintf("DurationNanos differ: %d vs %d", prof1.DurationNanos, prof2.DurationNanos)
	}
	if prof1.Period != prof2.Period {
		return false, fmt.Sprintf("Period differ: %d vs %d", prof1.Period, prof2.Period)
	}
	if prof1.DefaultSampleType != prof2.DefaultSampleType {
		return false, fmt.Sprintf(
			"DefaultSampleType differ: %q vs %q", prof1.DefaultSampleType, prof2.DefaultSampleType,
		)
	}

	// Compare period type
	if !isEqualValueType(prof1.PeriodType, prof2.PeriodType) {
		return false, fmt.Sprintf("PeriodType differ: %+v vs %+v", prof1.PeriodType, prof2.PeriodType)
	}

	// Compare comments
	if !reflect.DeepEqual(prof1.Comments, prof2.Comments) {
		return false, fmt.Sprintf("Comments differ: %v vs %v", prof1.Comments, prof2.Comments)
	}

	// Compare sample types
	if len(prof1.SampleType) != len(prof2.SampleType) {
		return false, fmt.Sprintf("SampleType length differs: %d vs %d", len(prof1.SampleType), len(prof2.SampleType))
	}
	for i, st1 := range prof1.SampleType {
		st2 := prof2.SampleType[i]
		if !isEqualValueType(st1, st2) {
			return false, fmt.Sprintf("SampleType[%d] differs: %+v vs %+v", i, st1, st2)
		}
	}

	// Compare mappings
	if len(prof1.Mapping) != len(prof2.Mapping) {
		return false, fmt.Sprintf("Mapping count differs: %d vs %d", len(prof1.Mapping), len(prof2.Mapping))
	}

	// Sort mappings by ID for comparison
	sortMappingsByID(prof1.Mapping)
	sortMappingsByID(prof2.Mapping)

	for i, m1 := range prof1.Mapping {
		m2 := prof2.Mapping[i]
		if equal, msg := isEqualMapping(m1, m2); !equal {
			return false, fmt.Sprintf("Mapping[%d] differs: %s", i, msg)
		}
	}

	// Compare functions
	if len(prof1.Function) != len(prof2.Function) {
		return false, fmt.Sprintf("Function count differs: %d vs %d", len(prof1.Function), len(prof2.Function))
	}

	// Sort functions by ID for comparison
	sortFunctionsByID(prof1.Function)
	sortFunctionsByID(prof2.Function)

	for i, f1 := range prof1.Function {
		f2 := prof2.Function[i]
		if equal, msg := isEqualFunction(f1, f2); !equal {
			return false, fmt.Sprintf("Function[%d] differs: %s", i, msg)
		}
	}

	// Compare locations
	if len(prof1.Location) != len(prof2.Location) {
		return false, fmt.Sprintf("Location count differs: %d vs %d", len(prof1.Location), len(prof2.Location))
	}

	// Sort locations by ID for comparison
	sortLocationsByID(prof1.Location)
	sortLocationsByID(prof2.Location)

	for i, l1 := range prof1.Location {
		l2 := prof2.Location[i]
		if equal, msg := isEqualLocation(l1, l2); !equal {
			return false, fmt.Sprintf("Location[%d] differs: %s", i, msg)
		}
	}

	// Compare samples
	if len(prof1.Sample) != len(prof2.Sample) {
		return false, fmt.Sprintf("Sample count differs: %d vs %d", len(prof1.Sample), len(prof2.Sample))
	}

	for i, s1 := range prof1.Sample {
		s2 := prof2.Sample[i]
		if equal, msg := isEqualSample(s1, s2); !equal {
			return false, fmt.Sprintf("Sample[%d] differs: %s", i, msg)
		}
	}

	return true, ""
}

// Helper functions for comparison
func isEqualValueType(vt1, vt2 *pprof.ValueType) bool {
	if vt1 == nil && vt2 == nil {
		return true
	}
	if vt1 == nil || vt2 == nil {
		return false
	}
	return vt1.Type == vt2.Type && vt1.Unit == vt2.Unit
}

func isEqualMapping(m1, m2 *pprof.Mapping) (bool, string) {
	str := cmp.Diff(m1, m2, cmp.AllowUnexported(pprof.Mapping{}))
	if str != "" {
		return false, str
	}
	return true, ""
}

func isEqualFunction(f1, f2 *pprof.Function) (bool, string) {
	str := cmp.Diff(f1, f2, cmp.AllowUnexported(pprof.Function{}))
	if str != "" {
		return false, str
	}
	return true, ""
}

func isEqualLocation(l1, l2 *pprof.Location) (bool, string) {
	if l1.Address != l2.Address {
		return false, fmt.Sprintf("Address differs: %d vs %d", l1.Address, l2.Address)
	}
	if l1.IsFolded != l2.IsFolded {
		return false, fmt.Sprintf("IsFolded differs: %t vs %t", l1.IsFolded, l2.IsFolded)
	}

	// Compare mapping references by content, not pointer
	if !isEqualMappingRef(l1.Mapping, l2.Mapping) {
		return false, "Mapping reference differs"
	}

	// Compare lines
	if len(l1.Line) != len(l2.Line) {
		return false, fmt.Sprintf("Line count differs: %d vs %d", len(l1.Line), len(l2.Line))
	}
	for i, line1 := range l1.Line {
		line2 := l2.Line[i]
		if equal, msg := isEqualLine(line1, line2); !equal {
			return false, fmt.Sprintf("Line[%d] differs: %s", i, msg)
		}
	}

	return true, ""
}

func isEqualMappingRef(m1, m2 *pprof.Mapping) bool {
	if m1 == nil && m2 == nil {
		return true
	}
	if m1 == nil || m2 == nil {
		return false
	}
	// Compare by content, not pointer
	equal, _ := isEqualMapping(m1, m2)
	return equal
}

func isEqualLine(l1, l2 pprof.Line) (bool, string) {
	str := cmp.Diff(l1, l2, cmp.AllowUnexported(pprof.Line{}, pprof.Function{}))
	if str != "" {
		return false, str
	}

	// Compare function references by content, not pointer
	if !isEqualFunctionRef(l1.Function, l2.Function) {
		return false, "Function reference differs"
	}

	return true, ""
}

func isEqualFunctionRef(f1, f2 *pprof.Function) bool {
	if f1 == nil && f2 == nil {
		return true
	}
	if f1 == nil || f2 == nil {
		return false
	}
	// Compare by content, not pointer
	equal, _ := isEqualFunction(f1, f2)
	return equal
}

func isEqualSample(s1, s2 *pprof.Sample) (bool, string) {
	// Compare values
	if !reflect.DeepEqual(s1.Value, s2.Value) {
		return false, fmt.Sprintf("Values differ: %v vs %v", s1.Value, s2.Value)
	}

	// Compare location references by content
	if len(s1.Location) != len(s2.Location) {
		return false, fmt.Sprintf("Location count differs: %d vs %d", len(s1.Location), len(s2.Location))
	}
	for i, loc1 := range s1.Location {
		loc2 := s2.Location[i]
		if !isEqualLocationRef(loc1, loc2) {
			return false, fmt.Sprintf("Location[%d] reference differs", i)
		}
	}

	// Compare labels - handle empty maps correctly
	if !isEqualStringMapSlice(s1.Label, s2.Label) {
		return false, fmt.Sprintf("Labels differ: %v vs %v", s1.Label, s2.Label)
	}
	if !isEqualInt64MapSlice(s1.NumLabel, s2.NumLabel) {
		return false, fmt.Sprintf("NumLabels differ: %v vs %v", s1.NumLabel, s2.NumLabel)
	}
	if !isEqualStringMapSlice(s1.NumUnit, s2.NumUnit) {
		return false, fmt.Sprintf("NumUnit differ: %v vs %v", s1.NumUnit, s2.NumUnit)
	}

	return true, ""
}

func isEqualLocationRef(l1, l2 *pprof.Location) bool {
	if l1 == nil && l2 == nil {
		return true
	}
	if l1 == nil || l2 == nil {
		return false
	}
	// Compare by content, not pointer
	equal, _ := isEqualLocation(l1, l2)
	return equal
}

// Helper functions for comparing maps with proper empty map handling
func isEqualStringMapSlice(m1, m2 map[string][]string) bool {
	// Handle nil vs empty map cases
	if len(m1) == 0 && len(m2) == 0 {
		return true
	}
	return reflect.DeepEqual(m1, m2)
}

func isEqualInt64MapSlice(m1, m2 map[string][]int64) bool {
	// Handle nil vs empty map cases
	if len(m1) == 0 && len(m2) == 0 {
		return true
	}
	return reflect.DeepEqual(m1, m2)
}

// Sorting functions for deterministic comparison
func sortMappingsByID(mappings []*pprof.Mapping) {
	sort.Slice(
		mappings, func(i, j int) bool {
			return mappings[i].ID < mappings[j].ID
		},
	)
}

func sortFunctionsByID(functions []*pprof.Function) {
	sort.Slice(
		functions, func(i, j int) bool {
			return functions[i].ID < functions[j].ID
		},
	)
}

func sortLocationsByID(locations []*pprof.Location) {
	sort.Slice(
		locations, func(i, j int) bool {
			return locations[i].ID < locations[j].ID
		},
	)
}

// Test the reverse conversion function
func TestConvertStefToPprof(t *testing.T) {
	testdataDir := "testdata"

	// Read all .prof files in testdata directory
	entries, err := os.ReadDir(testdataDir)
	require.NoError(t, err, "Failed to read testdata directory")

	var profFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".prof") {
			profFiles = append(profFiles, entry.Name())
		}
	}

	require.NotEmpty(t, profFiles, "No .prof files found in testdata directory")

	for _, fileName := range profFiles {
		t.Run(
			fileName, func(t *testing.T) {
				filePath := filepath.Join(testdataDir, fileName)

				// Load original pprof profile
				file, err := os.Open(filePath)
				require.NoError(t, err)
				defer file.Close()

				gz, err := gzip.NewReader(file)
				require.NoError(t, err)

				pprofData, err := io.ReadAll(gz)
				require.NoError(t, err)

				originalProf, err := pprof.ParseUncompressed(pprofData)
				require.NoError(t, err)
				require.NotNil(t, originalProf)

				// Convert to STEF
				stefBuf := bytes.NewBuffer(nil)
				err = convertPprofToStef(originalProf, stefBuf)
				require.NoError(t, err)

				// Convert back to pprof
				convertedProf, err := convertStefToPprof(bytes.NewReader(stefBuf.Bytes()))
				require.NoError(t, err)
				require.NotNil(t, convertedProf)

				// Compare the original and converted profiles
				equal, diff := isEqualPprof(originalProf, convertedProf)
				if !equal {
					t.Errorf("Round-trip conversion failed for %s: %s", fileName, diff)

					// Print detailed comparison for debugging
					t.Logf("Original profile summary:")
					t.Logf("  Samples: %d", len(originalProf.Sample))
					t.Logf("  Locations: %d", len(originalProf.Location))
					t.Logf("  Functions: %d", len(originalProf.Function))
					t.Logf("  Mappings: %d", len(originalProf.Mapping))
					t.Logf("  SampleTypes: %d", len(originalProf.SampleType))

					t.Logf("Converted profile summary:")
					t.Logf("  Samples: %d", len(convertedProf.Sample))
					t.Logf("  Locations: %d", len(convertedProf.Location))
					t.Logf("  Functions: %d", len(convertedProf.Function))
					t.Logf("  Mappings: %d", len(convertedProf.Mapping))
					t.Logf("  SampleTypes: %d", len(convertedProf.SampleType))
				}
			},
		)
	}
}

// Test basic functionality with a synthetic profile
func TestConvertStefToPprofBasic(t *testing.T) {
	// Create a simple synthetic pprof profile
	prof := &pprof.Profile{
		TimeNanos:     1234567890,
		DurationNanos: 1000000000,
		Period:        1000000,
		PeriodType: &pprof.ValueType{
			Type: "cpu",
			Unit: "nanoseconds",
		},
		SampleType: []*pprof.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
	}

	// Add a simple mapping
	mapping := &pprof.Mapping{
		ID:              1,
		Start:           0x1000,
		Limit:           0x2000,
		File:            "/bin/test",
		HasFunctions:    true,
		HasFilenames:    true,
		HasLineNumbers:  true,
		HasInlineFrames: false,
		Offset:          0x1000,
	}
	prof.Mapping = append(prof.Mapping, mapping)

	// Add a simple function
	function := &pprof.Function{
		ID:        1,
		Name:      "main.test",
		Filename:  "main.go",
		StartLine: 10,
	}
	prof.Function = append(prof.Function, function)

	// Add a simple location
	location := &pprof.Location{
		ID:      1,
		Address: 0x1500,
		Mapping: mapping,
		Line: []pprof.Line{
			{Function: function, Line: 15, Column: 25},
		},
	}
	prof.Location = append(prof.Location, location)

	// Add a simple sample
	sample := &pprof.Sample{
		Location: []*pprof.Location{location},
		Value:    []int64{1, 1000000},
		Label:    map[string][]string{"goroutine": {"1"}},
		NumLabel: map[string][]int64{"thread": {123}},
		NumUnit:  map[string][]string{"thread": {""}},
	}
	prof.Sample = append(prof.Sample, sample)

	// Convert to STEF
	stefBuf := bytes.NewBuffer(nil)
	err := convertPprofToStef(prof, stefBuf)
	require.NoError(t, err)

	// Convert back to pprof
	convertedProf, err := convertStefToPprof(bytes.NewReader(stefBuf.Bytes()))
	require.NoError(t, err)
	require.NotNil(t, convertedProf)

	// Compare profiles
	equal, diff := isEqualPprof(prof, convertedProf)
	if !equal {
		t.Errorf("Basic round-trip conversion failed: %s", diff)
	}
}
