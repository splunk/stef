package net.stef;

import net.stef.schema.Compatibility;
import net.stef.schema.WireSchema;

import java.io.BufferedInputStream;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.util.Arrays;

public class BaseReader {
    // Source input stream
    private BufferedInputStream source;

    public final FixedHeader fixedHeader = new FixedHeader();
    public final VarHeader varHeader = new VarHeader();
    private WireSchema schema;

    public final ReadBufs readBufs = new ReadBufs();

    private FrameDecoder frameDecoder = new FrameDecoder();
    public long frameRecordCount;
    public long recordCount;

    public BaseReader(BufferedInputStream source) throws IOException {
        this.source = source;
        readFixedHeader();
        frameDecoder.init(this.source, fixedHeader.getCompression());
    }

    private void readFixedHeader() throws IOException {
        byte[] hdrSignature = new byte[Constants.HdrSignature.length];
        if (source.read(hdrSignature) != hdrSignature.length) {
            throw new IOException("Failed to read header signature");
        }
        if (!Arrays.equals(hdrSignature, Constants.HdrSignature)) {
            throw new IOException("Invalid header signature");
        }

        long contentSize = Serde.readUvarint(source);
        if (contentSize < 2 || contentSize > Limits.HdrContentSizeLimit) {
            throw new IOException("Invalid header content size");
        }

        byte[] hdrContent = new byte[(int) contentSize];
        if (source.read(hdrContent) != contentSize) {
            throw new IOException("Failed to read header content");
        }

        byte versionAndType = hdrContent[0];
        byte version = (byte) (versionAndType & Constants.HdrFormatVersionMask);
        if (version != Constants.HdrFormatVersion) {
            throw new IOException("Invalid format version");
        }

        byte flags = hdrContent[1];

        fixedHeader.setCompression(Compression.fromValue((byte) (flags & Constants.HdrFlagsCompressionMethod)));
        if (!(fixedHeader.getCompression() == Compression.None || fixedHeader.getCompression() == Compression.Zstd)) {
            throw new IOException("Invalid compression");
        }
    }

    public void readVarHeader(WireSchema ownSchema) throws IOException {
        frameDecoder.next();

        long remaining = frameDecoder.getRemainingSize();
        byte[] hdrBytes = new byte[(int) remaining];
        int n = frameDecoder.read(hdrBytes);
        if (n < 0) throw new IOException("Failed to read var header");

        ByteArrayInputStream buf = new ByteArrayInputStream(hdrBytes, 0, n);
        varHeader.deserialize(buf);

        if (varHeader.getSchemaWireBytes() != null && varHeader.getSchemaWireBytes().length != 0) {
            buf = new ByteArrayInputStream(varHeader.getSchemaWireBytes());
            schema = new WireSchema();
            schema.deserialize(buf);
            if (ownSchema.compatible(schema)== Compatibility.Incompatible) {
                throw new IOException("Schema is not compatible with BaseReader");
            }
        }
    }

    public int nextFrame() throws IOException {
        int frameFlags = frameDecoder.next();
        frameRecordCount = Serde.readUvarint(frameDecoder);
        readBufs.readFrom(frameDecoder);
        return frameFlags;
    }

    public WireSchema getSchema() {
        return schema;
    }
}

