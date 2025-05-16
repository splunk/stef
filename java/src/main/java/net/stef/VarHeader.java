package net.stef;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

public class VarHeader {
    private static final int MAX_SCHEMA_WIRE_BYTES = 1024 * 1024;
    private static final int MAX_USER_DATA_VALUES = 1024;

    private byte[] schemaWireBytes;
    private Map<String, String> userData;

    public VarHeader() {
        schemaWireBytes = new byte[0];
        userData = new HashMap<>();
    }

    public VarHeader(byte[] bytes, Map<String, String> kvHashMap) {
        this.schemaWireBytes = bytes;
        this.userData = kvHashMap;
    }

    public byte[] getSchemaWireBytes() {
        return schemaWireBytes;
    }

    public void setSchemaWireBytes(byte[] schemaWireBytes) {
        this.schemaWireBytes = schemaWireBytes;
    }

    public Map<String, String> getUserData() {
        return userData;
    }

    public void setUserData(Map<String, String> userData) {
        this.userData = userData;
    }

    public void serialize(ByteArrayOutputStream dst) throws IOException {
        Serde.writeUvarint(schemaWireBytes.length, dst);
        dst.write(schemaWireBytes);

        Serde.writeUvarint(userData.size(), dst);
        for (Map.Entry<String, String> entry : userData.entrySet()) {
            Serde.writeString(entry.getKey(), dst);
            Serde.writeString(entry.getValue(), dst);
        }
    }

    public void deserialize(ByteArrayInputStream src) throws IOException {
        long schemaLength = Serde.readUvarint(src);
        if (schemaLength > MAX_SCHEMA_WIRE_BYTES) {
            throw new IOException("Schema too large: " + schemaLength + " > " + MAX_SCHEMA_WIRE_BYTES);
        }

        schemaWireBytes = new byte[(int) schemaLength];
        if (src.read(schemaWireBytes) != schemaLength) {
            throw new IOException("Failed to read schemaWireBytes");
        }

        long userDataCount = Serde.readUvarint(src);
        if (userDataCount > MAX_USER_DATA_VALUES) {
            throw new IOException("Too many user data values: " + userDataCount + " > " + MAX_USER_DATA_VALUES);
        }

        userData = new HashMap<>();
        for (int i = 0; i < userDataCount; i++) {
            String key = Serde.readString(src);
            String value = Serde.readString(src);
            userData.put(key, value);
        }
    }
}