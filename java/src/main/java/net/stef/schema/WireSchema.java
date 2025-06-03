package net.stef.schema;

import net.stef.Serde;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

public class WireSchema {
    private static final int MAX_STRUCT_OR_MULTIMAP_COUNT = 1024;

    private Map<String, Integer> structFieldCount = new HashMap<>();

    public int getFieldCount(String structName) {
        return structFieldCount.getOrDefault(structName, 0);
    }

    public void setFieldCount(String structName, int count) {
        structFieldCount.put(structName, count);
    }

    public void serialize(OutputStream dst) throws IOException {
        DataOutputStream dataOut = new DataOutputStream(dst);

        // Write the number of structs
        dataOut.writeInt(structFieldCount.size());

        // Sort for deterministic serialization
        List<String> structNames = new ArrayList<>(structFieldCount.keySet());
        Collections.sort(structNames);

        for (String structName : structNames) {
            int fieldCount = structFieldCount.get(structName);

            // Write struct name
            writeString(dataOut, structName);

            // Write field count
            dataOut.writeInt(fieldCount);
        }
    }

    public void deserialize(InputStream src) throws IOException {
        // Read the number of structs
        long count = Serde.readUvarint(src);
        if (count > MAX_STRUCT_OR_MULTIMAP_COUNT) {
            throw new IOException("Struct count limit exceeded");
        }

        structFieldCount = new HashMap<>();
        for (int i = 0; i < count; i++) {
            // Read struct name
            String structName = readString(src);

            // Read field count
            long fieldCount = Serde.readUvarint(src);

            structFieldCount.put(structName, (int) fieldCount);
        }
    }

    public Compatibility compatible(WireSchema oldSchema) throws IOException {
        boolean exactCompat = true;

        for (Map.Entry<String, Integer> entry : oldSchema.structFieldCount.entrySet()) {
            String structName = entry.getKey();
            int oldFieldCount = entry.getValue();

            Integer newFieldCount = structFieldCount.get(structName);
            if (newFieldCount == null) {
                throw new IOException("struct " + structName + " does not exist in new schema");
            }
            if (newFieldCount < oldFieldCount) {
                throw new IOException("struct " + structName + " has fewer fields in new schema (" + newFieldCount + " vs " + oldFieldCount + ")");
            } else if (newFieldCount > oldFieldCount) {
                exactCompat = false;
            }
        }

        return exactCompat ? Compatibility.Exact : Compatibility.Superset;
    }

    private void writeString(DataOutputStream out, String value) throws IOException {
        byte[] bytes = value.getBytes(StandardCharsets.UTF_8);
        out.writeInt(bytes.length);
        out.write(bytes);
    }

    private String readString(InputStream in) throws IOException {
        long length = Serde.readUvarint(in);
        byte[] bytes = new byte[(int) length];
        in.read(bytes);
        return new String(bytes, StandardCharsets.UTF_8);
    }
}
