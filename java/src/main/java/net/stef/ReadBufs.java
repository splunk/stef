package net.stef;

import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteBuffer;

public class ReadBufs {
    public final ReadColumnSet columns = new ReadColumnSet();
    private final BitsReader tempBuf = new BitsReader();
    private byte[] tempBufBytes = new byte[0];

    public void readFrom(InputStream buf) throws IOException {
        long bufSize = Serde.readUvarint(buf);
        tempBufBytes = new byte[(int) bufSize];
        int n = buf.readNBytes(tempBufBytes, 0, (int) bufSize);
        if (n!=(int)bufSize) {
            throw new IOException("Failed to read expected number of bytes from input stream: " + n + " != " + bufSize);
        }

        tempBuf.reset(ByteBuffer.wrap(tempBufBytes));

        columns.readSizesFrom(tempBuf);
        columns.readDataFrom(buf);
    }
}