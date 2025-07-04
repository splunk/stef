// Code generated by stefgen. DO NOT EDIT.
// Link Java class generated from template
package com.example.oteltef;

import net.stef.StringValue;
import net.stef.Types;
import net.stef.schema.WireSchema;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.util.*;

public class Link {
    // Field values.
    
    byte[] traceID;
    byte[] spanID;
    StringValue traceState;
    long flags;
    Attributes attributes;
    long droppedAttributesCount;

    // modifiedFields keeps track of which fields are modified.
    final ModifiedFields modifiedFields = new ModifiedFields();

    public static final String StructName = "Link";

    // Bitmasks for "modified" flags for each field.
    
    public static final long fieldModifiedTraceID = 1 << 0;
    public static final long fieldModifiedSpanID = 1 << 1;
    public static final long fieldModifiedTraceState = 1 << 2;
    public static final long fieldModifiedFlags = 1 << 3;
    public static final long fieldModifiedAttributes = 1 << 4;
    public static final long fieldModifiedDroppedAttributesCount = 1 << 5;

    

    public Link() {
        init(null, 0);
    }

    Link(ModifiedFields parentModifiedFields, long parentModifiedBit) {
        init(parentModifiedFields, parentModifiedBit);
    }

    private void init(ModifiedFields parentModifiedFields, long parentModifiedBit) {
        modifiedFields.parent = parentModifiedFields;
        modifiedFields.parentBit = parentModifiedBit;
        
        traceID = Types.emptyBytes;
        spanID = Types.emptyBytes;
        traceState = StringValue.empty;
        
        attributes = new Attributes(modifiedFields, fieldModifiedAttributes);
        
    }

    
    public byte[] getTraceID() {
        return traceID;
    }

    // setTraceID sets the value of TraceID field.
    public void setTraceID(byte[] v) {
        if (!Types.BytesEqual(this.traceID, v)) {
            this.traceID = v;
            this.markTraceIDModified();
        }
    }

    private void markTraceIDModified() {
        this.modifiedFields.markModified(fieldModifiedTraceID);
    }

