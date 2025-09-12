package net.stef.benchmarks;

import com.example.oteltef.MetricsReader;
import com.example.oteltef.MetricsWriter;
import net.stef.MemChunkWriter;
import net.stef.ReadOptions;
import net.stef.ReadResult;
import net.stef.WriterOptions;
import org.openjdk.jmh.annotations.*;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.TimeUnit;

@BenchmarkMode(Mode.AverageTime)
@OutputTimeUnit(TimeUnit.NANOSECONDS)
@Warmup(iterations = 1, time = 1)
@Measurement(iterations = 1, time = 1)
@Fork(1) // Set to 0 for debugging
@Threads(1)
@State(Scope.Thread)
public class STEF {
    public static final String stefRefFile = "../benchmarks/testdata/generated/hipstershop-otelmetrics.stefz";
    public static long stefRecordCount = 66816; // Number of records in stefRefFile

    public ByteBuffer stefData;
    @Setup
    public void stefDataSetup() throws IOException {
        // Read the reference STEF file.
        Path path = Paths.get(stefRefFile);
        path = path.toAbsolutePath();
        byte[] stefBytes = Files.readAllBytes(path);

        // Write a copy of STEF data into an in-memory buffer.
        MetricsReader reader = new MetricsReader(new ByteArrayInputStream(stefBytes));
        MemChunkWriter memBuf = new MemChunkWriter();
        MetricsWriter writer = new MetricsWriter(memBuf, WriterOptions.builder().build());
        for (int i = 0; i < stefRecordCount; i++) {
            ReadResult result = reader.read(ReadOptions.none);
            if (result != ReadResult.Success) {
                throw new RuntimeException("Read failed");
            }
            writer.record.copyFrom(reader.record);
            writer.write();
        }

        writer.flush();

        // Keep the in-memory buffer for benchmarks to use as an input.
        stefData = ByteBuffer.wrap(memBuf.getBytes());
    }
    @Benchmark
    public void Read() throws IOException {
        MetricsReader reader = new MetricsReader(new ByteArrayInputStream(stefData.array()));
        for (int i = 0; i < stefRecordCount; i++) {
            ReadResult result = reader.read(ReadOptions.none);
            if (result != ReadResult.Success) {
                throw new RuntimeException("Read failed: " + result);
            }
        }
    }

    @Benchmark
    public void ReadWrite() throws IOException {
        MetricsReader reader = new MetricsReader(new ByteArrayInputStream(stefData.array()));
        MetricsWriter writer = new MetricsWriter(new MemChunkWriter(), WriterOptions.builder().build());
        for (int i = 0; i < stefRecordCount; i++) {
            ReadResult result = reader.read(ReadOptions.none);
            if (result != ReadResult.Success) {
                throw new RuntimeException("Read failed: " + result);
            }
            writer.record.copyFrom(reader.record);
            writer.write();
        }
        writer.flush();
    }
}
