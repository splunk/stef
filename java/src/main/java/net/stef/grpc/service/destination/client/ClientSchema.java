/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import net.stef.schema.WireSchema;

public class ClientSchema {
    public final String rootStructName;
    public final WireSchema wireSchema;

    public ClientSchema(String rootStructName, WireSchema wireSchema) {
        this.rootStructName = rootStructName;
        this.wireSchema = wireSchema;
    }
}