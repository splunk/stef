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
import java.util.concurrent.TimeUnit;

public class StefClient {

    private STEFDestinationGrpc.STEFDestinationStub stub;
    private ClientSchema clientSchema;
    private ClientCallbacks clientCallbacks;

    private StreamObserver<STEFClientMessage> requestObserver;
    private CountDownLatch finishLatch;

    public StefClient(ManagedChannel channel, ClientSchema clientSchema, ClientCallbacks clientCallbacks) {
        // Initialize the client with the provided channel and configuration
        // This is where you would set up your gRPC stubs or other client logic

        // For example:
        this.stub = STEFDestinationGrpc.newStub(channel);
        this.clientSchema = clientSchema;
        this.clientCallbacks = clientCallbacks;
    }

    public void connect() {
        System.out.println("Connecting to STEF server...");
        finishLatch = new CountDownLatch(1);

        // Create a StreamObserver to handle incoming messages from the server
        this.requestObserver = stub.stream(new StreamObserver<STEFServerMessage>() {
            @Override
            public void onNext(STEFServerMessage message) {
                System.out.println("onNext()");
                if (message.hasCapabilities()) {
                    // Handle capabilities, schema negotiation, etc.
                    STEFDestinationCapabilities caps = message.getCapabilities();
                    // Parse schema, check compatibility, set options, etc.
                    // (You would implement schema compatibility logic here)
                } else if (message.hasResponse()) {
                    long ackId = message.getResponse().getAckRecordId();
                    boolean keepGoing = clientCallbacks.onAck.handle(ackId);
                    if (!keepGoing) {
                        disconnect();
                    }
                }
            }

            @Override
            public void onError(Throwable t) {
                System.out.println("OnError()");
                clientCallbacks.onDisconnect.accept(t);
            }

            @Override
            public void onCompleted() {
                System.out.println("onCompleted()");
                clientCallbacks.onDisconnect.accept(null);
            }
        });

        STEFClientFirstMessage firstMsg = STEFClientFirstMessage.newBuilder()
                .setRootStructName(clientSchema.rootStructName).build();

        STEFClientMessage clientMessage = STEFClientMessage.newBuilder()
                .setFirstMessage(firstMsg)
                .build();

        requestObserver.onNext(clientMessage);
    }

    public void disconnect() {
        if (requestObserver != null) {
            requestObserver.onCompleted();
        }
        try {
            finishLatch.await(5, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            System.out.println("InterruptedException while waiting for disconnect: " + e.getMessage());
            Thread.currentThread().interrupt();
        }
    }
}
