// Code generated by stefgen. DO NOT EDIT.
package com.example.oteltef;

import net.stef.Types;

import java.util.*;
import java.util.Objects;

// LinkArray is a variable size array.
public class LinkArray {
    Link[] elems = new Link[0];

    // elemsLen is the number of elements contains in the elems, elemsLen<=elems.length.
    int elemsLen = 0;

    private ModifiedFields parentModifiedFields;
    private long parentModifiedBit;

    LinkArray() {
        init(null, 0);
    }

    LinkArray(ModifiedFields parentModifiedFields, long parentModifiedBit) {
        init(parentModifiedFields, parentModifiedBit);
    }

    private void init(ModifiedFields parentModifiedFields, long parentModifiedBit) {
        this.parentModifiedFields = parentModifiedFields;
        this.parentModifiedBit = parentModifiedBit;
    }

    // clone() creates a deep copy of LinkArray
    public LinkArray clone() {
        LinkArray clone = new LinkArray();
        clone.copyFrom(this);
        return clone;
    }

    // byteSize returns approximate memory usage in bytes. Used to calculate
    // memory used by dictionaries.
    public int byteSize() {
        int size = 0; // calculate size of the array in bytes.
        
        for (var elem : elems) {
            size += elem.byteSize();
        }
        
        return size;
    }

    

    // Append a new element at the end of the array.
    public void append(Link val) {
        ensureElems(elemsLen + 1);
        elems[elemsLen] = val;
        elemsLen++;
        markModified();
    }

    public void markModified() {
        if (parentModifiedFields != null) {
            parentModifiedFields.markModified(parentModifiedBit);
        }
    }

    public void markUnmodified() {
        if (parentModifiedFields != null) {
            parentModifiedFields.markUnmodified();
        }
    }

    public void markModifiedRecursively() {
        
        for (int i=0; i<elemsLen; i++) {
            elems[i].markModifiedRecursively();
        }
        
    }

    public void markUnmodifiedRecursively() {
        
        for (int i=0; i<elemsLen; i++) {
            elems[i].markUnmodifiedRecursively();
        }
        
    }

    // markDiffModified marks fields in each element of this array modified if they differ from
    // the corresponding fields in v.
    boolean markDiffModified(LinkArray v) {
        boolean modified = false;
        if (elemsLen != v.elemsLen) {
            // Array lengths are different, so they are definitely different.
            modified = true;
        }
    
        // Scan the elements and mark them as modified if they are different.
        int minLen = Math.min(elemsLen, v.elemsLen);
        int i=0;
        for (; i < minLen; i++) {
            if (this.elems[i].markDiffModified(v.elems[i])) {
                modified = true;
            }
        }
        // Mark the rest of the elements as modified.
        for (; i<elemsLen; i++) {
            this.elems[i].markModifiedRecursively();
        }
        
    
        if (modified) {
            this.markModified();
        }
    
        return modified;
    }

    public void copyFrom(LinkArray src) {
        boolean isModified = false;
        
        int minLen = Math.min(elemsLen, src.elemsLen);
        if (elemsLen != src.elemsLen) {
            ensureElems(src.elemsLen);
            isModified = true;
        }
        
        int i = 0;
        
        // Copy elements in the part of the array that already had the necessary room.
        for (; i < minLen; i++) {
            elems[i].copyFrom(src.elems[i]);
            isModified = true;
        }
        if (minLen < elemsLen) {
            isModified = true;
            // Need to allocate new elements for the part of the array that has grown.
            int addLen = elemsLen - minLen;
            for (int j=0; j < addLen; j++) {
                // Init the element.
                elems[i+j] = new Link(parentModifiedFields, parentModifiedBit);
                // Copy the element.
                elems[i+j].copyFrom(src.elems[i+j]);
            }
        }
        if (isModified) {
            markModified();
        }
    }

    // len returns the number of elements in the array.
    public int len() {
        return elemsLen;
    }

    // at returns element at index i.
    public Link at(int i) {
        return elems[i];
    }

    // ensureElems ensures that elems array has at least newLen elements allocated.
    // It will grow/reallocate the array if needed.
    // elemsLen will be set to newLen.
    // This method does not initialize new elements in the array.
    void ensureElems(int newLen) {
        if (elems.length < newLen) {
            int allocLen = Math.max(newLen, elems.length * 2);
            Link[] newElems = new Link[allocLen];
            System.arraycopy(elems, 0, newElems, 0, elems.length);
            elems = newElems;
        }
        elemsLen = newLen;
    }

    // ensureLen ensures the length of the array is equal to newLen.
    // It will grow or shrink the array if needed, and initialize newly added elements
    // if the element type requires initialization.
    public void ensureLen(int newLen) {
        int oldLen = elemsLen;
        if (newLen==oldLen) {
            return; // No change needed.
        }

        if (newLen > oldLen) {
            ensureElems(newLen);
            markModified();
            
            // Initialize newly added elements.
            for (int i = oldLen; i < newLen; i++) {
                
                elems[i] = new Link(parentModifiedFields, parentModifiedBit);
            }
            
        } else if (oldLen > newLen) {
            // Shrink it
            elemsLen = newLen;
            markModified();
        }
    }

    // equals performs deep comparison and returns true if array is equal to val.
    public boolean equals(LinkArray val) {
        if (elemsLen != val.elemsLen) {
            return false;
        }
        for (int i = 0; i < elemsLen; i++) {
            
            if (!elems[i].equals(val.elems[i])) {
                return false;
            }
            
        }
        return true;
    }

    // compare performs deep comparison and returns an integer that
    // will be 0 if left == right, negative if left < right, positive if left > right.
    public static int compare(LinkArray left, LinkArray right) {
        int c = left.elemsLen - right.elemsLen;
        if (c != 0) {
            return c;
        }
        for (int i = 0; i < left.elemsLen; i++) {
            int fc = Link.compare(left.elems[i], right.elems[i]);
            if (fc < 0) {
                return -1;
            }
            if (fc > 0) {
                return 1;
            }
        }
        return 0;
    }

    // mutateRandom mutates fields in a random, deterministic manner using
    // random parameter as a deterministic generator.
    void mutateRandom(Random random) {
        if (random.nextInt(20) == 0) {
            ensureLen(len() + 1);
        }
        if (random.nextInt(20) == 0 && len() > 0) {
            ensureLen(len() - 1);
        }
        for (int i = 0; i < elemsLen; i++) {
            if (random.nextInt(2 * elemsLen) == 0) {
                
                elems[i].mutateRandom(random);
                
            }
        }
    }
}
