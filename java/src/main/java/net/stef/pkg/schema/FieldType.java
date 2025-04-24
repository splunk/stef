package net.stef.pkg.schema;

public class FieldType {
    public PrimitiveFieldType primitive;
    public FieldType array;
    public String struct;
    public String multiMap;
    public String dictName;

    public boolean isCompatibleWith(FieldType oldFieldType) {
        if ((this.primitive == null) != (oldFieldType.primitive == null)) {
            return false;
        }
        if (this.primitive != null && !this.primitive.equals(oldFieldType.primitive)) {
            return false;
        }
        if ((this.array == null) != (oldFieldType.array == null)) {
            return false;
        }
        if (this.array != null && !this.array.isCompatibleWith(oldFieldType.array)) {
            return false;
        }
        if (!this.struct.equals(oldFieldType.struct)) {
            return false;
        }
        if (!this.multiMap.equals(oldFieldType.multiMap)) {
            return false;
        }
        if (!this.dictName.equals(oldFieldType.dictName)) {
            return false;
        }
        return true;
    }

    // Getters and setters omitted for brevity
}
