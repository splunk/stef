// Code generated by stefgen. DO NOT EDIT.

package com.example.oteltef;

import java.util.*;

// MetricDecoderDict is the dictionary used by MetricDecoder
class MetricDecoderDict {
    private final List<Metric> dict = new ArrayList<>();

    public void init() {
        this.dict.clear();
        this.dict.add(null); // null Metric is RefNum 0
    }

    // Reset the dictionary to initial state. Used when a frame is
    // started with RestartDictionaries flag.
    public void reset() {
        this.init();
    }

    public Metric getByIndex(int index) {
        return this.dict.get(index);
    }

    public void add(Metric val) {
        this.dict.add(val);
    }

    public int size() {
        return this.dict.size();
    }
}