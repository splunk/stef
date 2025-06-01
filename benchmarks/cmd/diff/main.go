package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pkg"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: diff <file1> <file2>")
		os.Exit(1)
	}

	fname1 := os.Args[1]
	fname2 := os.Args[2]

	if err := diffFiles(fname1, fname2); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	os.Exit(0)
}

func diffFiles(fname1 string, fname2 string) error {
	reader1, err := openOtelStef(fname1)
	if err != nil {
		return err
	}

	reader2, err := openOtelStef(fname2)
	if err != nil {
		return err
	}

	for {
		var err1 error
		if err1 = reader1.Read(pkg.ReadOptions{}); err1 != nil {
			if err1 != io.EOF {
				return err1
			}
		}

		var err2 error
		if err2 = reader2.Read(pkg.ReadOptions{}); err2 != nil {
			if err2 != io.EOF {
				return err2
			}
		}

		if errors.Is(err1, io.EOF) {
			if errors.Is(err2, io.EOF) {
				break
			}
			return fmt.Errorf("%s has more records than %s", fname2, fname1)
		}
		if errors.Is(err2, io.EOF) {
			return fmt.Errorf("%s has more records than %s", fname1, fname2)
		}

		if !reader1.Record.IsEqual(&reader2.Record) {
			return fmt.Errorf("Record #%d differs.", reader1.RecordCount())
		}
	}

	fmt.Printf("%d records compared. Content is identical.\n", reader1.RecordCount())

	return nil
}

func openOtelStef(fname1 string) (*oteltef.MetricsReader, error) {
	file, err := os.Open(fname1)
	if err != nil {
		return nil, fmt.Errorf("Cannot open %s: %v", fname1, err)
	}

	reader, err := oteltef.NewMetricsReader(file)
	if err != nil {
		return nil, fmt.Errorf("Cannot open %s: %v", fname1, err)
	}

	return reader, nil
}
