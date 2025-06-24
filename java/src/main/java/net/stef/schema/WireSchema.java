package net.stef.schema;

import net.stef.Serde;

import java.io.DataOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

public class WireSchema {
    private static final int MAX_STRUCT_COUNT = 1024;

    // structCounts is a linearized list of the number of fields in each struct/oneof
    // in the schema. The order of the structs in this list is the same as the order
    // in which the structs are encoded/decoded in the schema.
    int[] structCounts;

    public void serialize(OutputStream dst) throws IOException {
        DataOutputStream dataOut = new DataOutputStream(dst);

        // Write the number of structs
        Serde.writeUvarint(structCounts.length, dataOut);

        for (int c : structCounts) {
            Serde.writeUvarint(c, dataOut);
        }
    }

    public void deserialize(InputStream src) throws IOException {
        long len = Serde.readUvarint(src);
        if (len > MAX_STRUCT_COUNT) {
            throw new IOException("struct count limit exceeded");
        }

        structCounts = new int[(int) len];
        for (int i = 0; i < len; i++) {
            structCounts[i] = (int) Serde.readUvarint(src);
        }
    }

    public Compatibility compatible(WireSchema oldSchema) {
        if (this.structCounts.length > oldSchema.structCounts.length) {
            return Compatibility.Superset;
        }

        if (this.structCounts.length < oldSchema.structCounts.length) {
            return Compatibility.Incompatible;
        }

        int newFieldTotal = 0;
        int oldFieldTotal = 0;

        for (int i = 0; i < this.structCounts.length; i++) {
            newFieldTotal += this.structCounts[i];
            oldFieldTotal += oldSchema.structCounts[i];
        }

        if (newFieldTotal > oldFieldTotal) {
            return Compatibility.Superset;
        }

        if (newFieldTotal < oldFieldTotal) {
            return Compatibility.Incompatible;
        }

        return Compatibility.Exact;
    }
}
