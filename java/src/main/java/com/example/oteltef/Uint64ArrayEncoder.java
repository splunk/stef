// Code generated by stefgen. DO NOT EDIT.
package com.example.oteltef;

import net.stef.BitsWriter;
import net.stef.SizeLimiter;
import net.stef.WriteColumnSet;
import net.stef.codecs.*;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

// Encoder for Uint64Array
class Uint64ArrayEncoder {
    private final BitsWriter buf = new BitsWriter();
    private SizeLimiter limiter;
    private Uint64Encoder elemEncoder;
    private WriterState state;
    private boolean isRecursive = false;
    private int prevLen = 0;

    public void init(WriterState state, WriteColumnSet columns) throws IOException {
        this.state = state;
        this.limiter = state.getLimiter();

        elemEncoder = new Uint64Encoder();
        elemEncoder.init(this.limiter, columns.addSubColumn());
    }

    public void reset() {
        if (!isRecursive) {
            elemEncoder.reset();
        }
        prevLen = 0;
    }

    public void encode(Uint64Array arr) throws IOException {
            int newLen = arr.elemsLen;
            int oldBitLen = buf.bitCount();
            int lenDelta = newLen - prevLen;
            prevLen = newLen;

            buf.writeVarintCompact(lenDelta);

            if (newLen > 0) {
                for (int i = 0; i < newLen; i++) {
                    elemEncoder.encode(arr.elems[i]);
                }
            }

            // Account written bits in the limiter.
            int newBitLen = buf.bitCount();
            limiter.addFrameBits(newBitLen - oldBitLen);
    }

    public void collectColumns(WriteColumnSet columnSet) {
        columnSet.setBits(buf);
        if (!isRecursive) {
            elemEncoder.collectColumns(columnSet.at(0));
        }
    }
}

