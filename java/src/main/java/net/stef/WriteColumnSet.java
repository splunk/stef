package net.stef;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.List;

public class WriteColumnSet {
    private ByteBuffer data;
    private final List<WriteColumnSet> subColumns = new ArrayList<>();

    public int totalCount() {
        int count = 0;
        for (WriteColumnSet column : subColumns) {
            count += column.totalCount();
        }
        return count + 1;
    }

    public WriteColumnSet addSubColumn() {
        WriteColumnSet subColumn = new WriteColumnSet();
        subColumns.add(subColumn);
        return subColumn;
    }

    public void setBits(BitsWriter b) {
        b.close();
        this.data = b.toBytes();
        b.reset();
    }

    public void setBytes(BytesWriter b) {
        this.data = b.toBytes();
        b.reset();
    }

    public void writeSizesTo(BitsWriter buf) {
        buf.writeUvarintCompact(data.capacity());

        if (data.capacity() == 0) {
            return;
        }

        for (WriteColumnSet subColumn : subColumns) {
            subColumn.writeSizesTo(buf);
        }
    }

    public void writeDataTo(OutputStream buf) throws IOException {
        buf.write(data.array(), data.arrayOffset(), data.capacity());

        if (data.capacity() == 0) {
            return;
        }

        for (WriteColumnSet subColumn : subColumns) {
            subColumn.writeDataTo(buf);
        }
    }

    public WriteColumnSet at(int index) {
        return subColumns.get(index);
    }
}