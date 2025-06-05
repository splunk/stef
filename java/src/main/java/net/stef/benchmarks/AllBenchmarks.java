package net.stef.benchmarks;

import org.openjdk.jmh.results.Result;
import org.openjdk.jmh.results.RunResult;
import org.openjdk.jmh.runner.Runner;
import org.openjdk.jmh.runner.options.Options;
import org.openjdk.jmh.runner.options.OptionsBuilder;

import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

public class AllBenchmarks {
    public static void main(String[] args) throws Exception {
        Options opt = new OptionsBuilder()
                // Add all benchmark classes here
                .include(STEF.class.getSimpleName())
                .include(BitStream.class.getSimpleName())
                .build();

        Collection<RunResult> results = new Runner(opt).run();

        Map<String, Double> scores = new HashMap<String, Double>();
        String scoreUnit = "";

        // Additionally print time per record for STEF benchmarks
        System.out.println("\nTimes per record:");
        for (RunResult result : results) {
            Result primRes = result.getPrimaryResult();
            if (result.getParams().getBenchmark().contains("STEF.")) {
                String label = primRes.getLabel();
                System.out.printf("STEF.%-20s %.2f %s\n",
                        label,
                        primRes.getScore() / STEF.stefRecordCount,
                        primRes.getScoreUnit());

                scores.put(label, primRes.getScore());
                scoreUnit = primRes.getScoreUnit();
            }
        }

        // We manually calculate Write time as a delta between Read and ReadWrite times.
        double writeTime = scores.get("ReadWrite") - scores.get("Read");

        if (scoreUnit!="") {
            System.out.printf("STEF.%-20s %.2f %s\n",
                    "Write",
                    writeTime / STEF.stefRecordCount,
                    scoreUnit);
        }
    }
}
