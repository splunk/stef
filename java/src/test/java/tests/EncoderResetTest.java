package tests;

import com.example.otelstef.AnyValue;
import com.example.otelstef.LogsWriter;
import com.example.otelstef.Metrics;
import com.example.otelstef.MetricsReader;
import com.example.otelstef.MetricsWriter;
import com.example.otelstef.PointValue;
import com.example.otelstef.QuantileValueArray;
import net.stef.ChunkWriter;
import net.stef.FrameFlags;
import net.stef.MemChunkWriter;
import net.stef.ReadOptions;
import net.stef.ReadResult;
import net.stef.StringValue;
import net.stef.WriterOptions;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class EncoderResetTest {
    private static AnyValue newAnyValue() throws Exception {
        Constructor<AnyValue> constructor = AnyValue.class.getDeclaredConstructor();
        constructor.setAccessible(true);
        return constructor.newInstance();
    }

    private static class CapturingChunkWriter implements ChunkWriter {
        private final List<byte[]> chunks = new ArrayList<>();

        @Override
        public void writeChunk(byte[] header, byte[] content) throws IOException {
            ByteArrayOutputStream chunk = new ByteArrayOutputStream();
            chunk.write(header);
            if (content != null) {
                chunk.write(content);
            }
            chunks.add(chunk.toByteArray());
        }

        byte[] streamWithLastDataFrame() throws IOException {
            // A standalone stream consists of the fixed header, variable-header frame,
            // and one independently decodable data frame.
            ByteArrayOutputStream stream = new ByteArrayOutputStream();
            stream.write(chunks.get(0));
            stream.write(chunks.get(1));
            stream.write(chunks.get(chunks.size() - 1));
            return stream.toByteArray();
        }
    }

    @Test
    void forceModifiedFieldsIsConsumedByNextWrite() throws Exception {
        LogsWriter writer = new LogsWriter(new MemChunkWriter(), WriterOptions.builder().build());

        Field encoderField = LogsWriter.class.getDeclaredField("encoder");
        encoderField.setAccessible(true);
        Object encoder = encoderField.get(writer);

        Method reset = encoder.getClass().getDeclaredMethod("reset");
        reset.setAccessible(true);
        reset.invoke(encoder);

        writer.write();

        Field forceModifiedFields = encoder.getClass().getDeclaredField("forceModifiedFields");
        forceModifiedFields.setAccessible(true);
        assertFalse(forceModifiedFields.getBoolean(encoder));
    }

    @Test
    void restartedFrameFullyEncodesEveryCollectionElement() throws Exception {
        CapturingChunkWriter output = new CapturingChunkWriter();
        int restartFlags = FrameFlags.RestartDictionaries | FrameFlags.RestartCodecs;
        MetricsWriter writer = new MetricsWriter(
            output,
            WriterOptions.builder().frameRestartFlags(restartFlags).build()
        );

        writer.record.getPoint().getValue().setType(PointValue.Type.TypeSummary);
        QuantileValueArray quantiles = writer.record.getPoint().getValue().getSummary().getQuantileValues();
        quantiles.ensureLen(2);
        quantiles.at(0).setQuantile(0.5);
        quantiles.at(0).setValue(10.0);
        quantiles.at(1).setQuantile(0.99);
        quantiles.at(1).setValue(20.0);

        AnyValue firstValue = newAnyValue();
        firstValue.setInt64(1);
        writer.record.getAttributes().append(new StringValue("first"), firstValue);
        AnyValue secondValue = newAnyValue();
        secondValue.setInt64(2);
        writer.record.getAttributes().append(new StringValue("second"), secondValue);

        Metrics expected = writer.record.clone();
        writer.write();
        // Write the unchanged record again. Its modification masks are now clear,
        // but the restarted frame must still contain a complete representation.
        writer.write();

        MetricsReader reader = new MetricsReader(
            new ByteArrayInputStream(output.streamWithLastDataFrame())
        );
        assertEquals(ReadResult.Success, reader.read(ReadOptions.none));
        assertTrue(reader.record.equals(expected));
    }
}
