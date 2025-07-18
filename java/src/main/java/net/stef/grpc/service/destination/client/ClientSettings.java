/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import net.stef.grpc.service.destination.STEFDestinationGrpc;

public class ClientSettings {
    private STEFDestinationGrpc.STEFDestinationStub grpcClient;
    private ClientSchema clientSchema;
    private ClientCallbacks callbacks;

    public ClientSettings(STEFDestinationGrpc.STEFDestinationStub grpcClient, ClientSchema clientSchema, ClientCallbacks callbacks) {
        this.grpcClient = grpcClient;
        this.clientSchema = clientSchema;
        this.callbacks = callbacks;
    }

    public STEFDestinationGrpc.STEFDestinationStub getGrpcClient() {
        return grpcClient;
    }

    public ClientSchema getClientSchema() {
        return clientSchema;
    }

    public ClientCallbacks getCallbacks() {
        return callbacks;
    }
}