# Sequential Tabular Encoding Format (STEF)

Tigran Najaryan<br/>
December 2024

Status: [Development]

<details>
<summary>Table of Contents</summary>

<!-- toc -->

- [Glossary](#glossary)
- [Conventions](#conventions)
- [Record Schema](#record-schema)
  * [Recursive Type Definitions](#recursive-type-definitions)
  * [Dictionary Encoding](#dictionary-encoding)
  * [Schema Tree](#schema-tree)
- [STEF Data Format](#stef-data-format)
  * [Header](#header)
  * [Frame](#frame)
  * [VarHeader Frame](#varheader-frame)
- [Data Frame](#data-frame)
  * [Columnar Representation](#columnar-representation)
- [Codecs](#codecs)
  * [Struct Codec](#struct-codec)
    + [Full Struct Encoding](#full-struct-encoding)
    + [Dictionary Struct Encoding](#dictionary-struct-encoding)
  * [OneOf Codec](#oneof-codec)
  * [Array Codec](#array-codec)
  * [MultiMap Codec](#multimap-codec)
    + [Full MultiMap Encoding](#full-multimap-encoding)
    + [Value-Only MultiMap Encoding](#value-only-multimap-encoding)
  * [String Codec](#string-codec)
    + [Direct String Encoding](#direct-string-encoding)
    + [Reference String Encoding](#reference-string-encoding)
  * [Uint64 and Int64 Codecs](#uint64-and-int64-codecs)
  * [Float64 Codec](#float64-codec)
  * [Bool Codec](#bool-codec)
- [Value Formats](#value-formats)
  * [Uvarint64](#uvarint64)
  * [Varint64](#varint64)
  * [UvarintCompact](#uvarintcompact)
  * [VarintCompact](#varintcompact)
- [Dictionaries](#dictionaries)
  * [Resetting Dictionaries](#resetting-dictionaries)
- [STEF/gRPC Protocol](#stefgrpc-protocol)
  * [stef_bytes field](#stef_bytes-field)
  * [is_end_of_chunk field](#is_end_of_chunk-field)
  * [Record Ids](#record-ids)
  * [Streaming](#streaming)
  * [Connection Management](#connection-management)
  * [Dictionary Limits](#dictionary-limits)

<!-- tocstop -->

</details>

## Glossary

- **STEF Data Stream** - a sequence of bytes that is compliant with this specification.
- **Writer** - an implementation of STEF protocol that can create a STEF Stream.
- **Reader** - an implementation of STEF protocol that can read and decode a STEF Stream.

## Conventions

The data format described in this specification uses a custom notation. Unless 
otherwise specified all data is tightly packed. Fields may occupy a number of bits 
that is not an integer multiply of 8 and may cross the boundaries of bytes. There is no 
implied padding after such fields. When any padding between fields is used it is
explicitly called out.

Fields are MSB-ordered, i.e. most-significant bits of the byte are
used first. For example a 3-bit field A, followed by a 2-bit field B looks like this
in a byte:

```
+-----------+-------+-----------+
|A2  A1  A0 |B1  B0 |  unused   |
+-----------+-------+-----------+
  7   6   5   4   3   2   1   0
```

Bit fields spanning more than one byte continue using the same MSB ordering, i.e.
adding a 5-bit field C to the above example results in this:

```
            Byte 0                                Byte 1

+-----------+-------+-----------+    +-------+-----------------------+
|A2  A1  A0 |B1  B0 |C4  C3  C2 |    |C1  C0 |        unused         |
+-----------+-------+-----------+    +-------+-----------------------+
  7   6   5   4   3   2   1   0        7   6   5   4   3   2   1   0
```

Formats of data structures are described in the following notation:

```
<Name of the structure> {
    <Field Name 1>: <Field Size or Type>
    <Field Name 2>: <Field Size or Type>
    ... more fields
}
```

There is one or more fields in the structure. A field is named and has a size or type 
notation after the colon.

Here is the list of field size and type notations:

`Field: N` - Field is exactly N bits long.

`Field: N..M` - Field is between N and M bits long. The actual size is specified in 
the text that follows. If M is omitted then the upper bound is unspecified.

`Field: U64` - Field holds an unsigned integer value using the [Uvarint64](#uvarint64)
variable-length encoding.

`Field: S64` - Field holds an signed integer value using the [Varint64](#varint64)
variable-length encoding.

`Field: UC` - Field holds an unsigned integer value using the
[UvarintCompact](#uvarintcompact) variable-length encoding.

`Field: SC` - Field holds an signed integer value using the
[VarintCompact](#varintcompact) variable-length encoding.

`Field: <ST> = C` - Field has a constant value of C. Size/type <ST> can be any one of the 
notations above.

`/Field: <ST>/` - Field is optional.

`*Field: <ST>` - Field is repeated zero or more times and each instance has a 
size/type <ST>.

## Record Schema

STEF data stream logically represents a sequence of records. A record is a collection of 
fields. The structure of a record is called the "schema" of the particular STEF data 
stream, or simply STEF schema.

The schema of the STEF data stream is defined statically and all records in the data 
stream comply with the same schema. STEF schema describes fields in a record, field types, 
possible values for the fields. Supported field types are:

- Primitive: either a bool, int64, uint64, float64, string or bytes.
- Array: an array of elements of any type. The array may contain zero or more elements.
- Oneof: can contain only one out of specified values at a time or the sentinel None 
  value.
- Multimap: an associative array of key-value pairs. The array can have zero or more 
  pair. Keys and values can be of arbitrary type.
- Struct: a collection of fields of any type. Fields may be marked "optional" in the 
  struct, in which case their "presence" is tracked explicitly. The struct that 
  represents the structure of the STEF data stream record is called the "root" struct.

Here is an example schema for a basic record that holds a single measurement of a metric:

```
struct Measurement root {
    MetricName string
    Attributes Attributes
    Timestamp uint64
    Value PointValue
}

oneof PointValue {
    Int64 int64
    Float64 float64
}

multimap Attributes {
    key string
    value string
}
```

This schema defines a root struct `Measurement` that represents the STEF stream records.
Each record will contain MetricName string, zero or more Attributes (each a key-value 
pair), a Timestamp and a PointValue, which can be either an int64 or a float64 value.

A STEF stream of this schema may for example contain the following records:

| MetricName       | Attributes                   | Timestamp  | Value        |
|------------------|------------------------------|------------|--------------|
| "cpu.usage"      | Key="cpu" Value="1"          | 1783726193 | Float64=0.4  |
| "cpu.usage"      | Key="cpu" Value="2"          | 1783726193 | Float64=0.1  |
| "memory.usage"   | Key="memory" Value="virtual" | 1783726194 | Int64=100000 |
| "system.healthy" |                              | 1783726194 | Int64=1      |
| "system.healthy" |                              | 1783726195 | Int64=0      |

### Recursive Type Definitions

Non-primitive types in STEF schema may recursively refer to each other or to themselves. 
For example let's look at a definition of ArrayishValue:

```
oneof ArrayishValue {
    Int64 int64
    Array []ArrayishValue
}
```

An ArrayishValue like this can for example represent values:
- `15`
- `[1,2]`
- `[]`
- `[1,2,[32,-1,[14]],[]]`.

Loops in data defined via recursive data types are prohibited. STEF record is strictly
a tree of fields.

### Dictionary Encoding

Primitive values of string and bytes types and structs can be optionally 
dictionary-encoded. When dictionary encoding is enabled for a particular field, STEF 
encoders will add previously seen values of the field into a dictionary and when a 
repeat value is seen instead of encoding it again, a reference to an existing 
dictionary entry will be encoded instead. Here is an example with dictionary encoding 
of metric names:

```
struct Measurement root {
    MetricName string dict(Names)
    Attributes Attributes
    Timestamp uint64
    Value PointValue
}
```

Here the `MetricName` field is dictionary-encoded and the name of the dictionary is 
`Names`. Fields of the same type may share dictionaries, e.g:

```
struct Person root {
    First string dict(Names)
    Last string dict(Names)
}
```

In this example `Names` is a single dictionary that contains previously seen value both
from field `First` and field `Last`.

A field of struct type can be also dictionary encoded, by declaring the struct itself 
dictionary-encoded. For example the `Country` struct is dictionary-encoded below:

```
struct Address root {
    Street string
    City string
    State string
    Country Country
}

struct Country dict(Countries) {
    Name string
    ISOCode string
}
```

A dictionary-encoded struct that is used as a field in multiple places in the schema will 
use one dictionary that is shared for all structs of that type.

### Schema Tree

STEF schema definition forms a tree of fields. The root of the tree is the root struct, 
each field a children of the root and then recursively each field of non-primitive 
type containing children that represent the non-primitive type. We can represent the 
above `Measurement` example by the following tree:

```
root Measurement
 |- MetricName string
 |- Attributes Attributes
 |   |- key string
 |   |- value string
 |- Timestamp uint64
 |- Value PointValue
     |- Int64 int64
     |- Float64 float64
```

The tree is constructed by traversal of schema definition in the following 
order:

- We start at the root struct and assign the root node to the root struct.
- If the node type is struct or oneof we visit each field in the order of declaration and 
  assign each field a child node, one child per field.
- If the node type is array we add one child to the node. The child represents array's 
  element ype.
- If the node type is multimap we add two children to the node. The first child 
  represents the key, the second child represents the value.
- The process is repeated for all non-primitive node types or until a 
  declaration loop is detected due to types declarations in the schema referring to 
  each other or to themselves recursively. If a loop is detected the branch of the 
  tree is considered constructed and the algorithm backtracks.

Leaf nodes of STEF schema tree are always primitive types or are non-primitive types 
where a type recursion is detected.

Let's illustrate a more complete example that includes recursive type definitions:

```
struct Measurement root {
    MetricName string
    Attributes Attributes
    Timestamp uint64
    Value PointValue
}

oneof PointValue {
    Int64 int64
    Float64 float64
}

multimap Attributes {
    key string
    value AnyValue
}

oneof AnyValue {
    String string
    Array []AnyValue
    KVList KVList
}

multimap KVList {
    key string
    value AnyValue
}
```

In the above schema AnyValue is a self-referential recursive type and the types AnyValue 
and KVList mutually refer to each other. The corresponding schema tree looks like this:

```
root Measurement
 |- MetricName string
 |- Attributes Attributes
 |   |- key string
 |   |- value AnyValue   
 |       |- String string
 |       |- Array []   
 |           |- AnyValue          <--- loop detected here, backtrack. Non-primitive leaf.
 |       |- KVList KVList
 |           |- key string 
 |           |- value AnyValue    <--- loop detected here, backtrack. Non-primitive leaf.
 |- Timestamp uint64
 |- Value PointValue
     |- Int64 int64
     |- Float64 float64
```

Each record in STEF data stream logically is a collection of values associated with 
leaf nodes in this tree.

It is important to note 2 things.

Firstly, for each individual record not every leaf node in a tree always has an 
associated value. This is because a oneof may only contain one value at a 
time (and thus only one of its children will have a value associated) and optional 
fields may have no value at all. Furthermore, entire subtrees may have no values 
associated with them if they are contained in a non-present optional field or are in a 
oneof that stores a different choice.

Secondly, because the schema allows recursive types a record may contain more than one 
value associated with the same node in the schema tree. Consider the following `AnyValue`:

```
AnyValue = { KVList = { key = "abc", value = { AnyValue = { String = "xyx" } } } }
```

Represented as a tree this AnyValue can be laid out as:

```
AnyValue
 |- KVList
     |- key = "abc"   
     |- Value
         |- AnyValue
             |- String = "xyz"
```

Note how there are 2 AnyValue elements, although the schema will only have one 
AnyValue node.

If we associate these elements with the schema tree, we will have essentially 2 
AnyValue elements associated with the AnyValue node in the schema tree.

## STEF Data Format

STEF data is represented as a byte stream and has the following format:

```
STEF {
    FixedHeader: ..
    VarHeader Frame: ..
    *Data Frame: ..
}
```

It starts with a FixedHeader, followed by VarHeader frame, followed by zero or more 
Data Frames.

### Header

```
FixedHeader {
    Signature: 32 = 4 ASCII bytes: "S", "T", "E", "F"
    Version: 4
    Compression: 2
    Random: 2
}
```

`Version` specifies the major version number of STEF stream format:
  - 0 - Current version number defined by this specification.
  - 1-15 - Reserved. The readers compliant with this specification MUST refuse to
    decode STEF stream if these values are specified.

`Compression` is the compression method for VarHeader and Data frame content:
  - 0 - no compression.
  - 1 - Zstd compression.
  - 2 and 3 - Reserved. The readers compliant with this specification MUST refuse to
    decode STEF stream if these values are specified.

`Random` - a random value beween 0 and 3.

### Frame

Frame structure is used to represent either a VarHeader or a Data Frame.

```
Frame {
    RestartDictionaries: 1
    RestartCompression: 1
    RestartCodecs: 1
    Random: 5
    UncompressedSize: U64
    /CompressedSize: U64/
    Content: ..
}
```

RestartDictionaries indicates that all dictionaries are cleared and
restarted when this frame starts. Can be used to limit the size of the
dictionaries that the readers must keep in memory. Has no effect on VarHeader and SHOULD
be used only for Data Frames.

RestartCompression indicates that the compression stream is started
anew from this frame's Content. Reader's state of compression decoder MUST be
reset when this bit is set. If this bit is unset the state of the compression encoder
carries over through the Content field of frames. This bit has effect only if Flags
field in the Header specifies a compression.

RestartCodecs indicates that the state of encoders/decoders is cleared and started
anew from this frame's Content.

The UncompressedSize size field specifies the total size in bytes of the frame Content
in uncompressed form. If no compression is used the UncompressedSize field is
also the same as the size of the Content field in bytes.

The CompressedSize field is optional and is only present if Compression field in the
Header specifies a compression. The CompressedSize field specifies the size of
the Content field in bytes.

### VarHeader Frame

VarHeader is a Frame with a `Content` field that contains JSON-encoded Schema of STEF data. 
TODO: specify JSON fields.

## Data Frame

Data Frame has the `Content` of the following structure:

```
Content {
    RecordCount: U64
    SizeOfSizes: U64
    *Size: UC
    *ColumnData: 0..
}
```

`RecordCount` is the number of record in this data frame.

`SizeOfSizes` is the size of the following `Sizes` field in bytes.

`Size` contains a sequence of UvarintCompact values, one value per column,
each column representing the number of bytes of data in that column. 0 is valid value 
for `Size` and indicates that the column has no data. The Size field is repeated as 
many time as there are columns in the schema.

Note if a column has no data, all its subcolumns also have no data. In this case the Size
field of value 0 is recorded only for the parent column and is omitted for all child
columns.

`ColumnData` represents the data for one column. The ColumnData field is repeated as 
many time as there are columns in the schema.

Each column contains a sequence of data elements. The elements may of different sizes 
within the same column. Columns may contain a different number of elements. The next 
section describes the column data.

### Columnar Representation

[Above](#record-schema) we described the logical data model, how the data STEF data is
presented to readers and writers of records. Now we are going to look into how this
data is encoded in bits and bytes in STEF data stream.

STEF format uses columnar representation. Each node in STEF schema tree is represented
by one column, including nodes of primitive and non-primitive types. The number of
columns in STEF data stream is fixed and is defined by its schema. Each column is
assigned a number that corresponds to the depth-first traversal order of the
constructed schema tree.

Note that some fields in a schema may contain no values at all for any of the records.
In this case the column will be empty. Nevertheless the column is still numbered.

The encoded data for each of these columns is contained in one `ColumnData` field of a 
data frame. The `ColumnData` contains encoded data of the particular column for all 
records that the frame represents.

Here is an example column number assignment and codec types for the sample schema we were 
looking into earlier:

```
Tree                         Column     Codec Type
root Measurement                1        struct
 |- MetricName string           2        string
 |- Attributes Attributes       3        multimap
 |   |- key string              4        string
 |   |- value AnyValue          5        oneof    
 |       |- String string       6        string   
 |       |- Array []            7        array 
 |           |- AnyValue        5        oneof (note column 5 because recursion is detected)
 |       |- KVList KVList       8        multimap                       
 |           |- key string      9        string                         
 |           |- value AnyValue  5        oneof (note column 5 because recursion is detected)    
 |- Timestamp uint64            10       uint64
 |- Value PointValue            11       oneof
     |- Int64 int64             12       int64
     |- Float64 float64         13       float64
```


Now, let's see how the encoding of records into `ColumnData` fields work.

The input to the encoding process is a single record, laid out as a tree of values. The 
encoder also has the constructed schema tree and `ColumnData` that correspond to the 
tree nodes. The encoding algorithm is the following:

- Start at the root of the record tree (remember, this is a struct).
- Visit the record tree in depth-first order and perform encoding operation of the 
  visited node into the column of the associated schema node, using the appropriate 
  codec for the node type.

## Codecs

This section describes codecs for various data types that can be used in the schema.

Some codecs encode data in the form of delta (or difference) from a previous
instance of that same field. Note that choice of words: we say the "previous instance"
instead of saying "the value in the previous record". This distinction is important.

Although, in the simplest cases the previous instance is indeed in the previous
record, in more complex cases it may not be the case. This is for example possible if
the previous record did not contain a value for this field at all (this is possible if
the field is optional). This is also possible if the schema declares an array of
structs and encoder is encoding the elements of that array. In that case "the
previous instance" refers to the previous array element (provided that the element has
a value for that field!).

### Struct Codec

Struct codec has 2 encoding modes: full and dictionary.

#### Full Struct Encoding

```
Struct {
    /FullEncoding: 1 = 1/
    ModifiedMask: N
    /PresenceMask: M/
}
```

If the schema defines dictionary encoding for this struct then `FullEncoding` is 
written with a value of 1. This indicates that the following struct is fully-encoded 
as opposed to being a reference to an existing dictionary entry.

If the schema does not define dictionary encoding for this struct then `FullEncoding` is
not present.

`ModifiedMask` is a field composed of N bits, where N equals the number of fields 
in the struct, with bits laid out in the order matching the declaration order of the 
fields. If the field's value is different in this instance of the struct compared to 
that same field's value in the previous instance of the struct then the corresponding 
bit will be set to 1 otherwise the bit will be set to 0.

Note: here we refer to "value in the previous instance of the struct" instead of 
saying "the value in the previous record". Although, in simplest cases the previous 
instance is indeed in the previous record, in more complex cases it may not be the case.
This is for example possible if the previous record did not contain a value for this 
field at all (this is possible if the field is optional). This is also possible if the 
schema declares an array of structs and encoder is encoding the elements of that array.
In that case "the previous instance" refers to the previous array element (provided 
that the element has a value for that field!).

In short, the encoder is tasked with encoding values of the specified field one after 
another. This values may or may not be sequential values of the same field taken from 
consecutive records. It is the encoder's job to have a state that tracks a "previous 
instance".

`PresenceMask` field is optional and is only present if the struct contains any 
fields that are declared optional in STEF schema. `PresenceMask` will contain M bits, 
where M equals the number of optional fields in the struct, with bits laid out in
the order matching the declaration order of the fields. Note: non-optional fields do 
not have corresponding bits in the `PresenceMask`, so generally speaking M<=N. If the 
optional field is present in this instance of the struct then the corresponding
bit will be set to 1 otherwise the bit will be set to 0.

After the `ModifiedMask` and `PresenceMask` fields are encoded the struct encoder will 
recursively call encoders for struct fields, in the order of field declaration. The 
encoders for fields will then encode field values in their corresponding columns.

If the schema defines dictionary encoding for this struct then this value of the struct
will be added to the dictionary to make it available for future references.

#### Dictionary Struct Encoding

```
Struct {
    FullEncoding: 1 = 0
    RefNum: UC
}
```

If the schema defines dictionary encoding for this struct then the encoder will first 
lookup up the struct's value in the dictionary. If the value is not found in the 
dictionary then [full encoding](#full-struct-encoding) will be used for the struct.

If the struct is found in the dictionary, then a reference to the existing entry will 
be recorded instead.

The `FullEncoding` bit will be set to 0 to indicate the dictionary entry reference 
follows. `RefNum` will be set to the number of the found entry in the dictionary.

### OneOf Codec

```
OneOf {
    ChoiceDelta: SC
}
```

Fields declared in oneof are numbered, starting from number 1 and assigned 
incrementing numbers in the order of declaration. This assigned number is called the 
"oneof choice number". Choice number 0 corresponds to "None" choice.

OneOf codec encodes choice number in the column in delta encoding:
- If this is the first oneof value since encoder was started ChoiceDelta is equal to the 
  choice number.
- For subsequent oneof values ChoiceDelta is the delta between this oneof instance's 
  choice number and the previous oneof instance's choice number.

The chosen field's value is encoded in the child column using field's codec.
After the ChoiceDelta is encoded the OneOf encoder will call the encoders for the 
currently chosen field, that will then encode the chosen field's values in its 
corresponding column. Note that all other fields of the oneof (other than the current 
chosen field) will have nothing appended to their columns.

### Array Codec

```
Array {
    LengthDelta: SC
}
```

Array codec encodes in one column the lengths of the array in delta encoding:
- If this is the first array instance since encoder was inited or was reset then 
  LengthDelta is equal to the length of the array.
- For subsequent array instances LengthDelta is the delta between the array's
  length and the previous array's length.

The array elements are encoded one by one in the child column using child element's codec
using the differential encoding applicable to the child element encoder, i.e. the
array encoder will compare the current element with the previous element and encode only
fields that are modified in the current element compared to the previous element.
This essentially requires the array codec to keep a state of the last value (last 
child element). Note that the "last value" may be a primitive type or may be a composite
type.

The state of the encoder is maintained through arrays, i.e. the first element of an array
is encoded differentially from the last element of the previous array of the same type.

#### Recursive Arrays

When arrays are part of a recursive type definition, a separate state of encoders is 
kept for each recursion level. The description above applies to each of these 
recursion level independently, i.e. the array codec for recursive types maintains
one "last element value" per each recursion level.

Let's illustrate this with an example. Consider the following schema:

```
struct Root root {
  X int64
  A []Root
}
```

Let's say we have the following root record to encode (for brevity we are omitting
`A=[]` when the array is empty).:

```
{
  X=0,
  A=[
    {X=1, A=[X=10,X=11,X=12]},
    {X=1, A=[X=20,X=20,X=22]},
    {X=3, A=[X=30,X=31,X=32]},
  ]
}
```

To encode this data the encoder for `Root` struct will maintain a stack of states 
containing 2 levels, similarly the encoder for `A` array will maintain a stack of states
of 2 levels.

The top-level array, containing X values of (1,1,3) will be encoded using one of the
states, recording the length of the array as +1 to the previous length of 0 (the initial
length - since this is the first array at that level).

The 3 top-level Root structs similarly will be encoded the top-level state of the encoder,
resulting in the second record recognizing that the value of field X is 1, unchanged
from the previous value, which will be reflected in the output of the struct codec
that specifies which fields are modified in the struct.

The 3 second-level arrays, containing X values of (10,11,12), (20,20,22) and (30,
31,32) will be encoded using the second-level state, thus resulting in the output of
delta array lengths of (3,0,0). Note how the length of the top-level array does not 
influence in any way how the length of the second-level arrays is encoded.

The second-level Root structs will be encoded using the second-level state of the
Root struct encoder, resulting in structs containing repeat values of X=20 to
be encoded as unmodified value of field X.

As expected all values for field X are all encoded using the same `int64` codec, the 
values are sent to the codec in the usual depth-first traversal order of the data, 
resulting in the following input to the codec: (1,10,11,12,1,20,20,22,3,30,31,32). 
These values are all encoded in one column used by one `int64` codec since they all 
belong the same field X.

Because codecs traverse data structures in depth-first order, the typical implementation
that maintains the "last value" per recursion level will have a stack of "last values".
The stack grows when the codec detects an increase in recursion level and shrinks
when the codec detects a decrease in recursion level. When the stack shrinks, the element
that was removed from the stack is kept in memory since it will be used as the
"last value" for the next element at the same recursion level.

Note that codec resets (e.g. when the `RestartCodecs` bit is set in the frame) will
clear the stack of "last values" and will start with an empty stack and the state
of all previously seen "last values" will be forgotten.

### MultiMap Codec

MultiMaps have 2 different encodings possible: value-only and full.

#### Full MultiMap Encoding

The following is written to the MultiMap column:

```
MultiMap {
    LengthTransposed: U64
}
```

LengthTransposed is computed as `(Length << 1) | 1`, where `Length` is the 
number of key-value pairs in the MultiMap.

The multimap's keys are encoded in the key child column and values in the value 
child column, in the order they are listed in the multimap.

#### Value-Only MultiMap Encoding

This encoding is used when the current instance of the MultiMap has exactly the same keys 
as the previous instance of the MultiMap. In that case the keys are not encoded at all 
and only values that differ from the previous instance's corresponding value are encoded.

The encoder compares the values of this instance of the MultiMap with the values of 
the previous instance, key-by-key. For all values at index i bit number i is set in a 
value `ChangedKeys`. After all values are iterated `ChangedKeysShifted` is computed 
as `ChangedKeys << 1`.

The following is written to the MultiMap column:

```
MultiMap {
    ChangedKeysShifted: U64
}
```

All values that are different in this instance are then written to the value child 
column. Nothing is written to the key child column.

Value-only encoding can be used if the number of key-value pairs in the MultiMap is 
less than or equal to 62. If the MultiMap has more than 62 key-value pairs then
[Full MultiMap Encoding](#full-multimap-encoding) is always used even if the MultiMap
has the same keys as the previous instance.

### String Codec

String is represented using one of the 2 possible encodings: direct-encoded and by 
reference.

The first field of both encodings is a Varint64 number. After reading this number
the sign of the number will indicate which of the encodings is used:
[direct string](#direct-string) or [string reference](#string-reference).

#### Direct String Encoding

Direct string encoding:

| Field   | Value    | Notes                                                          |
|---------|----------|----------------------------------------------------------------|
| Len     | Varint64 | Length of the string in bytes. A non-negative Varint64 number. |
| Bytes   | Bytes    | Len bytes of the string.                                       |

If the length of the string in bytes is >=2, the encoded string is 
appended to the corresponding dictionary at the current RefNum and the current RefNum 
of the dictionary is incremented. To understand which dictionary is used see the 
description of the block or of the field contains this String value.

#### Reference String Encoding

By reference:

| Field   | Value      | Notes                      |
|---------|------------|----------------------------|
| RefNum  | Varint64=X | A negative Varint64 number |

The RefNum is calculated as RefNum = -X-1. The value of the string is the one that is
stored in the corresponding dictionary at RefNum index.

### Uint64 and Int64 Codecs

```
Uint64 {
    DeltaOfDelta: s64
}
```

Uint64 and Int64 values are represented a delta of delta from the previous value. Given 
previous value `PrevVal` and current value `CurVal`, `DeltaOfDelta` is computed as:

```
Initial state:
PrevDelta = 0
PrevVal = 0

On each iteration:
Delta = CurVal - PrevVal
DeltaOfDelta = Delta - PrevDelta
PrevDelta = Delta
PrevVal = CurVal
```
### Float64 Codec

64-bit IEEE numbers are encoded using 
[Gorilla encoding](https://www.vldb.org/pvldb/vol8/p1816-teller.pdf) (section 4.1.2)

### Bool Codec

Bool values are encoded as single bits.

## Value Formats

### Uvarint64

Unsigned [LEB128 encoded](https://en.wikipedia.org/wiki/LEB128#Unsigned_LEB128)
64-bit number, 1-10 bytes.

### Varint64

Varint64 is a signed integer in [-2^63..2^63-1] range, encoded using
[zigzag encoding](https://en.wikipedia.org/wiki/Variable-length_quantity#Zigzag_encoding)
into [Uvarint64](#uvarint64).

### UvarintCompact

UvarintCompact is an unsigned number in [0..2^48-1] range, written to BitStream
in compact form, in continuous bits in big endian format, the number of total bits
ranges from 1 to 56.

The format is the following:

| Prefix Bits | Followed by big endian bits  |
|-------------|------------------------------|
| 1           | No bits. Encodes value of 0. |
| 01          | 2 bit value                  |
| 001         | 5 bit value                  |
| 0001        | 12 bit value                 |
| 00001       | 19 bit value                 |
| 000001      | 26 bit value                 |
| 0000001     | 33 bit value                 |
| 00000001    | 48 bit value                 |


### VarintCompact

TBD

## Dictionaries

Writer and reader maintain multiple dictionaries that allow referencing previously
seen values.

Each of the dictionaries maintains a set of elements and their associated RefNum.
When STEF data stream is started the dictionaries are empty (contain no elements).

Dictionaries maintained by writer and reader of the STEF data are kept in sync as
they advance through the STEF data stream. This allows the reader to correctly locate 
the dictionary element referenced by RefNum number in the STEF data stream written 
previously by the writer. 

### Resetting Dictionaries

Dictionaries can be reset by the writer. When the dictionaries are reset they are set to
the empty state. Immediately after resetting 
dictionaries the writer MUST close the frame and start a new frame with 
RestartDictionaries bit set in the flags. The RestartDictionaries flag indicates to the 
reader that it MUST also reset its dictionaries.

## STEF/gRPC Protocol

STEF data can be communicated over gRPC via
[sTEFDestination](../go/grpc/proto/destination.proto) service.

`STEFDestination` service is used to deliver STEF data from a Sender to a Receiver.
The Sender is responsible for producing a STEF Data Stream and uses a STEF Writer.
The Receiver correspondingly uses a STEF Reader to decode the received STEF Data Stream.

The sequence diagram of the typical operations is the following:

```
Sender                               Receiver
    
  │             gRPC Connect              │  
  ├──────────────────────────────────────►│  
  │       STEFClientFirstMessage          │  
  ├──────────────────────────────────────►│  
  │     STEFDestinationCapabilities       │  
  │◄──────────────────────────────────────┤  
  │         STEFClientMessage             │  
  ├──────────────────────────────────────►│  
  .                 ...                   .  
  │         STEFDataResponse              │  
  │◄──────────────────────────────────────┤  
  │         STEFClientMessage             │  
  .                 ...                   .  
  ├──────────────────────────────────────►│  
  │         STEFDataResponse              │  
  .                 ...                   .  
```

1. The Sender opens a gRPC connection to the Receiver and starts a Stream() to
   `STEFDestination`.
2. Once the stream is open the Sender MUST send a `STEFClientMessage` with
   `first_message` field set.
3. The Destination MUST reply with a `STEFServerMessage` with `capabilities` field set.
4. The Sender MUST examine received `capabilities` and if the client is
   able to operate as requested by `capabilities` the Sender MUST begin sending
   `STEFClientMessage` containing STEF data. The Destination MUST periodically
   respond with `STEFServerMessage` containing `STEFDataResponse`.
   It is not required that there is always a corresponding `STEFDataResponse` for each
   `STEFClientMessage` - the Receiver may choose to respond once to acknowledge several
   previously received `STEFClientMessage` messages.

### first_message field

The Sender MUST set this field in the the first `STEFClientMessage` sent.
All other fields MUST be unset when `first_message` is set.
All subsequent messages MUST have `first_message` unset.

### stef_bytes field

The field `stef_bytes` contains a sequence of bytes of the STEF stream.

The receiver of the `stef_bytes` field is responsible for assembling the STEF data
stream from a sequence of `STEFClientMessage` messages by concatenating bytes from
`stef_bytes` in the order the messages are received.

### is_end_of_chunk field

The `is_end_of_chunk` field indicates that the last byte of `stef_bytes` field is an
end of a chunk. A chunk is either the [STEF header](#header), or a
[STEF frame](#frame).

This flag normally is used by receivers to accumulates STEF bytes until the end of
the chunk is encountered and only then start decoding the chunk.
Senders MUST ensure they mark this field true at least once in a while otherwise
receivers may never start decoding the data.

### Record Ids

gRPC protocol responses allow acknowledging receipt or specifying that the decoder
failed at a particular place in the STEF stream. This indication is done via record ids.

The encoder and decoder that operate on two ends of the gRPC stream maintain the
record id and increment it precisely in lockstep with every record processed.

The receiver of the STEF stream MUST use the record id value of the decoder and
reference that value in `ack_record_id` field of `STEFDataResponse` to indicate
successful receipt of STEF stream up to that sequence id.

If for whatever reason the decoded stream contains invalid values that cannot be
accepted by the receiver, although STEF format decoding itself worked correctly and the
decoder can continue accepting and decoding the subsequence STEF stream bytes, the
receiver MUST indicate such invalid values back to the sender by referencing the
record id of invalid value in `bad_data_record_ids` of `STEFDataResponse`. The sequence of
operations in this case looks like this:

```
Sender                               Receiver
    
  │        STEFClientMessage              │  
  ├──────────────────────────────────────►│  
  │ STEFDataResponse{bad_data_record_ids} │  
  │◄──────────────────────────────────────┤  
  │        STEFClientMessage              │  
  ├──────────────────────────────────────►│  
  │        STEFClientMessage              │  
  ├──────────────────────────────────────►│  
  .                 ...                   .  
```

As opposed to the above scenario, if the receiver is unable to decode the STEF stream,
such that it loses the synchronization and is unable to continue correctly decoding
the stream, then in addition to reporting the failure via `bad_data_record_ids` the
receiver MUST also close the gRPC stream and let the sender re-open the stream. The
sender in this case MAY NOT re-send the data that was marked as bad via previously
reported `bad_data_record_ids` field. This is important to make sure the receiver's
decoder does not fail again on the same data, causing an infinite loop of failures and
retries. The sequence of operations in this case looks like this:

```
Sender                               Receiver
    
  │        STEFClientMessage              │  
  ├──────────────────────────────────────►│  
  │ STEFDataResponse{bad_data_record_ids} │  
  │◄──────────────────────────────────────┤
  │            gRPC Disconnect            │  
  │◄──────────────────────────────────────┤
  │             gRPC Connect              │  
  ├──────────────────────────────────────►│  
  │       STEFClientFirstMessage          │  
  ├──────────────────────────────────────►│  
  │     STEFDestinationCapabilities       │  
  │◄──────────────────────────────────────┤  
  │         STEFClientMessage             │  
  ├──────────────────────────────────────►│  
  .                 ...                   .  
```

### Streaming

STEF gRPC protocols can keep gRPC stream open and stream new STEF data as it arrives
and is ready to be encoded.

The concept of chunks and indication of chunk end allows STEF decoders to decode
complete chunks that are already received and then yield the execution to the caller
and indicate EOF, then when a new complete chunks arrive and become available the same
decoder can easily resume decoding from where it left. This is particularly important for
efficient implementation of STEF destinations that need to receive thousands of
gRPC connection and process received data as it becomes available without blocking and
without the need for decoder implementation to be able to pause and resume operation
from an arbitrary byte in STEF stream.

### Connection Management

Senders SHOULD periodically close the gRPC stream and open a new one. This is important
to ensure gRPC streams remain balanced in load-balanced scenarios with rolling servers
behind the load balancer.

### Dictionary Limits

Receivers, which anticipate a large number of incoming gRPC streams should take into
account memory usage required per stream. To limit the memory usage the receivers can
indicate dictionary size limits in `STEFDestinationCapabilities` message.

Senders MUST honor dictionary limits returned in DictionaryLimits message by resetting
dictionaries when their total byte size grows to the specified limit. See
[Resetting Dictionaries](#resetting-dictionaries) for explanation on how the
writer (i.e. gRPC Sender) and reader (i.e. gRPC receiver) MUST handle dictionary
resets.

The total byte size of dictionaries is calculated by Sender approximately as the total
number of bytes all elements occupy in memory. This calculation is to certain degree
platform and implementation dependent. Receivers MUST allow for a margin of error in this
calculation and set the limits conservatively.
