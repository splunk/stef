package net.stef.schema;

public class Multimap {
    public String name;
    public MultimapField key;
    public MultimapField value;

    public Multimap(String name, MultimapField key, MultimapField value) {
        this.name = name;
        this.key = key;
        this.value = value;
    }

    public Compatibility compatibleWith(Multimap oldMap) throws Exception {
        if (!this.key.type.isCompatibleWith(oldMap.key.type)) {
            throw new Exception("Multimap " + this.name + " key type does not match");
        }
        if (!this.value.type.isCompatibleWith(oldMap.value.type)) {
            throw new Exception("Multimap " + this.name + " value type does not match");
        }
        return Compatibility.Exact;
    }

    // Getters and setters omitted for brevity
}
