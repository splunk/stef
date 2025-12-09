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
import java.nio.file.StandardOpenOption;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;


class ReadWriteTest {
    private boolean testCopySeed(long seed) throws IOException {
        boolean retVal = true;

        List<String> files = Arrays.asList(
                "hipstershop-otelmetrics.stefz",
                "hostandcollector-otelmetrics.stefz",
                "astronomy-otelmetrics.stefz"
        );

        for (String file : files) {
            try {
                Path path = Paths.get("../benchmarks/testdata/generated/" + file);
                path = path.toAbsolutePath();
                if (!Files.exists(path)) {
                    System.out.printf("Skipping %s (not found), seed %d\n", file, seed);
                    continue;
                }

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
                        assertEquals(ReadResult.Success, result, "seed " + seed);
                    } catch (EOFException e) {
                        break;
                    }

                    copyModified(stefWriter.record, stefReader.record);
                    stefWriter.write();
                    recCount++;
                }
                stefWriter.flush();
                System.out.printf("%-30s %12d (seed %d)\n", file, cw.getBytes().length, seed);

            } catch (Exception e) {
                System.err.printf("Test failed with seed %d: %s\n", seed, e.getMessage());
                e.printStackTrace();
                retVal = false;
            }
        }

        return retVal;
    }

    @Test
    public void testCopy() throws IOException {
        Path seedFilePath = Paths.get("src/test/resources/seeds/ReadWriteTest_seeds.txt");

        List<Long> seeds = new ArrayList<>();
        if (Files.exists(seedFilePath)) {
            String content = Files.readString(seedFilePath).trim();
            if (!content.isEmpty()) {
                seeds = Arrays.stream(content.split("\n"))
                        .map(String::trim)
                        .filter(s -> !s.isEmpty())
                        .map(Long::parseLong)
                        .collect(Collectors.toList());
            }
        }

        // Test all previously-failing seeds first
        for (Long seed : seeds) {
            System.out.printf("Testing with seed from file: %d\n", seed);
            boolean passed = testCopySeed(seed);
            if (!passed) {
                fail("Previously-failing seed " + seed + " still fails");
            }
        }

        long seed = System.nanoTime();
        System.out.printf("Testing with new random seed: %d\n", seed);

        boolean succeeded = testCopySeed(seed);

        if (!succeeded) {
            System.out.printf("Test failed with seed %d, adding to seed file\n", seed);

            Files.createDirectories(seedFilePath.getParent());

            String seedLine = seed + "\n";
            if (Files.exists(seedFilePath)) {
                Files.writeString(seedFilePath, seedLine,
                        StandardOpenOption.APPEND, StandardOpenOption.CREATE);
            } else {
                Files.writeString(seedFilePath, seedLine, StandardOpenOption.CREATE);
            }
            fail("Test failed with seed " + seed);
        }
    }

    private void copyModified(Metrics dst, Metrics src) {
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

