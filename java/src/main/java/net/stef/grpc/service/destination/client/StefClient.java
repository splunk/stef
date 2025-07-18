/*
 * Copyright (c) AppDynamics, Inc., and its affiliates
 * 2025
 * All Rights Reserved
 * THIS IS UNPUBLISHED PROPRIETARY CODE OF APPDYNAMICS, INC.
 * The copyright notice above does not evidence any actual or intended publication of such source code
 */

package net.stef.grpc.service.destination.client;

import io.grpc.stub.StreamObserver;
import net.stef.WriterOptions;
import net.stef.grpc.service.destination.*;
import net.stef.schema.Compatibility;
import net.stef.schema.WireSchema;

import java.io.IOException;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicReference;
import java.util.logging.Level;
import java.util.logging.Logger;

public class StefClient {

    private static final Logger LOGGER = Logger.getLogger(StefClient.class.getName());

    private final STEFDestinationGrpc.STEFDestinationStub asyncStub;
    private final ClientSchema clientSchema;
    private final ClientCallbacks clientCallbacks;
    private StreamObserver<STEFClientMessage> requestObserver;

    public StefClient(ClientSettings settings) {

        if (isRootStructNameEmpty(settings)) {
            throw new IllegalArgumentException("Client schema root struct name is empty");
        }

        if (isClientSchemaWireSchemaNull(settings)) {
            throw new IllegalArgumentException("In client schema wire schema is null");
        }

        this.asyncStub = settings.getGrpcClient();
        this.clientSchema = settings.getClientSchema();
        this.clientCallbacks = settings.getCallbacks();
    }

