package net.stef;

import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.util.HashMap;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

public class VarHeaderTest {

    private static final VarHeader[] VAR_HEADER_TESTS = {
            new VarHeader(),
            new VarHeader(new byte[]{}, new HashMap<>()),
            new VarHeader("012".getBytes(), new HashMap<>()),
            new VarHeader("012345".getBytes(), Map.of("abc", "def", "0", "world"))
    };

    @Test
    public void testVarHeaderSerialization() {
        for (VarHeader original : VAR_HEADER_TESTS) {
            ByteArrayOutputStream buffer = new ByteArrayOutputStream();
            assertDoesNotThrow(() -> original.serialize(buffer));

            ByteArrayInputStream input = new ByteArrayInputStream(buffer.toByteArray());
            VarHeader copy = new VarHeader();
            assertDoesNotThrow(() -> copy.deserialize(input));

            assertArrayEquals(copy.getSchemaWireBytes(), original.getSchemaWireBytes());
            assertEquals(copy.getUserData(),original.getUserData());
        }
    }
}