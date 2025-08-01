// Code generated by stefgen. DO NOT EDIT.
// EnvelopeDecoder implements decoding of Envelope
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class EnvelopeDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private Envelope lastVal;
    private int fieldCount;

    
    private EnvelopeAttributesDecoder attributesDecoder = new EnvelopeAttributesDecoder();
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this encoder in the state so that we can detect recursion.
        if (state.EnvelopeDecoder != null) {
            throw new IllegalStateException("cannot initialize EnvelopeDecoder: already initialized");
        }
        state.EnvelopeDecoder = this;

        try {
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("Envelope");
                fieldCount = fieldCount;
            } else {
                fieldCount = 1;
            }
            column = columns.getColumn();
            
            lastVal = new Envelope(null, 0);
            
            if (this.fieldCount <= 0) {
                return; // Attributes and subsequent fields are skipped.
            }
            attributesDecoder.init(state, columns.addSubColumn());
        } finally {
            state.EnvelopeDecoder = null;
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
            return; // Attributes and subsequent fields are skipped.
        }
        this.attributesDecoder.continueDecoding();
    }

    public void reset() {
        this.attributesDecoder.reset();
    }

    public Envelope decode(Envelope dstPtr) throws IOException {
        Envelope val = dstPtr;
        // Read bits that indicate which fields follow.
        val.modifiedFields.mask = buf.readBits(fieldCount);
        
        
        if ((val.modifiedFields.mask & Envelope.fieldModifiedAttributes) != 0) {
            // Field is changed and is present, decode it.
            val.attributes = attributesDecoder.decode(val.attributes);
        }
        
        
        return val;
    }
}

