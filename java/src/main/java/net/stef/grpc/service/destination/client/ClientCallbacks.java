/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import java.util.function.Consumer;

public class ClientCallbacks {
        public Consumer<Throwable> onDisconnect = err -> {};
        public AckHandler onAck = ackId -> true; // return false to disconnect

        @FunctionalInterface
        public interface AckHandler {
            boolean handle(long ackId);
        }
    }