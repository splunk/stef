/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

public class SchemaWriterOptions {

    private boolean isSchemaCompatible;
    private long maxDictBytes;

    boolean isSchemaCompatible() {
        return isSchemaCompatible;
    }

    long getMaxDictBytes() {
        return maxDictBytes;
    }

    void setSchemaCompatible(boolean schemaCompatible) {
        isSchemaCompatible = schemaCompatible;
    }

    void setMaxDictBytes(long maxDictBytes) {
        this.maxDictBytes = maxDictBytes;
    }
}
