package net.stef;

import com.github.luben.zstd.ZstdOutputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;

public class FrameEncoder {
    private ChunkWriter dest;
    private OutputStream frameContent;
    private int uncompressedSize;
    private ByteArrayOutputStream compressedBuf = new ByteArrayOutputStream();
    private Compression compression;
    private ZstdOutputStream compressor;
    private int hdrByte;

    public void init(ChunkWriter dest, Compression compression) throws IOException {
        this.compression = compression;
        this.dest = dest;

        switch (compression) {
            case NONE:
                this.frameContent = compressedBuf;
                break;
            case ZSTD:
                this.compressor = new ZstdOutputStream(compressedBuf);
                this.frameContent = compressor;
                break;
            default:
                throw new IllegalArgumentException("Unknown compression: " + compression);
        }
    }

    public void openFrame(int resetFlags) throws IOException {
        this.hdrByte = resetFlags;
        if ((resetFlags & FrameFlags.RESTART_COMPRESSION)!=0 && compressor != null) {
            compressor.close();
            this.compressor = new ZstdOutputStream(compressedBuf);
            this.frameContent = compressor;
        }
    }

    public void closeFrame() throws IOException {
        ByteArrayOutputStream frameHdr = new ByteArrayOutputStream();
        frameHdr.write(hdrByte);
        Serde.writeUvarint(uncompressedSize, frameHdr);

        if (compression == Compression.ZSTD) {
            compressor.close();
            Serde.writeUvarint(compressedBuf.size(), frameHdr);
        }

        dest.writeChunk(frameHdr.toByteArray(), compressedBuf.toByteArray());
        compressedBuf.reset();
        uncompressedSize = 0;
    }

    public int write(byte[] data) throws IOException {
        frameContent.write(data);
        uncompressedSize += data.length;
        return data.length;
    }

    public int getUncompressedSize() {
        return uncompressedSize;
    }
}