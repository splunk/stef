// Code generated by stefgen. DO NOT EDIT.
package com.example.oteltef;

import net.stef.BitsWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;
import net.stef.codecs.*;

import java.io.IOException;

// Encoder for EventArray
class EventArrayEncoder {
    private final BitsWriter buf = new BitsWriter();
    private SizeLimiter limiter;
    private EventEncoder encoder;
    private int prevLen = 0;
    private WriterState state;
    private Event lastVal;

    public void init(WriterState state, WriteColumnSet columns) throws IOException {
        this.state = state;
        this.limiter = state.getLimiter();
        encoder = new EventEncoder();
        state.EventEncoder = encoder;
        encoder.init(state, columns.addSubColumn());
        lastVal = new Event(null, 0);
    }

    public void reset() {
        prevLen = 0;
        encoder.reset();
    }

    public void encode(EventArray arr) throws IOException {
        int newLen = arr.elemsLen;
        int oldBitLen = buf.bitCount();
        int lenDelta = newLen - prevLen;
        prevLen = newLen;
        buf.writeVarintCompact(lenDelta);
        for (int i = 0; i < newLen; i++) {
            
            // Copy into last encoded value. This will correctly set "modified" field flags.
            lastVal.copyFrom(arr.elems[i]);
            // Encode it.
            encoder.encode(lastVal);
            // Reset modified flags so that next modification attempt correctly sets
            // the modified flags and the next encoding attempt is not skipped.
            arr.elems[i].markUnmodified();
            
        }
        int newBitLen = buf.bitCount();
        limiter.addFrameBits(newBitLen - oldBitLen);
    }

    public void collectColumns(WriteColumnSet columnSet) {
        columnSet.setBits(buf);
        
        encoder.collectColumns(columnSet.at(0));
        
    }
}

