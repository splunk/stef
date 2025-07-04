// Code generated by stefgen. DO NOT EDIT.
// AnyValueDecoder implements decoding of AnyValue
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class AnyValueDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private AnyValue lastValPtr;
    private AnyValue lastVal = new AnyValue();
    private int fieldCount;
    private AnyValue.Type prevType;

    
    private StringDecoder stringDecoder = new StringDecoder();
    private BoolDecoder boolDecoder = new BoolDecoder();
    private Int64Decoder int64Decoder = new Int64Decoder();
    private Float64Decoder float64Decoder = new Float64Decoder();
    private AnyValueArrayDecoder arrayDecoder = new AnyValueArrayDecoder();
    private KeyValueListDecoder kVListDecoder = new KeyValueListDecoder();
    private BytesDecoder bytesDecoder = new BytesDecoder();
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this decoder in the state so that we can detect recursion.
        if (state.AnyValueDecoder != null) {
            throw new IllegalStateException("cannot initialize AnyValueDecoder: already initialized");
        }
        state.AnyValueDecoder = this;

        try {
            prevType = AnyValue.Type.TypeNone;
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("AnyValue");
                this.fieldCount = fieldCount;
            } else {
                this.fieldCount = 7;
            }
            this.column = columns.getColumn();
            this.lastVal.init(null, 0);
            this.lastValPtr = this.lastVal;
            Exception err = null;
            
            if (this.fieldCount <= 0) {
                return; // String and subsequent fields are skipped.
            }
            this.stringDecoder.init(state.AnyValueString, columns.addSubColumn());
            if (this.fieldCount <= 1) {
                return; // Bool and subsequent fields are skipped.
            }
            this.boolDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 2) {
                return; // Int64 and subsequent fields are skipped.
            }
            this.int64Decoder.init(columns.addSubColumn());
            if (this.fieldCount <= 3) {
                return; // Float64 and subsequent fields are skipped.
            }
            this.float64Decoder.init(columns.addSubColumn());
            if (this.fieldCount <= 4) {
                return; // Array and subsequent fields are skipped.
            }
            this.arrayDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 5) {
                return; // KVList and subsequent fields are skipped.
            }
            this.kVListDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 6) {
                return; // Bytes and subsequent fields are skipped.
            }
            this.bytesDecoder.init(null, columns.addSubColumn());
        } finally {
            state.AnyValueDecoder = null;
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
            return; // String and subsequent fields are skipped.
        }
        this.stringDecoder.continueDecoding();
        if (this.fieldCount <= 1) {
            return; // Bool and subsequent fields are skipped.
        }
        this.boolDecoder.continueDecoding();
        if (this.fieldCount <= 2) {
            return; // Int64 and subsequent fields are skipped.
        }
        this.int64Decoder.continueDecoding();
        if (this.fieldCount <= 3) {
            return; // Float64 and subsequent fields are skipped.
        }
        this.float64Decoder.continueDecoding();
        if (this.fieldCount <= 4) {
            return; // Array and subsequent fields are skipped.
        }
        this.arrayDecoder.continueDecoding();
        if (this.fieldCount <= 5) {
            return; // KVList and subsequent fields are skipped.
        }
        this.kVListDecoder.continueDecoding();
        if (this.fieldCount <= 6) {
            return; // Bytes and subsequent fields are skipped.
        }
        this.bytesDecoder.continueDecoding();
    }

    public void reset() {
        prevType = AnyValue.Type.TypeNone;
        stringDecoder.reset();
        boolDecoder.reset();
        int64Decoder.reset();
        float64Decoder.reset();
        arrayDecoder.reset();
        kVListDecoder.reset();
        bytesDecoder.reset();
    }

    // Decode decodes a value from the buffer into dst.
    public AnyValue decode(AnyValue dst) throws IOException {
        // Read type delta
        long typeDelta = this.buf.readVarintCompact();
        long typ = prevType.getValue() + typeDelta;
        if (typ < 0 || typ >= AnyValue.Type.values().length) {
            throw new IOException("Invalid oneof type");
        }
        dst.typ = AnyValue.Type.values()[(int)typ];
        prevType = dst.typ;
        this.lastValPtr = dst;
        // Decode selected field
        switch (dst.typ) {
        case TypeString:
            dst.string = this.stringDecoder.decode();
            break;
        case TypeBool:
            dst.bool = this.boolDecoder.decode();
            break;
        case TypeInt64:
            dst.int64 = this.int64Decoder.decode();
            break;
        case TypeFloat64:
            dst.float64 = this.float64Decoder.decode();
            break;
        case TypeArray:
            dst.array = this.arrayDecoder.decode(dst.array);
            break;
        case TypeKVList:
            dst.kVList = this.kVListDecoder.decode(dst.kVList);
            break;
        case TypeBytes:
            dst.bytes = this.bytesDecoder.decode();
            break;
        default:
            break;
        }
        return dst;
    }
}
