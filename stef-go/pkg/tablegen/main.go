package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
)

var prefixes = []uint64{
	0b1,
	0b01,
	0b001,
	0b0001,
	0b00001,
	0b000001,
	0b0000001,
	0b00000001,
}

var prefixBitCounts = []uint64{
	1,
	2,
	3,
	4,
	5,
	6,
	7,
	8,
}

var payloadBitCounts = []uint64{
	0,
	2,
	5,
	12,
	19,
	26,
	33,
	48,
}

func main() {
	fmt.Println("Generating lookup tables...")
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	filePath := path.Join(dir, "bitstream_lookuptables.go")
	fmt.Printf("Writting file %s...\n", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	f = f
	fmt.Fprintf(f, "package pkg\n\n")

	fmt.Fprintf(f, "var writeBitsCountByZeros = [65]uint{\n")
	fmt.Fprintf(f, "\t0,0,0,0,0,0,0,0,\n")
	fmt.Fprintf(f, "\t0,0,0,0,0,0,0,0,\n")
	for i := 2; i <= 7; i++ {
		fmt.Fprintf(f, "\t")
		for j := 0; j <= 7; j++ {
			var bitCount uint64
			zeros := 64 - uint64(i*8+j)
			var prefIdx int
			for prefIdx = 0; prefIdx < len(payloadBitCounts); prefIdx++ {
				bitCount = payloadBitCounts[prefIdx]
				if bitCount >= zeros {
					break
				}
			}
			if bitCount < zeros {
				panic("impossible to represent")
			}
			fmt.Fprintf(f, "%2d,", prefixBitCounts[prefIdx]+payloadBitCounts[prefIdx])
		}
		fmt.Fprintf(f, "\n")
	}
	fmt.Fprintf(f, "\t1,\n}\n\n")

	fmt.Fprintf(f, "var writeMaskByZeros = [65]uint64{\n")
	fmt.Fprintf(f, "\t0,0,0,0,0,0,0,0,\n")
	fmt.Fprintf(f, "\t0,0,0,0,0,0,0,0,\n")
	for i := 2; i <= 7; i++ {
		fmt.Fprintf(f, "\t")
		for j := 0; j <= 7; j++ {
			var bitCount uint64
			zeros := 64 - uint64(i*8+j)
			var prefIdx int
			for prefIdx = 0; prefIdx < len(payloadBitCounts); prefIdx++ {
				bitCount = payloadBitCounts[prefIdx]
				if bitCount >= zeros {
					break
				}
			}
			if bitCount < zeros {
				panic("impossible to represent")
			}
			fmt.Fprintf(f, "0x%X,", prefixes[prefIdx]<<payloadBitCounts[prefIdx])
		}
		fmt.Fprintf(f, "\n")
	}
	fmt.Fprintf(f, "	0x01,\n}\n\n")

	fmt.Fprintf(f, "var readShiftByZeros = [65]uint64{\n")
	fmt.Fprintf(f, "\t0, 0, 0, 0, 0, 0, 0, 0,\n\t")
	for i := 0; i < len(prefixes); i++ {
		prefIdx := i
		fmt.Fprintf(f, "%d,", 56-prefixBitCounts[prefIdx]-payloadBitCounts[prefIdx])
	}
	fmt.Fprintf(f, "1,\n}\n\n")

	fmt.Fprintf(f, "var readMaskByZeros = [65]uint64{\n")
	fmt.Fprintf(f, "\t0, 0, 0, 0, 0, 0, 0, 0,\n\t")
	for i := 0; i < len(prefixes); i++ {
		prefIdx := i
		fmt.Fprintf(f, "0x%X,", uint64(math.MaxUint64>>(64-payloadBitCounts[prefIdx])))
	}
	fmt.Fprintf(f, "\n}\n\n")

	fmt.Fprintf(f, "var readConsumeCountByZeros = [65]uint{\n")
	fmt.Fprintf(f, "\t 0, 0, 0, 0, 0, 0, 0, 0,\n\t")
	for i := 0; i < len(prefixes); i++ {
		prefIdx := i
		fmt.Fprintf(f, "%2d,", prefixBitCounts[prefIdx]+payloadBitCounts[prefIdx])
	}
	fmt.Fprintf(f, "\n}\n\n")
}
