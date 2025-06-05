package net.stef.benchmarks;

import net.stef.BitsReader;
import net.stef.BitsWriter;
import org.openjdk.jmh.annotations.*;

import java.nio.ByteBuffer;
import java.util.concurrent.TimeUnit;

@BenchmarkMode(Mode.AverageTime)
@OutputTimeUnit(TimeUnit.NANOSECONDS)
@Warmup(iterations = 1, time = 1)
@Measurement(iterations = 1, time = 1)
@Fork(1) // Set to 0 for debugging
@Threads(1)
@State(Scope.Thread)
public class BitStream {
    @Benchmark
    public void bstreamWriteBit() {
        BitsWriter bw = new BitsWriter();
        for (int j = 0; j < 1_000_000; j++) {
            bw.writeBit(j % 2);
        }
    }

    public ByteBuffer bstreamReadBitData;
    @Setup
    public void bstreamReadBitSetup() {
        BitsWriter bw = new BitsWriter();
        for (int j = 0; j < 1_000_000; j++) {
            bw.writeBit(j % 2);
        }
        bw.close();
        bstreamReadBitData = bw.toBytes();
    }
    @Benchmark
    public void bstreamReadBit() {
        BitsReader br = new BitsReader();
        br.reset(bstreamReadBitData);
        for (int j = 0; j < 1_000_000; j++) {
            long v = br.readBit();
            if (v != (j % 2)) {
                throw new RuntimeException("invalid value");
            }
        }
    }

    public ByteBuffer bstreamReadBitsData;
    @Setup
    public void bstreamReadBitsSetup() {
        BitsWriter bw = new BitsWriter();
        long val = 1L;
        for (int j = 1; j < 64; j++) {
            bw.writeBits(val, j);
            val *= 2;
        }
        bw.close();
        bstreamReadBitsData = bw.toBytes();
    }
    @Benchmark
    @OutputTimeUnit(TimeUnit.NANOSECONDS)
    public void bstreamReadBits() {
        BitsReader br = new BitsReader();
        br.reset(bstreamReadBitsData);
        long val = 1L;
        for (int j = 1; j < 64; j++) {
            long v = br.readBits(j);
            if (v != val) {
                throw new RuntimeException("mismatch");
            }
            val *= 2;
        }
    }
}
