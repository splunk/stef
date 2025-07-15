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
import net.stef.grpc.service.destination.*;

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

    public void connect() {
        LOGGER.log(Level.INFO, "Connecting to STEF server...");

        final CountDownLatch finishLatch = new CountDownLatch(1);

        // Create a Response StreamObserver to handle incoming messages from the server & pass on to async stub
        StreamObserver<STEFClientMessage> requestObserver = asyncStub.stream(new StreamObserver<STEFServerMessage>() {
            @Override
            public void onNext(STEFServerMessage message) {
                LOGGER.log(Level.INFO, "onNext():Processing Server incoming messages");
                if (message.hasCapabilities()) {
                    // Handle capabilities, schema negotiation, etc.
                    STEFDestinationCapabilities caps = message.getCapabilities();
                    // Parse schema, check compatibility, set options, etc.
                    // (You would implement schema compatibility logic here)
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
            // Send the message to the server
            requestObserver.onNext(clientMessage);
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Failed to send first message", e);
            requestObserver.onError(e);
            return; // Exit if we can't send the first message
        }

        // Wait for the server to complete the stream or for an error to occur
        try {
            if (!finishLatch.await(5, java.util.concurrent.TimeUnit.SECONDS)) {
                LOGGER.log(Level.WARNING, "Connection timed out");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            LOGGER.log(Level.SEVERE, "Connection interrupted", e);
        } finally {
            LOGGER.log(Level.INFO, "RequestObserver: Client Completing the stream");
            requestObserver.onCompleted();
            LOGGER.log(Level.INFO, "RequestObserver: Client Completed the stream");
        }
    }
}
