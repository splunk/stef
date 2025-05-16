package net.stef;

import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteBuffer;

public class ReadBufs {
    private final ReadColumnSet columns = new ReadColumnSet();
    private final BitsReader tempBuf = new BitsReader();
    private byte[] tempBufBytes = new byte[0];

    public void readFrom(InputStream buf) throws IOException {
        long bufSize = Serde.readUvarint(buf);
        tempBufBytes = new byte[(int) bufSize];
        buf.read(tempBufBytes);
        tempBuf.reset(ByteBuffer.wrap(tempBufBytes));

        columns.readSizesFrom(tempBuf);
        columns.readDataFrom(buf);
    }
}