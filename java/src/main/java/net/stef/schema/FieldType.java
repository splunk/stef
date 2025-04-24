package net.stef.schema;

public class FieldType {
    public PrimitiveFieldType primitive;
    public FieldType array;
    public String struct;
    public String multiMap;
    public String dictName;

    public FieldType(PrimitiveFieldType primitive) {
        this.primitive = primitive;
    }

    public FieldType(PrimitiveFieldType primitive, FieldType array, String struct, String multi, String dictName) {
        this.primitive = primitive;
        this.array = array;
        this.struct = struct;
        this.multiMap = multi;
        this.dictName = dictName;
    }

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
