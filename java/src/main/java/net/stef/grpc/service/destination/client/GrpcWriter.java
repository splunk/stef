/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import io.grpc.stub.StreamObserver;
import net.stef.ChunkWriter;
import net.stef.WriterOptions;
import net.stef.grpc.service.destination.STEFClientMessage;

import java.io.IOException;

public class GrpcWriter implements ChunkWriter {
    private StreamObserver<STEFClientMessage> stream;
    private WriterOptions writerOptions;

    GrpcWriter(StreamObserver<STEFClientMessage> stream, WriterOptions writerOptions) {
        this.stream = stream;
        this.writerOptions = writerOptions;
    }

    @Override
    public void writeChunk(byte[] header, byte[] content) throws IOException {
        // Build the message with header and content
        STEFClientMessage.Builder builder = STEFClientMessage.newBuilder();
        builder.clearStefBytes();
        builder.setStefBytes(com.google.protobuf.ByteString.copyFrom(header));
        builder.setStefBytes(com.google.protobuf.ByteString.copyFrom(content));
        builder.setIsEndOfChunk(true);

        // TODO: split the chunk into multiple messages if it is too big to fit in one gRPC message.

        try {
            stream.onNext(builder.build());
        } catch (Exception e) {
            throw new IOException("stef GrpcWrite error: " + e.getMessage(), e);
        }
    }
}