    // isTraceIDModified returns true if the value of TraceID field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isTraceIDModified() {
        return (this.modifiedFields.mask & fieldModifiedTraceID) != 0;
    }
    
    public byte[] getSpanID() {
        return spanID;
    }

    // setSpanID sets the value of SpanID field.
    public void setSpanID(byte[] v) {
        if (!Types.BytesEqual(this.spanID, v)) {
            this.spanID = v;
            this.markSpanIDModified();
        }
    }

    private void markSpanIDModified() {
        this.modifiedFields.markModified(fieldModifiedSpanID);
    }

    // isSpanIDModified returns true if the value of SpanID field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isSpanIDModified() {
        return (this.modifiedFields.mask & fieldModifiedSpanID) != 0;
    }
    
    public StringValue getTraceState() {
        return traceState;
    }

    // setTraceState sets the value of TraceState field.
    public void setTraceState(StringValue v) {
        if (!Types.StringEqual(this.traceState, v)) {
            this.traceState = v;
            this.markTraceStateModified();
        }
    }

    private void markTraceStateModified() {
        this.modifiedFields.markModified(fieldModifiedTraceState);
    }

    // isTraceStateModified returns true if the value of TraceState field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isTraceStateModified() {
        return (this.modifiedFields.mask & fieldModifiedTraceState) != 0;
    }
    
    public long getFlags() {
        return flags;
    }

    // setFlags sets the value of Flags field.
    public void setFlags(long v) {
        if (!Types.Uint64Equal(this.flags, v)) {
            this.flags = v;
            this.markFlagsModified();
        }
    }

    private void markFlagsModified() {
        this.modifiedFields.markModified(fieldModifiedFlags);
    }

    // isFlagsModified returns true if the value of Flags field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isFlagsModified() {
        return (this.modifiedFields.mask & fieldModifiedFlags) != 0;
    }
    
    public Attributes getAttributes() {
        return this.attributes;
    }

    // isAttributesModified returns true if the value of Attributes field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isAttributesModified() {
        return (this.modifiedFields.mask & fieldModifiedAttributes) != 0;
    }
    
    public long getDroppedAttributesCount() {
        return droppedAttributesCount;
    }

    // setDroppedAttributesCount sets the value of DroppedAttributesCount field.
    public void setDroppedAttributesCount(long v) {
        if (!Types.Uint64Equal(this.droppedAttributesCount, v)) {
            this.droppedAttributesCount = v;
            this.markDroppedAttributesCountModified();
        }
    }

    private void markDroppedAttributesCountModified() {
        this.modifiedFields.markModified(fieldModifiedDroppedAttributesCount);
    }

    // isDroppedAttributesCountModified returns true if the value of DroppedAttributesCount field was modified since
    // Link was created, encoded or decoded. If the field is modified
    // it will be encoded by the next Write() operation. If the field is decoded by the
    // next Read() operation the modified flag will be set.
    public boolean isDroppedAttributesCountModified() {
        return (this.modifiedFields.mask & fieldModifiedDroppedAttributesCount) != 0;
    }
    

    void markUnmodified() {
        modifiedFields.markUnmodified();
        if (this.isAttributesModified()) {
            this.attributes.markUnmodified();
        }
    }

    void markModifiedRecursively() {
        attributes.markModifiedRecursively();
        modifiedFields.mask =
            fieldModifiedTraceID | 
            fieldModifiedSpanID | 
            fieldModifiedTraceState | 
            fieldModifiedFlags | 
            fieldModifiedAttributes | 
            fieldModifiedDroppedAttributesCount | 0;
    }

    void markUnmodifiedRecursively() {
        if (isAttributesModified()) {
            attributes.markUnmodifiedRecursively();
        }
        modifiedFields.mask = 0;
    }

    // markDiffModified marks fields in this struct modified if they differ from
    // the corresponding fields in v.
    boolean markDiffModified(Link v) {
        boolean modified = false;
        if (!Types.BytesEqual(traceID, v.traceID)) {
            markTraceIDModified();
            modified = true;
        }
        
        if (!Types.BytesEqual(spanID, v.spanID)) {
            markSpanIDModified();
            modified = true;
        }
        
        if (!Types.StringEqual(traceState, v.traceState)) {
            markTraceStateModified();
            modified = true;
        }
        
        if (!Types.Uint64Equal(flags, v.flags)) {
            markFlagsModified();
            modified = true;
        }
        
        if (attributes.markDiffModified(v.attributes)) {
            modifiedFields.markModified(fieldModifiedAttributes);
            modified = true;
        }
        
        if (!Types.Uint64Equal(droppedAttributesCount, v.droppedAttributesCount)) {
            markDroppedAttributesCountModified();
            modified = true;
        }
        
        return modified;
    }

    public Link clone() {
        Link cpy = new Link();
        cpy.traceID = this.traceID;
        cpy.spanID = this.spanID;
        cpy.traceState = this.traceState;
        cpy.flags = this.flags;
        cpy.attributes = this.attributes.clone();
        cpy.droppedAttributesCount = this.droppedAttributesCount;
        return cpy;
    }

    // ByteSize returns approximate memory usage in bytes. Used to calculate memory used by dictionaries.
    int byteSize() {
        int size = 0; // TODO: calculate the size of this object.
        
        
        
        
        size += this.attributes.byteSize();
        
        return size;
    }

    // Performs a deep copy from src to dst.
    public void copyFrom(Link src) {
        setTraceID(src.getTraceID());
        setSpanID(src.getSpanID());
        setTraceState(src.getTraceState());
        setFlags(src.getFlags());
        attributes.copyFrom(src.attributes);
        setDroppedAttributesCount(src.getDroppedAttributesCount());
    }

    // equals performs deep comparison and returns true if struct is equal to val.
    public boolean equals(Link val) {
        if (!Types.BytesEqual(this.traceID, val.traceID)) {
            return false;
        }
        if (!Types.BytesEqual(this.spanID, val.spanID)) {
            return false;
        }
        if (!Types.StringEqual(this.traceState, val.traceState)) {
            return false;
        }
        if (!Types.Uint64Equal(this.flags, val.flags)) {
            return false;
        }
        if (!this.attributes.equals(val.attributes)) {
            return false;
        }
        if (!Types.Uint64Equal(this.droppedAttributesCount, val.droppedAttributesCount)) {
            return false;
        }
        return true;
    }

    public static boolean equals(Link left, Link right) {
        return left.equals(right);
    }

    // compare performs deep comparison and returns an integer that
    // will be 0 if left == right, negative if left < right, positive if left > right.
    public static int compare(Link left, Link right) {
        if (left == null) {
            if (right == null) {
                return 0;
            }
            return -1;
        }
        if (right == null) {
            return 1;
        }
        int c;
        
        c = Types.BytesCompare(left.traceID, right.traceID);
        if (c != 0) {
            return c;
        }
        
        c = Types.BytesCompare(left.spanID, right.spanID);
        if (c != 0) {
            return c;
        }
        
        c = Types.StringCompare(left.traceState, right.traceState);
        if (c != 0) {
            return c;
        }
        
        c = Types.Uint64Compare(left.flags, right.flags);
        if (c != 0) {
            return c;
        }
        
        c = Attributes.compare(left.attributes, right.attributes);
        if (c != 0) {
            return c;
        }
        
        c = Types.Uint64Compare(left.droppedAttributesCount, right.droppedAttributesCount);
        if (c != 0) {
            return c;
        }
        
        return 0;
    }

    // mutateRandom mutates fields in a random, deterministic manner using random as a deterministic generator.
    void mutateRandom(Random random) {
        final int fieldCount = 6;
        
        if (random.nextInt(fieldCount) == 0) {
            this.setTraceID(Types.BytesRandom(random));
        }
        
        if (random.nextInt(fieldCount) == 0) {
            this.setSpanID(Types.BytesRandom(random));
        }
        
        if (random.nextInt(fieldCount) == 0) {
            this.setTraceState(Types.StringRandom(random));
        }
        
        if (random.nextInt(fieldCount) == 0) {
            this.setFlags(Types.Uint64Random(random));
        }
        
        if (random.nextInt(fieldCount) == 0) {
            this.attributes.mutateRandom(random);
        }
        
        if (random.nextInt(fieldCount) == 0) {
            this.setDroppedAttributesCount(Types.Uint64Random(random));
        }
        
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        return equals((Link)o);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            traceID,
            spanID,
            traceState,
            flags,
            attributes,
            droppedAttributesCount
        );
    }

    
}
