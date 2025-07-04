// Code generated by stefgen. DO NOT EDIT.
// SpanDecoder implements decoding of Span
package com.example.oteltef;

import net.stef.BitsReader;
import net.stef.ReadColumnSet;
import net.stef.ReadableColumn;
import net.stef.codecs.*;

import java.io.IOException;

class SpanDecoder {
    private final BitsReader buf = new BitsReader();
    private ReadableColumn column;
    private Span lastVal;
    private int fieldCount;

    
    private BytesDecoder traceIDDecoder = new BytesDecoder();
    private BytesDecoder spanIDDecoder = new BytesDecoder();
    private StringDecoder traceStateDecoder = new StringDecoder();
    private BytesDecoder parentSpanIDDecoder = new BytesDecoder();
    private Uint64Decoder flagsDecoder = new Uint64Decoder();
    private StringDecoder nameDecoder = new StringDecoder();
    private Uint64Decoder kindDecoder = new Uint64Decoder();
    private Uint64Decoder startTimeUnixNanoDecoder = new Uint64Decoder();
    private Uint64Decoder endTimeUnixNanoDecoder = new Uint64Decoder();
    private AttributesDecoder attributesDecoder = new AttributesDecoder();
    private Uint64Decoder droppedAttributesCountDecoder = new Uint64Decoder();
    private EventArrayDecoder eventsDecoder = new EventArrayDecoder();
    private LinkArrayDecoder linksDecoder = new LinkArrayDecoder();
    private SpanStatusDecoder statusDecoder = new SpanStatusDecoder();
    

