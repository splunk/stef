// Code generated by stefgen. DO NOT EDIT.
// HistogramValueDecoder implements decoding of HistogramValue
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class HistogramValueDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private HistogramValue lastVal;
    private int fieldCount;

    
    private Int64Decoder countDecoder = new Int64Decoder();
    private Float64Decoder sumDecoder = new Float64Decoder();
    private Float64Decoder minDecoder = new Float64Decoder();
    private Float64Decoder maxDecoder = new Float64Decoder();
    private Int64ArrayDecoder bucketCountsDecoder = new Int64ArrayDecoder();
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this encoder in the state so that we can detect recursion.
        if (state.HistogramValueDecoder != null) {
            throw new IllegalStateException("cannot initialize HistogramValueDecoder: already initialized");
        }
        state.HistogramValueDecoder = this;

        try {
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("HistogramValue");
                fieldCount = fieldCount;
            } else {
                fieldCount = 5;
            }
            column = columns.getColumn();
            
            lastVal = new HistogramValue(null, 0);
            
            if (this.fieldCount <= 0) {
                return; // Count and subsequent fields are skipped.
            }
            countDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 1) {
                return; // Sum and subsequent fields are skipped.
            }
            sumDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 2) {
                return; // Min and subsequent fields are skipped.
            }
            minDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 3) {
                return; // Max and subsequent fields are skipped.
            }
            maxDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 4) {
                return; // BucketCounts and subsequent fields are skipped.
            }
            bucketCountsDecoder.init(state, columns.addSubColumn());
        } finally {
            state.HistogramValueDecoder = null;
        }
    }

    // continueDecoding is called at the start of the frame to continue decoding column data.
    // This should set the decoder's source buffer, so the new decoding continues from
    // the supplied column data. This should NOT reset the internal state of the decoder,
    // since columns can cross frame boundaries and the new column data is considered
    // continuation of that same column in the previous frame.
    public void continueDecoding() {
        this.buf.reset(this.column.getData());
        
        if (this.fieldCount <= 0) {
            return; // Count and subsequent fields are skipped.
        }
        this.countDecoder.continueDecoding();
        if (this.fieldCount <= 1) {
            return; // Sum and subsequent fields are skipped.
        }
        this.sumDecoder.continueDecoding();
        if (this.fieldCount <= 2) {
            return; // Min and subsequent fields are skipped.
        }
        this.minDecoder.continueDecoding();
        if (this.fieldCount <= 3) {
            return; // Max and subsequent fields are skipped.
        }
        this.maxDecoder.continueDecoding();
        if (this.fieldCount <= 4) {
            return; // BucketCounts and subsequent fields are skipped.
        }
        this.bucketCountsDecoder.continueDecoding();
    }

    public void reset() {
        this.countDecoder.reset();
        this.sumDecoder.reset();
        this.minDecoder.reset();
        this.maxDecoder.reset();
        this.bucketCountsDecoder.reset();
    }

    public HistogramValue decode(HistogramValue dstPtr) throws IOException {
        HistogramValue val = dstPtr;
        // Read bits that indicate which fields follow.
        val.modifiedFields.mask = buf.readBits(fieldCount);
        
        // Write bits to indicate which optional fields are set.
        val.optionalFieldsPresent = buf.readBits(3);
        
        if ((val.modifiedFields.mask & HistogramValue.fieldModifiedCount) != 0) {
            // Field is changed and is present, decode it.
            val.count = countDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & HistogramValue.fieldModifiedSum) != 0 && (val.optionalFieldsPresent & HistogramValue.fieldPresentSum) != 0) {
            // Field is changed and is present, decode it.
            val.sum = sumDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & HistogramValue.fieldModifiedMin) != 0 && (val.optionalFieldsPresent & HistogramValue.fieldPresentMin) != 0) {
            // Field is changed and is present, decode it.
            val.min = minDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & HistogramValue.fieldModifiedMax) != 0 && (val.optionalFieldsPresent & HistogramValue.fieldPresentMax) != 0) {
            // Field is changed and is present, decode it.
            val.max = maxDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & HistogramValue.fieldModifiedBucketCounts) != 0) {
            // Field is changed and is present, decode it.
            val.bucketCounts = bucketCountsDecoder.decode(val.bucketCounts);
        }
        
        
        return val;
    }
}

