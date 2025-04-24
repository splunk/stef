package net.stef.schema;

import java.util.ArrayList;
import java.util.List;

public class Struct {
    public String name;
    public boolean oneOf;
    public String dictName;
    public boolean isRoot;
    public List<StructField> fields = new ArrayList<>();

    public Struct(String name, boolean oneOf, String dictName, boolean isRoot) {
        this.name = name;
        this.oneOf = oneOf;
        this.dictName = dictName;
        this.isRoot = isRoot;
    }

    public Struct(String name, boolean oneOf, String dictName, boolean isRoot, List<StructField> fields) {
        this.name = name;
        this.oneOf = oneOf;
        this.dictName = dictName;
        this.isRoot = isRoot;
        this.fields = fields;
    }

    public Compatibility compatibleWith(Struct oldStruct) throws Exception {
        if (this.fields.size() < oldStruct.fields.size()) {
            throw new Exception("New struct " + this.name + " has fewer fields than old struct");
        }

        if (this.oneOf != oldStruct.oneOf) {
            throw new Exception("New struct " + this.name + " has different oneOf flag than the old struct");
        }

        if (!this.dictName.equals(oldStruct.dictName)) {
            throw new Exception("New struct " + this.name + " dictionary name is different from the old struct");
        }

        boolean exact = this.fields.size() == oldStruct.fields.size();

        for (int i = 0; i < oldStruct.fields.size(); i++) {
            StructField newField = this.fields.get(i);
            StructField oldField = oldStruct.fields.get(i);
            if (!newField.isCompatibleWith(oldField)) {
                throw new Exception("Field " + i + " in new struct " + this.name + " is incompatible with the old struct");
            }
        }

        return exact ? Compatibility.EXACT : Compatibility.SUPERSET;
    }

    // Getters and setters omitted for brevity
}