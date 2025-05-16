package net.stef;

import com.github.luben.zstd.ZstdInputStream;

import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;

public class FrameDecoder {
    private ByteAndBlockReader src;
    private Compression compression;
    private long uncompressedSize;
    private int ofs;
    private InputStream frameContentSrc;
    private ZstdInputStream decompressor;
    private ChunkedReader chunkReader = new ChunkedReader();
    private int flags;
    private boolean frameLoaded;
    private boolean notFirstFrame;

    public void init(ByteAndBlockReader src, Compression compression) throws IOException {
        this.src = src;
        this.compression = compression;
        chunkReader.init(src);
        chunkReader.setNextChunk(this::nextFrame);

        switch (compression) {
            case NONE:
                this.frameContentSrc = chunkReader;
                break;
            case ZSTD:
                this.decompressor = new ZstdInputStream(chunkReader);
                this.frameContentSrc = decompressor;
                break;
            default:
                throw new IllegalArgumentException("Unknown compression: " + compression);
        }
    }

    private void nextFrame() throws IOException {
        int hdrByte = src.read();
        if (hdrByte == -1) {
            throw new EOFException();
        }
        this.flags = hdrByte;

        if (!FrameFlags.isValid(flags)) {
            throw new IOException("Invalid frame flags");
        }

        this.uncompressedSize = Serde.readUvarint(src);

        if (compression != Compression.NONE) {
            long compressedSize = Serde.readUvarint(src);
            chunkReader.setLimit(compressedSize);

            if (!notFirstFrame || (flags & FrameFlags.RESTART_COMPRESSION)!=0) {
                notFirstFrame = true;
                decompressor.close();
                decompressor = new ZstdInputStream(chunkReader);
                this.frameContentSrc = decompressor;
            }
        } else {
            chunkReader.setLimit(uncompressedSize);
        }

        frameLoaded = true;
        ofs = 0;
    }

    /*
        Returns the frame flags.
     */
    public int next() throws IOException {
        while (uncompressedSize > 0) {
            byte[] tmp = new byte[4096];
            int readSize = (int) Math.min(uncompressedSize, tmp.length);
            int n = frameContentSrc.read(tmp, 0, readSize);
            if (n == -1) {
                throw new IOException("Unexpected end of frame");
            }
            uncompressedSize -= n;
            ofs += n;
        }

        nextFrame();
        return flags;
    }

    public long getRemainingSize() {
        return uncompressedSize;
    }

    public int read(byte[] buffer) throws IOException {
        if (uncompressedSize == 0) {
            frameLoaded = false;
            throw new IOException("End of frame");
        }

        int toRead = (int) Math.min(uncompressedSize, buffer.length);
        int n = frameContentSrc.read(buffer, 0, toRead);
        if (n == -1) {
            throw new IOException("Unexpected end of frame");
        }

        uncompressedSize -= n;
        ofs += n;
        return n;
    }

    public int readByte() throws IOException {
        if (uncompressedSize == 0) {
            frameLoaded = false;
            throw new IOException("End of frame");
        }

        uncompressedSize--;
        ofs++;
        int b= frameContentSrc.read();
        if (b==-1) {
            throw new EOFException();
        }
        return b;
    }
}
