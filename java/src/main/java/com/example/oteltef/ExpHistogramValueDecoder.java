// Code generated by stefgen. DO NOT EDIT.
// ExpHistogramValueDecoder implements decoding of ExpHistogramValue
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class ExpHistogramValueDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private ExpHistogramValue lastVal;
    private int fieldCount;

    
    private Uint64Decoder countDecoder = new Uint64Decoder();
    private Float64Decoder sumDecoder = new Float64Decoder();
    private Float64Decoder minDecoder = new Float64Decoder();
    private Float64Decoder maxDecoder = new Float64Decoder();
    private Int64Decoder scaleDecoder = new Int64Decoder();
    private Uint64Decoder zeroCountDecoder = new Uint64Decoder();
    private ExpHistogramBucketsDecoder positiveBucketsDecoder = new ExpHistogramBucketsDecoder();
    private ExpHistogramBucketsDecoder negativeBucketsDecoder = new ExpHistogramBucketsDecoder();
    private Float64Decoder zeroThresholdDecoder = new Float64Decoder();
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this encoder in the state so that we can detect recursion.
        if (state.ExpHistogramValueDecoder != null) {
            throw new IllegalStateException("cannot initialize ExpHistogramValueDecoder: already initialized");
        }
        state.ExpHistogramValueDecoder = this;

        try {
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("ExpHistogramValue");
                fieldCount = fieldCount;
            } else {
                fieldCount = 9;
            }
            column = columns.getColumn();
            
            lastVal = new ExpHistogramValue(null, 0);
            
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
                return; // Scale and subsequent fields are skipped.
            }
            scaleDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 5) {
                return; // ZeroCount and subsequent fields are skipped.
            }
            zeroCountDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 6) {
                return; // PositiveBuckets and subsequent fields are skipped.
            }
            positiveBucketsDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 7) {
                return; // NegativeBuckets and subsequent fields are skipped.
            }
            negativeBucketsDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 8) {
                return; // ZeroThreshold and subsequent fields are skipped.
            }
            zeroThresholdDecoder.init(columns.addSubColumn());
        } finally {
            state.ExpHistogramValueDecoder = null;
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
            return; // Scale and subsequent fields are skipped.
        }
        this.scaleDecoder.continueDecoding();
        if (this.fieldCount <= 5) {
            return; // ZeroCount and subsequent fields are skipped.
        }
        this.zeroCountDecoder.continueDecoding();
        if (this.fieldCount <= 6) {
            return; // PositiveBuckets and subsequent fields are skipped.
        }
        this.positiveBucketsDecoder.continueDecoding();
        if (this.fieldCount <= 7) {
            return; // NegativeBuckets and subsequent fields are skipped.
        }
        this.negativeBucketsDecoder.continueDecoding();
        if (this.fieldCount <= 8) {
            return; // ZeroThreshold and subsequent fields are skipped.
        }
        this.zeroThresholdDecoder.continueDecoding();
    }

    public void reset() {
        this.countDecoder.reset();
        this.sumDecoder.reset();
        this.minDecoder.reset();
        this.maxDecoder.reset();
        this.scaleDecoder.reset();
        this.zeroCountDecoder.reset();
        this.positiveBucketsDecoder.reset();
        this.negativeBucketsDecoder.reset();
        this.zeroThresholdDecoder.reset();
    }

    public ExpHistogramValue decode(ExpHistogramValue dstPtr) throws IOException {
        ExpHistogramValue val = dstPtr;
        // Read bits that indicate which fields follow.
        val.modifiedFields.mask = buf.readBits(fieldCount);
        
        // Write bits to indicate which optional fields are set.
        val.optionalFieldsPresent = buf.readBits(3);
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedCount) != 0) {
            // Field is changed and is present, decode it.
            val.count = countDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedSum) != 0 && (val.optionalFieldsPresent & ExpHistogramValue.fieldPresentSum) != 0) {
            // Field is changed and is present, decode it.
            val.sum = sumDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedMin) != 0 && (val.optionalFieldsPresent & ExpHistogramValue.fieldPresentMin) != 0) {
            // Field is changed and is present, decode it.
            val.min = minDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedMax) != 0 && (val.optionalFieldsPresent & ExpHistogramValue.fieldPresentMax) != 0) {
            // Field is changed and is present, decode it.
            val.max = maxDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedScale) != 0) {
            // Field is changed and is present, decode it.
            val.scale = scaleDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedZeroCount) != 0) {
            // Field is changed and is present, decode it.
            val.zeroCount = zeroCountDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedPositiveBuckets) != 0) {
            // Field is changed and is present, decode it.
            val.positiveBuckets = positiveBucketsDecoder.decode(val.positiveBuckets);
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedNegativeBuckets) != 0) {
            // Field is changed and is present, decode it.
            val.negativeBuckets = negativeBucketsDecoder.decode(val.negativeBuckets);
        }
        
        if ((val.modifiedFields.mask & ExpHistogramValue.fieldModifiedZeroThreshold) != 0) {
            // Field is changed and is present, decode it.
            val.zeroThreshold = zeroThresholdDecoder.decode();
        }
        
        
        return val;
    }
}

