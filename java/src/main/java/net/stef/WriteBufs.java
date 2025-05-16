package net.stef;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.ByteBuffer;

public class WriteBufs {
    private final WriteColumnSet columns = new WriteColumnSet();
    private final BitsWriter tempBuf = new BitsWriter();
    private byte[] bytes = new byte[0];

    public void writeTo(OutputStream buf) throws IOException {
        tempBuf.reset();
        columns.writeSizesTo(tempBuf);
        tempBuf.close();

        long bufSize = tempBuf.toBytes().capacity();
        Serde.writeUvarint(bufSize, buf);

        ByteBuffer src =tempBuf.toBytes();
        buf.write(src.array(), src.arrayOffset(), src.capacity());
        columns.writeDataTo(buf);
    }
}