    // Init is called once in the lifetime of the stream.
    public void init(ReaderState state, ReadColumnSet columns) throws IOException {
        // Remember this encoder in the state so that we can detect recursion.
        if (state.SpanDecoder != null) {
            throw new IllegalStateException("cannot initialize SpanDecoder: already initialized");
        }
        state.SpanDecoder = this;

        try {
            if (state.getOverrideSchema() != null) {
                int fieldCount = state.getOverrideSchema().getFieldCount("Span");
                fieldCount = fieldCount;
            } else {
                fieldCount = 14;
            }
            column = columns.getColumn();
            
            lastVal = new Span(null, 0);
            
            if (this.fieldCount <= 0) {
                return; // TraceID and subsequent fields are skipped.
            }
            traceIDDecoder.init(null, columns.addSubColumn());
            if (this.fieldCount <= 1) {
                return; // SpanID and subsequent fields are skipped.
            }
            spanIDDecoder.init(null, columns.addSubColumn());
            if (this.fieldCount <= 2) {
                return; // TraceState and subsequent fields are skipped.
            }
            traceStateDecoder.init(null, columns.addSubColumn());
            if (this.fieldCount <= 3) {
                return; // ParentSpanID and subsequent fields are skipped.
            }
            parentSpanIDDecoder.init(null, columns.addSubColumn());
            if (this.fieldCount <= 4) {
                return; // Flags and subsequent fields are skipped.
            }
            flagsDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 5) {
                return; // Name and subsequent fields are skipped.
            }
            nameDecoder.init(state.SpanName, columns.addSubColumn());
            if (this.fieldCount <= 6) {
                return; // Kind and subsequent fields are skipped.
            }
            kindDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 7) {
                return; // StartTimeUnixNano and subsequent fields are skipped.
            }
            startTimeUnixNanoDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 8) {
                return; // EndTimeUnixNano and subsequent fields are skipped.
            }
            endTimeUnixNanoDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 9) {
                return; // Attributes and subsequent fields are skipped.
            }
            attributesDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 10) {
                return; // DroppedAttributesCount and subsequent fields are skipped.
            }
            droppedAttributesCountDecoder.init(columns.addSubColumn());
            if (this.fieldCount <= 11) {
                return; // Events and subsequent fields are skipped.
            }
            eventsDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 12) {
                return; // Links and subsequent fields are skipped.
            }
            linksDecoder.init(state, columns.addSubColumn());
            if (this.fieldCount <= 13) {
                return; // Status and subsequent fields are skipped.
            }
            statusDecoder.init(state, columns.addSubColumn());
        } finally {
            state.SpanDecoder = null;
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
            return; // TraceID and subsequent fields are skipped.
        }
        this.traceIDDecoder.continueDecoding();
        if (this.fieldCount <= 1) {
            return; // SpanID and subsequent fields are skipped.
        }
        this.spanIDDecoder.continueDecoding();
        if (this.fieldCount <= 2) {
            return; // TraceState and subsequent fields are skipped.
        }
        this.traceStateDecoder.continueDecoding();
        if (this.fieldCount <= 3) {
            return; // ParentSpanID and subsequent fields are skipped.
        }
        this.parentSpanIDDecoder.continueDecoding();
        if (this.fieldCount <= 4) {
            return; // Flags and subsequent fields are skipped.
        }
        this.flagsDecoder.continueDecoding();
        if (this.fieldCount <= 5) {
            return; // Name and subsequent fields are skipped.
        }
        this.nameDecoder.continueDecoding();
        if (this.fieldCount <= 6) {
            return; // Kind and subsequent fields are skipped.
        }
        this.kindDecoder.continueDecoding();
        if (this.fieldCount <= 7) {
            return; // StartTimeUnixNano and subsequent fields are skipped.
        }
        this.startTimeUnixNanoDecoder.continueDecoding();
        if (this.fieldCount <= 8) {
            return; // EndTimeUnixNano and subsequent fields are skipped.
        }
        this.endTimeUnixNanoDecoder.continueDecoding();
        if (this.fieldCount <= 9) {
            return; // Attributes and subsequent fields are skipped.
        }
        this.attributesDecoder.continueDecoding();
        if (this.fieldCount <= 10) {
            return; // DroppedAttributesCount and subsequent fields are skipped.
        }
        this.droppedAttributesCountDecoder.continueDecoding();
        if (this.fieldCount <= 11) {
            return; // Events and subsequent fields are skipped.
        }
        this.eventsDecoder.continueDecoding();
        if (this.fieldCount <= 12) {
            return; // Links and subsequent fields are skipped.
        }
        this.linksDecoder.continueDecoding();
        if (this.fieldCount <= 13) {
            return; // Status and subsequent fields are skipped.
        }
        this.statusDecoder.continueDecoding();
    }

    public void reset() {
        this.traceIDDecoder.reset();
        this.spanIDDecoder.reset();
        this.traceStateDecoder.reset();
        this.parentSpanIDDecoder.reset();
        this.flagsDecoder.reset();
        this.nameDecoder.reset();
        this.kindDecoder.reset();
        this.startTimeUnixNanoDecoder.reset();
        this.endTimeUnixNanoDecoder.reset();
        this.attributesDecoder.reset();
        this.droppedAttributesCountDecoder.reset();
        this.eventsDecoder.reset();
        this.linksDecoder.reset();
        this.statusDecoder.reset();
    }

    public Span decode(Span dstPtr) throws IOException {
        Span val = dstPtr;
        // Read bits that indicate which fields follow.
        val.modifiedFields.mask = buf.readBits(fieldCount);
        
        
        if ((val.modifiedFields.mask & Span.fieldModifiedTraceID) != 0) {
            // Field is changed and is present, decode it.
            val.traceID = traceIDDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedSpanID) != 0) {
            // Field is changed and is present, decode it.
            val.spanID = spanIDDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedTraceState) != 0) {
            // Field is changed and is present, decode it.
            val.traceState = traceStateDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedParentSpanID) != 0) {
            // Field is changed and is present, decode it.
            val.parentSpanID = parentSpanIDDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedFlags) != 0) {
            // Field is changed and is present, decode it.
            val.flags = flagsDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedName) != 0) {
            // Field is changed and is present, decode it.
            val.name = nameDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedKind) != 0) {
            // Field is changed and is present, decode it.
            val.kind = kindDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedStartTimeUnixNano) != 0) {
            // Field is changed and is present, decode it.
            val.startTimeUnixNano = startTimeUnixNanoDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedEndTimeUnixNano) != 0) {
            // Field is changed and is present, decode it.
            val.endTimeUnixNano = endTimeUnixNanoDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedAttributes) != 0) {
            // Field is changed and is present, decode it.
            val.attributes = attributesDecoder.decode(val.attributes);
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedDroppedAttributesCount) != 0) {
            // Field is changed and is present, decode it.
            val.droppedAttributesCount = droppedAttributesCountDecoder.decode();
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedEvents) != 0) {
            // Field is changed and is present, decode it.
            val.events = eventsDecoder.decode(val.events);
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedLinks) != 0) {
            // Field is changed and is present, decode it.
            val.links = linksDecoder.decode(val.links);
        }
        
        if ((val.modifiedFields.mask & Span.fieldModifiedStatus) != 0) {
            // Field is changed and is present, decode it.
            val.status = statusDecoder.decode(val.status);
        }
        
        
        return val;
    }
}

