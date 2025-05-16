package net.stef.schema;

public class StructField {
    public FieldType fieldType;
    public String name;
    public boolean optional;

    public StructField(FieldType fieldType, String name, boolean optional) {
        this.fieldType = fieldType;
        this.name = name;
        this.optional = optional;
    }

    public boolean isCompatibleWith(StructField oldField) {
        if (this.optional != oldField.optional) {
            return false;
        }
        return this.fieldType.isCompatibleWith(oldField.fieldType);
    }

    // Getters and setters omitted for brevity
}
