
package net.stef;

import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.List;

public class ReadColumnSet {
    private final ReadableColumn column = new ReadableColumn();
    private final List<ReadColumnSet> subColumns = new ArrayList<>();

    public ReadableColumn getColumn() {
        return column;
    }

    public ReadColumnSet addSubColumn() {
        ReadColumnSet subColumn = new ReadColumnSet();
        subColumns.add(subColumn);
        return subColumn;
    }

    public int subColumnLen() {
        return subColumns.size();
    }

    public void readSizesFrom(BitsReader buf) throws IOException {
        long dataSize = buf.readUvarintCompact();

        column.setData(new byte[(int) dataSize]);

        if (dataSize == 0) {
            for (ReadColumnSet subColumn : subColumns) {
                subColumn.resetData();
            }
            return;
        }

        for (ReadColumnSet subColumn : subColumns) {
            subColumn.readSizesFrom(buf);
        }
    }

    public void readDataFrom(InputStream buf) throws IOException {
        byte[] data = column.getData();
        int n = buf.readNBytes(data, 0, data.length);
        if (n!=data.length) {
            throw new IOException("cannot read column data, expected " + data.length + " bytes, got " + n);
        }

        for (ReadColumnSet subColumn : subColumns) {
            subColumn.readDataFrom(buf);
        }
    }

    private static byte[] emptyData = new byte[0];

    public void resetData() {
        column.setData(emptyData);
        for (ReadColumnSet subColumn : subColumns) {
            subColumn.resetData();
        }
    }
}