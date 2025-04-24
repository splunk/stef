package net.stef.pkg.schema;

public class StructField {
    public FieldType fieldType;
    public String name;
    public boolean optional;

    public boolean isCompatibleWith(StructField oldField) {
        if (this.optional != oldField.optional) {
            return false;
        }
        return this.fieldType.isCompatibleWith(oldField.fieldType);
    }

    // Getters and setters omitted for brevity
}
