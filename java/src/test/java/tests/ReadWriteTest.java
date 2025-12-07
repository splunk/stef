package tests;

import com.example.otelstef.Metrics;
import com.example.otelstef.MetricsReader;
import com.example.otelstef.MetricsWriter;
import net.stef.MemChunkWriter;
import net.stef.ReadOptions;
import net.stef.ReadResult;
import net.stef.WriterOptions;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.EOFException;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;

class ReadWriteTest {
    @Test
    public void testCopy() throws IOException {
        List<String> files = Arrays.asList(
            "hipstershop-otelmetrics.stefz",
            "hostandcollector-otelmetrics.stefz",
            "astronomy-otelmetrics.stefz"
        );

        System.out.printf("%-30s %12s\n", "File", "Uncompressed");

        for (String file : files) {
            Path path = Paths.get("../benchmarks/testdata/generated/" + file);
            path = path.toAbsolutePath();
            byte[] stefBytes = Files.readAllBytes(path);
            assertNotNull(stefBytes);

            // Read using MetricsReader
            MetricsReader stefReader = new MetricsReader(new ByteArrayInputStream(stefBytes));
            MemChunkWriter cw = new MemChunkWriter();
            MetricsWriter stefWriter = new MetricsWriter(cw, WriterOptions.builder().build());


            int recCount = 0;
            while (true) {
                try {
                    ReadResult result = stefReader.read(ReadOptions.none);
                    assertEquals(ReadResult.Success, result);
                } catch (EOFException e) {
                    break;
                }

                copyModified(stefWriter.record, stefReader.record);

                // Optionally, write the record to the writer
                stefWriter.write();
                recCount++;
            }
            stefWriter.flush();
            System.out.printf("%-30s %12d\n", file, cw.getBytes().length);
        }
    }

    void copyModified(Metrics dst, Metrics src) {
        if (src.isEnvelopeModified()) {
            dst.getEnvelope().copyFrom(src.getEnvelope());
        }

        if (src.isResourceModified()) {
            dst.getResource().copyFrom(src.getResource());
        }

        if (src.isScopeModified()) {
            dst.getScope().copyFrom(src.getScope());
        }

        if (src.isMetricModified()) {
            dst.getMetric().copyFrom(src.getMetric());
        }

        if (src.isAttributesModified()) {
            dst.getAttributes().copyFrom(src.getAttributes());
        }

        if (src.isPointModified()) {
            dst.getPoint().copyFrom(src.getPoint());
        }
    }

}

