package net.stef.schema;

import java.io.IOException;

// WireSchemaIter is an iterator over the structs in a WireSchema.
public class WireSchemaIter {
    private final WireSchema schema;
    private int structIdx;

    public WireSchemaIter(WireSchema schema) {
        this.schema = schema;
        this.structIdx = 0;
    }

    // nextFieldCount returns the field count for the next struct in the schema.
    public int nextFieldCount() throws IOException {
        if (structIdx >= schema.structCounts.length) {
            throw new IOException("struct count limit exceeded");
        }

        int count = schema.structCounts[structIdx];
        structIdx++;
        return count;
    }

    public boolean done() {
        return structIdx >= schema.structCounts.length;
    }
}
