// Code generated by stefgen. DO NOT EDIT.
// ScopeDecoder implements decoding of Scope
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class ScopeDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private Scope lastVal;
    private int fieldCount;

    
    private StringDecoder nameDecoder = new StringDecoder();
    private StringDecoder versionDecoder = new StringDecoder();
    private StringDecoder schemaURLDecoder = new StringDecoder();
    private AttributesDecoder attributesDecoder = new AttributesDecoder();
    private Uint64Decoder droppedAttributesCountDecoder = new Uint64Decoder();
    
    private ScopeDecoderDict dict;
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this encoder in the state so that we can detect recursion.
        if (state.ScopeDecoder != null) {
            throw new IllegalStateException("cannot initialize ScopeDecoder: already initialized");
        }
        state.ScopeDecoder = this;

        try {
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("Scope");
                fieldCount = fieldCount;
            } else {
                fieldCount = 5;
            }
            column = columns.getColumn();
            
            lastVal = new Scope(null, 0);
            dict = state.Scope;
            
            if (this.fieldCount <= 0) {
                return; // Name and subsequent fields are skipped.
            }
            nameDecoder.init(state.ScopeName, columns.addSubColumn());
            if (this.fieldCount <= 1) {
                return; // Version and subsequent fields are skipped.
            }
            versionDecoder.init(state.ScopeVersion, columns.addSubColumn());
            if (this.fieldCount <= 2) {
                return; // SchemaURL and subsequent fields are skipped.
            }
            schemaURLDecoder.init(state.SchemaURL, columns.addSubColumn());
            if (this.fieldCount <= 3) {
                return; // Attributes and subsequent fields are skipped.
            }
            attributesDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 4) {
                return; // DroppedAttributesCount and subsequent fields are skipped.
            }
            droppedAttributesCountDecoder.init(columns.addSubColumn());
        } finally {
            state.ScopeDecoder = null;
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
            return; // Name and subsequent fields are skipped.
        }
        this.nameDecoder.continueDecoding();
        if (this.fieldCount <= 1) {
            return; // Version and subsequent fields are skipped.
        }
        this.versionDecoder.continueDecoding();
        if (this.fieldCount <= 2) {
            return; // SchemaURL and subsequent fields are skipped.
        }
        this.schemaURLDecoder.continueDecoding();
        if (this.fieldCount <= 3) {
            return; // Attributes and subsequent fields are skipped.
        }
        this.attributesDecoder.continueDecoding();
        if (this.fieldCount <= 4) {
            return; // DroppedAttributesCount and subsequent fields are skipped.
        }
        this.droppedAttributesCountDecoder.continueDecoding();
    }

    public void reset() {
        this.nameDecoder.reset();
        this.versionDecoder.reset();
        this.schemaURLDecoder.reset();
        this.attributesDecoder.reset();
        this.droppedAttributesCountDecoder.reset();
    }

    public Scope decode(Scope dstPtr) throws IOException {
        // Check if the Scope exists in the dictionary.
        int dictFlag = buf.readBit();
        if (dictFlag == 0) {
            long refNum = buf.readUvarintCompact();
            if (refNum >= dict.size()) {
                throw new IOException("Invalid refNum");
            }
            lastVal = dict.getByIndex((int)refNum);
            dstPtr = lastVal;
            return dstPtr;
        }

        // lastValPtr here is pointing to an element in the dictionary. We are not allowed
        // to modify it. Make a clone of it and decode into the clone.
        Scope val = lastVal.clone();
        lastVal = val;
        dstPtr = val;
        // Read bits that indicate which fields follow.
        val.modifiedFields.mask = buf.readBits(fieldCount);
        
        
        if ((val.modifiedFields.mask & Scope.fieldModifiedName) != 0) {
            // Field is changed and is present, decode it.
            val.name = nameDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Scope.fieldModifiedVersion) != 0) {
            // Field is changed and is present, decode it.
            val.version = versionDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Scope.fieldModifiedSchemaURL) != 0) {
            // Field is changed and is present, decode it.
            val.schemaURL = schemaURLDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Scope.fieldModifiedAttributes) != 0) {
            // Field is changed and is present, decode it.
            val.attributes = attributesDecoder.decode(val.attributes);
        }
        
        if ((val.modifiedFields.mask & Scope.fieldModifiedDroppedAttributesCount) != 0) {
            // Field is changed and is present, decode it.
            val.droppedAttributesCount = droppedAttributesCountDecoder.decode();
        }
        
        
        dict.add(val);
        
        return val;
    }
}