    /**
     * Establishes a connection to the server and sets up a {@link GrpcWriter} if the schemas are compatible.
     * <p>
     * This method will block until the connection is established or a timeout occurs.
     * <p>
     * The returned {@link GrpcWriter} can be used to write data to the server.
     * <p>
     * If the schemas are incompatible, or if an error occurs during the connection process, this method will return
     * {@code null}.
     *
     * @return A {@link GrpcWriter} if the schemas are compatible, or {@code null} if not.
     */
    public GrpcWriter connect() {
        LOGGER.log(Level.INFO, "Connecting to STEF server...");

        GrpcWriter grpcWriter = null;
        final ClientSchema clientSchema = this.clientSchema;
        final CountDownLatch finishLatch = new CountDownLatch(1);
        final AtomicReference<STEFServerMessage> serverMessageAtomicReference = new AtomicReference<>();
        final AtomicReference<Throwable> errorAtomicReference = new AtomicReference<>();

        // Create a Response StreamObserver to handle incoming messag   es from the server & pass on to async stub
        this.requestObserver = asyncStub.stream(new StreamObserver<STEFServerMessage>() {
            @Override
            public void onNext(STEFServerMessage message) {
                LOGGER.log(Level.INFO, "onNext():Processing Server incoming messages");
                serverMessageAtomicReference.set(message);
                finishLatch.countDown();
            }

            @Override
            public void onError(Throwable t) {
                LOGGER.log(Level.WARNING, "onError(): Error received from server", t);
                errorAtomicReference.set(t);
                clientCallbacks.onDisconnect.accept(t);
                finishLatch.countDown();
            }

            @Override
            public void onCompleted() {
                LOGGER.log(Level.INFO, "onCompleted(): Server has completed the stream");
                clientCallbacks.onDisconnect.accept(null);
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

            LOGGER.log(Level.INFO, "RequestObserver: Client sent first message to server: {0}", clientMessage);

            // Wait for the server to complete the stream or for an error to occur
            if (!finishLatch.await(5, java.util.concurrent.TimeUnit.SECONDS)) {
                LOGGER.log(Level.WARNING, "Connection timed out");
            }

            // Check for errors
            Throwable error = errorAtomicReference.get();
            if (error != null) {
                throw new IOException("Failed to receive capabilities", error);
            }

            STEFServerMessage serverMessage = serverMessageAtomicReference.get();

            if (serverMessage == null || !serverMessage.hasCapabilities()) {
                LOGGER.log(Level.WARNING, "onNext(): Received message without capabilities: {0}", serverMessage);
                throw new RuntimeException("onNext(): Invalid Server response");
            } else if (serverMessage.hasResponse()) {
                long ackId = serverMessage.getResponse().getAckRecordId();
                boolean keepGoing = clientCallbacks.onAck.handle(ackId);
                if (!keepGoing) {
                    disconnect();
                }
            }

            WireSchema wireSchema = new WireSchema();
            STEFDestinationCapabilities capabilities = serverMessage.getCapabilities();

            // Deserialize the server schema
            wireSchema.deserialize(capabilities.getSchema().newInput());

            // Check if the schemas are compatible
            Compatibility compatibility = compatible(wireSchema, clientSchema);

            // Build the schema writer options
            WriterOptions writerOptions = buildSchemaWriterOptions(compatibility, capabilities);

            // If the schema is compatible, create a GrpcWriter to write data to the server
            grpcWriter = new GrpcWriter(requestObserver, writerOptions);

        } catch (IOException e) {
            LOGGER.log(Level.WARNING, "onNext(): Error occurred during schema processing", e);
            requestObserver.onError(e);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            LOGGER.log(Level.SEVERE, "Connection interrupted", e);
            requestObserver.onError(e);
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, "Failed to send first message", e);
            requestObserver.onError(e);
        }

        return grpcWriter;
    }

    /**
     * Closes the connection to the server.
     *
     * <p>This method does not throw any checked exceptions. If there is an error
     * closing the connection, it is logged at the {@link Level#WARNING} level.
     */
    public void disconnect() {
        if (this.requestObserver != null) {
            this.requestObserver.onCompleted();
        }
    }

    /**
     * Determines the compatibility between the server's WireSchema and the client's ClientSchema.
     *
     * <p>This method first checks if the server's WireSchema is compatible with the client's schema.
     * If they match exactly, it logs the match and returns {@link Compatibility#Exact}.
     * If the server schema is a superset of the client schema, it logs this information and
     * returns {@link Compatibility#Superset}.
     *
     * <p>If the server's schema is neither exact nor a superset, the method checks if the client's schema
     * is backward compatible with the server's schema. If the schemas are incompatible, it logs
     * the error and throws a {@link RuntimeException}. If the client schema is a superset of the server's
     * schema, it logs this status and returns {@link Compatibility#Superset}. Any unknown compatibility
     * status results in a warning log and a thrown {@link RuntimeException}.
     *
     * @param wireSchema   the server's schema to compare
     * @param clientSchema the client's schema to compare
     * @return the determined compatibility between the client and server schemas
     * @throws IOException      if there is an error during compatibility checking
     * @throws RuntimeException if schemas are incompatible or an unknown compatibility status is encountered
     */
    private Compatibility compatible(WireSchema wireSchema, ClientSchema clientSchema) throws IOException, RuntimeException {
        Compatibility compatibility = wireSchema.compatible(clientSchema.wireSchema);
        if (compatibility == Compatibility.Exact) {
            LOGGER.log(Level.INFO, "onNext():Schemas match exactly, can start sending data.");
            return Compatibility.Exact;
        } else if (compatibility == Compatibility.Superset) {
            LOGGER.log(Level.INFO, "onNext(): Server schema is superset of client schema");
            return Compatibility.Superset;
        } else {
            LOGGER.log(Level.INFO, "onNext(): Checking if client schema is backward compatible with server schema");
            Compatibility clientCompatibility = clientSchema.wireSchema.compatible(wireSchema);

            if (clientCompatibility == Compatibility.Incompatible) {
                LOGGER.log(Level.SEVERE, "onNext(): client and server schemas are incompatible");
                throw new RuntimeException("Client and Server schemas are incompatible");
            } else if (clientCompatibility == Compatibility.Superset) {
                LOGGER.log(Level.INFO, "onNext(): Client schema is superset of server schema");
                return Compatibility.Superset;
            } else {
                LOGGER.log(Level.WARNING, "onNext(): Unknown compatibility: {0}", clientCompatibility);
                throw new RuntimeException("Unknown compatibility: " + clientCompatibility);
            }
        }
    }

    /**
     * Returns a {@link WriterOptions} for writing records to the server, based on
     * the compatibility of the client and server schemas.
     *
     * <p>If the compatibility is a superset, the returned {@link WriterOptions} will
     * include the descriptor and have the maximum total dictionary size set to the
     * value returned by {@link STEFDestinationCapabilities#getDictionaryLimits()}.
     *
     * <p>If the compatibility is not a superset, the returned {@link WriterOptions}
     * will be the default options.
     *
     * @param compatibility the compatibility of the client and server schemas
     * @param capabilities  the capabilities of the server
     * @return a {@link WriterOptions} for writing records to the server
     */
    private WriterOptions buildSchemaWriterOptions(Compatibility compatibility, STEFDestinationCapabilities capabilities) {
        if (compatibility == Compatibility.Superset) {
            return WriterOptions.builder()
                    .maxTotalDictSize(capabilities.getDictionaryLimits().getMaxDictBytes())
                    .includeDescriptor(true)
                    .schema(clientSchema.wireSchema)
                    .build();
        }
        return WriterOptions.builder().build();
    }

    private boolean isRootStructNameEmpty(ClientSettings settings) {
        return settings.getClientSchema().rootStructName == null || settings.getClientSchema().rootStructName.isEmpty();
    }

    private boolean isClientSchemaWireSchemaNull(ClientSettings settings) {
        return settings.getClientSchema().wireSchema == null;
    }
}
