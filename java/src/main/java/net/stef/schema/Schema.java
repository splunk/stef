package net.stef.schema;

import com.google.gson.annotations.SerializedName;

import java.util.HashMap;
import java.util.Map;

public class Schema {
    @SerializedName("package") private String packageName;
    private Map<String, Struct> structs = new HashMap<>();
    private Map<String, Multimap> multimaps = new HashMap<>();
    private Map<String, EnumType> enums = new HashMap<>();

    public <K, V> Schema(String pkg, Map<String, Struct> structs) {
        this.packageName = pkg;
        this.structs = structs;
    }

    public Schema(String pkg, Map<String, Struct> structs, Map<String, Multimap> multimaps) {
        this.packageName = pkg;
        this.structs = structs;
        this.multimaps = multimaps;
    }

    public Schema() {}

    public Compatibility compatible(Schema oldSchema) throws Exception {
        boolean exact = this.structs.size() == oldSchema.structs.size();

        for (Map.Entry<String, Struct> entry : oldSchema.structs.entrySet()) {
            String name = entry.getKey();
            Struct oldStruct = entry.getValue();
            Struct newStruct = this.structs.get(name);

            if (newStruct == null) {
                throw new Exception("Struct " + name + " does not exist in new schema");
            }

            Compatibility comp = newStruct.compatibleWith(oldStruct);
            if (comp == Compatibility.Incompatible) {
                return Compatibility.Incompatible;
            }
            if (comp == Compatibility.Superset) {
                exact = false;
            }
        }

        for (Map.Entry<String, Multimap> entry : oldSchema.multimaps.entrySet()) {
            String name = entry.getKey();
            Multimap oldMap = entry.getValue();
            Multimap newMap = this.multimaps.get(name);

            if (newMap == null) {
                throw new Exception("Multimap " + name + " does not exist in new schema");
            }

            Compatibility comp = newMap.compatibleWith(oldMap);
            if (comp == Compatibility.Incompatible) {
                return Compatibility.Incompatible;
            }
            if (comp == Compatibility.Superset) {
                exact = false;
            }
        }

        return exact ? Compatibility.Exact : Compatibility.Superset;
    }

    public Schema prunedForRoot(String rootStructName) throws Exception {
        Schema prunedSchema = new Schema();
        copyPrunedStruct(rootStructName, prunedSchema);
        return prunedSchema;
    }

    private void copyPrunedStruct(String structName, Schema dst) throws Exception {
        if (dst.structs.containsKey(structName)) {
            return; // already copied
        }

        Struct srcStruct = this.structs.get(structName);
        if (srcStruct == null) {
            throw new Exception("No struct named " + structName + " found");
        }

        Struct dstStruct = new Struct(srcStruct.name, srcStruct.oneOf, srcStruct.dictName, srcStruct.isRoot);
        dst.structs.put(structName, dstStruct);

        for (StructField field : srcStruct.fields) {
            dstStruct.fields.add(field);
            copyPrunedFieldType(field.fieldType, dst);
        }
    }

    private void copyPrunedFieldType(FieldType fieldType, Schema dst) throws Exception {
        if (fieldType.struct != null) {
            copyPrunedStruct(fieldType.struct, dst);
        } else if (fieldType.multiMap != null) {
            copyPrunedMultiMap(fieldType.multiMap, dst);
        } else if (fieldType.array != null) {
            copyPrunedFieldType(fieldType.array, dst);
        }
    }

    private void copyPrunedMultiMap(String multiMapName, Schema dst) throws Exception {
        if (dst.multimaps.containsKey(multiMapName)) {
            return; // already copied
        }

        Multimap srcMultiMap = this.multimaps.get(multiMapName);
        if (srcMultiMap == null) {
            throw new Exception("No multimap named " + multiMapName + " found");
        }

        Multimap dstMultimap = new Multimap(srcMultiMap.name, srcMultiMap.key, srcMultiMap.value);
        dst.multimaps.put(multiMapName, dstMultimap);

        copyPrunedFieldType(srcMultiMap.key.type, dst);
        copyPrunedFieldType(srcMultiMap.value.type, dst);
    }
}
