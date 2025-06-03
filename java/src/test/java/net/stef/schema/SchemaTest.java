package net.stef.schema;

import com.google.gson.Gson;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class SchemaTest {

    private static final Gson gson = new Gson();

    @Test
    void testSchemaSelfCompatible() {
        PrimitiveFieldType p = PrimitiveFieldType.STRING;

        Schema[] schemas = new Schema[]{
            new Schema("pkg", Map.of("Root", new Struct("Root", false, null, true))),
            new Schema("pkg", Map.of(
                "Root", new Struct("Root", false, null, true, List.of(
                    new StructField(new FieldType(null, null, null, "Multi", null), "F1", false)
                ))
            ), Map.of(
                "Multi", new Multimap("Multi", new MultimapField(new FieldType(p)), new MultimapField(new FieldType(p)))
            ))
        };

        for (Schema schema : schemas) {
            WireSchema wireSchema = schema.toWire();
            assertDoesNotThrow(() -> {
                assertEquals(Compatibility.Exact, wireSchema.compatible(wireSchema));
            });
        }
    }

    @Test
    void testSchemaSuperset() {
        PrimitiveFieldType primitiveTypeInt64 = PrimitiveFieldType.INT64;

        Schema oldSchema = new Schema("abc", Map.of(
            "Root", new Struct("Root", false, null, true, List.of(
                new StructField(new FieldType(primitiveTypeInt64), "F1", false)
            ))
        ));

        Schema newSchema = new Schema("def", Map.of(
            "Root", new Struct("Root", false, null, true, List.of(
                new StructField(new FieldType(primitiveTypeInt64), "F1", false),
                new StructField(new FieldType(primitiveTypeInt64), "F2", false)
            ))
        ));

        WireSchema oldWireSchema = oldSchema.toWire();
        WireSchema newWireSchema = newSchema.toWire();

        assertDoesNotThrow(() -> {
            assertEquals(Compatibility.Superset, newWireSchema.compatible(oldWireSchema));
        });
    }

    @Test
    void testSchemaIncompatible() {
        PrimitiveFieldType primitiveTypeInt64 = PrimitiveFieldType.INT64;

        Schema oldSchema = new Schema("abc", Map.of(
            "Root", new Struct("Root", false, null, true, List.of(
                new StructField(new FieldType(primitiveTypeInt64), "F1", false),
                new StructField(new FieldType(primitiveTypeInt64), "F2", false)
            ))
        ));

        Schema newSchema = new Schema("def", Map.of(
            "Root", new Struct("Root", false, null, true, List.of(
                new StructField(new FieldType(primitiveTypeInt64), "F1", false)
            ))
        ));

        WireSchema oldWireSchema = oldSchema.toWire();
        WireSchema newWireSchema = newSchema.toWire();

        Exception exception = assertThrows(Exception.class, () -> {
            newWireSchema.compatible(oldWireSchema);
        });

        assertEquals("struct Root has fewer fields in new schema (1 vs 2)", exception.getMessage());
    }
}