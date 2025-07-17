/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import io.grpc.ManagedChannel;
import io.grpc.stub.StreamObserver;
import net.stef.WriterOptions;
import net.stef.grpc.service.destination.*;
import net.stef.schema.Compatibility;
import net.stef.schema.WireSchema;

import java.io.IOException;
import java.util.concurrent.CountDownLatch;
import java.util.logging.Level;
import java.util.logging.Logger;

public class StefClient {

    private static final Logger LOGGER = Logger.getLogger(StefClient.class.getName());

    private final STEFDestinationGrpc.STEFDestinationStub asyncStub;
    private final ClientSchema clientSchema;

    public StefClient(ManagedChannel channel, ClientSchema clientSchema) {
        // Initialize the client with the provided channel and configuration
        this.asyncStub = STEFDestinationGrpc.newStub(channel);
        this.clientSchema = clientSchema;
    }

    public GrpcWriter connect() {
        LOGGER.log(Level.INFO, "Connecting to STEF server...");

        final SchemaWriterOptions schemaWriterOptions = new SchemaWriterOptions();
        final ClientSchema clientSchema = this.clientSchema;
        final CountDownLatch finishLatch = new CountDownLatch(1);

        // Create a Response StreamObserver to handle incoming messages from the server & pass on to async stub
        StreamObserver<STEFClientMessage> requestObserver = asyncStub.stream(new StreamObserver<STEFServerMessage>() {
            @Override
            public void onNext(STEFServerMessage message) {
                LOGGER.log(Level.INFO, "onNext():Processing Server incoming messages");
                if (message.hasCapabilities()) {
                    // Handle capabilities, schema negotiation, etc.
                    STEFDestinationCapabilities caps = message.getCapabilities();
                    WireSchema wireSchema = new WireSchema();
                    try {
                        // Deserialize Server Schema
                        wireSchema.deserialize(caps.getSchema().newInput());
                    } catch (IOException e) {
                        LOGGER.log(Level.WARNING, "onNext(): Error occurred during deserialization of server schema", e);
                        throw new RuntimeException(e);
                    }

                    try {
                        // Check if server schema is backward compatible with client schema.
                        Compatibility compatibility = wireSchema.compatible(clientSchema.wireSchema);
                        switch (compatibility) {
                            case Exact:
                                // Schemas match exactly, nothing else is needed, can start sending data.
                                LOGGER.log(Level.INFO, "onNext():Schemas match exactly, nothing else is needed, can start sending data.");
                            case Superset:
                                // ServerStream schema is superset of client schema. The client MUST specify its schema
                                // in the STEF header.
                                LOGGER.log(Level.INFO, "onNext(): Server schema is superset of client schema");
                                schemaWriterOptions.setSchemaCompatible(true);
                                schemaWriterOptions.setMaxDictBytes(caps.getDictionaryLimits().getMaxDictBytes());
                            case Incompatible:
                                // It is neither exact match nor is server schema a superset, but server schema maybe subset.
                                // Check the opposite direction: if client schema is backward compatible with server schema.
                                LOGGER.log(Level.INFO, "onNext(): if client schema is backward compatible with server schema");
                                Compatibility clientCompatibility = clientSchema.wireSchema.compatible(wireSchema);
                                if (clientCompatibility == Compatibility.Incompatible) {
                                    // Client schema is incompatible with server schema.
                                    LOGGER.log(Level.WARNING, "onNext(): Client schema is incompatible with server schema");
                                    throw new RuntimeException("Client schema is incompatible with server schema");
                                } else if (clientCompatibility == Compatibility.Superset) {
                                    // Client schema is superset of server schema. The client MUST specify its schema
                                    // in the STEF header.
                                    LOGGER.log(Level.INFO, "onNext(): Client schema is superset of server schema");
                                    schemaWriterOptions.setSchemaCompatible(true);
                                    schemaWriterOptions.setMaxDictBytes(caps.getDictionaryLimits().getMaxDictBytes());
                                } else {
                                    LOGGER.log(Level.WARNING, "onNext(): Unknown compatibility: {0}", clientCompatibility);
                                    throw new RuntimeException("Unknown compatibility: " + clientCompatibility);
                                }
                            default:
                                LOGGER.log(Level.WARNING, "onNext(): Unknown compatibility: {0}", compatibility);
                                throw new RuntimeException("Unknown compatibility: " + compatibility);
                        }
                    } catch (IOException e) {
                        LOGGER.log(Level.WARNING, "onNext(): Error occurred during client & server schema compatibility", e);
                        throw new RuntimeException(e);
                    }
                } else {
                    LOGGER.log(Level.WARNING, "onNext(): Received message without capabilities: {0}", message);
                }
            }

            @Override
            public void onError(Throwable t) {
                LOGGER.log(Level.WARNING, "onError(): Error received from server", t);
                finishLatch.countDown();
            }

            @Override
            public void onCompleted() {
                LOGGER.log(Level.INFO, "onCompleted(): Server has completed the stream");
                finishLatch.countDown();
            }
        });

        STEFClientFirstMessage firstMsg = STEFClientFirstMessage.newBuilder()
                .setRootStructName(clientSchema.rootStructName).build();

        STEFClientMessage clientMessage = STEFClientMessage.newBuilder()
                .setFirstMessage(firstMsg)
                .build();

        try {
            LOGGER.log(Level.INFO, "RequestObserver: Client sending first message to server: {0}", clientMessage);
            // Send the message to the server
            requestObserver.onNext(clientMessage);
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Failed to send first message", e);
            requestObserver.onError(e);
            return null; // Exit if we can't send the first message
        }

        // Wait for the server to complete the stream or for an error to occur
        try {
            if (!finishLatch.await(5, java.util.concurrent.TimeUnit.SECONDS)) {
                LOGGER.log(Level.WARNING, "Connection timed out");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            LOGGER.log(Level.SEVERE, "Connection interrupted", e);
        }

        // If the schema is compatible, create a GrpcWriter to write data to the server
        GrpcWriter grpcWriter = null;
        if (schemaWriterOptions.isSchemaCompatible()) {
            WriterOptions wo = WriterOptions.builder()
                    .maxTotalDictSize(schemaWriterOptions.getMaxDictBytes())
                    .includeDescriptor(true)
                    .schema(clientSchema.wireSchema)
                    .build();

            grpcWriter = new GrpcWriter(requestObserver, wo);
        }

        return grpcWriter;
    }
}
