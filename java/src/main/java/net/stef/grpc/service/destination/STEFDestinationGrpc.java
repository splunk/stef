package net.stef.grpc.service.destination;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Destination is a service to which STEF data can be sent.
 * </pre>
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class STEFDestinationGrpc {

  private STEFDestinationGrpc() {}

  public static final java.lang.String SERVICE_NAME = "STEFDestination";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<net.stef.grpc.service.destination.STEFClientMessage,
      net.stef.grpc.service.destination.STEFServerMessage> getStreamMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Stream",
      requestType = net.stef.grpc.service.destination.STEFClientMessage.class,
      responseType = net.stef.grpc.service.destination.STEFServerMessage.class,
      methodType = io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
  public static io.grpc.MethodDescriptor<net.stef.grpc.service.destination.STEFClientMessage,
      net.stef.grpc.service.destination.STEFServerMessage> getStreamMethod() {
    io.grpc.MethodDescriptor<net.stef.grpc.service.destination.STEFClientMessage, net.stef.grpc.service.destination.STEFServerMessage> getStreamMethod;
    if ((getStreamMethod = STEFDestinationGrpc.getStreamMethod) == null) {
      synchronized (STEFDestinationGrpc.class) {
        if ((getStreamMethod = STEFDestinationGrpc.getStreamMethod) == null) {
          STEFDestinationGrpc.getStreamMethod = getStreamMethod =
              io.grpc.MethodDescriptor.<net.stef.grpc.service.destination.STEFClientMessage, net.stef.grpc.service.destination.STEFServerMessage>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Stream"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  net.stef.grpc.service.destination.STEFClientMessage.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  net.stef.grpc.service.destination.STEFServerMessage.getDefaultInstance()))
              .setSchemaDescriptor(new STEFDestinationMethodDescriptorSupplier("Stream"))
              .build();
        }
      }
    }
    return getStreamMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static STEFDestinationStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<STEFDestinationStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<STEFDestinationStub>() {
        @java.lang.Override
        public STEFDestinationStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new STEFDestinationStub(channel, callOptions);
        }
      };
    return STEFDestinationStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static STEFDestinationBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<STEFDestinationBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<STEFDestinationBlockingV2Stub>() {
        @java.lang.Override
        public STEFDestinationBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new STEFDestinationBlockingV2Stub(channel, callOptions);
        }
      };
    return STEFDestinationBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static STEFDestinationBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<STEFDestinationBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<STEFDestinationBlockingStub>() {
        @java.lang.Override
        public STEFDestinationBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new STEFDestinationBlockingStub(channel, callOptions);
        }
      };
    return STEFDestinationBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static STEFDestinationFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<STEFDestinationFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<STEFDestinationFutureStub>() {
        @java.lang.Override
        public STEFDestinationFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new STEFDestinationFutureStub(channel, callOptions);
        }
      };
    return STEFDestinationFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public interface AsyncService {

    /**
     * <pre>
     * Stream is a channel to send STEF data from the client to Destination.
     * Once the stream is open the Destination MUST send a ServerMessage with
     * DestCapabilities field set.
     * The client MUST examine received DestCapabilities and if the client is
     * able to operate as requested by DestCapabilities the client MUST begin sending
     * ClientMessage containing STEF data. The Destination MUST periodically
     * respond with ServerMessage containing ExportResponse field.
     * One gRPC stream corresponds to one STEF byte stream.
     * </pre>
     */
    default io.grpc.stub.StreamObserver<net.stef.grpc.service.destination.STEFClientMessage> stream(
        io.grpc.stub.StreamObserver<net.stef.grpc.service.destination.STEFServerMessage> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getStreamMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service STEFDestination.
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public static abstract class STEFDestinationImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return STEFDestinationGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service STEFDestination.
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public static final class STEFDestinationStub
      extends io.grpc.stub.AbstractAsyncStub<STEFDestinationStub> {
    private STEFDestinationStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected STEFDestinationStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new STEFDestinationStub(channel, callOptions);
    }

    /**
     * <pre>
     * Stream is a channel to send STEF data from the client to Destination.
     * Once the stream is open the Destination MUST send a ServerMessage with
     * DestCapabilities field set.
     * The client MUST examine received DestCapabilities and if the client is
     * able to operate as requested by DestCapabilities the client MUST begin sending
     * ClientMessage containing STEF data. The Destination MUST periodically
     * respond with ServerMessage containing ExportResponse field.
     * One gRPC stream corresponds to one STEF byte stream.
     * </pre>
     */
    public io.grpc.stub.StreamObserver<net.stef.grpc.service.destination.STEFClientMessage> stream(
        io.grpc.stub.StreamObserver<net.stef.grpc.service.destination.STEFServerMessage> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getStreamMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service STEFDestination.
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public static final class STEFDestinationBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<STEFDestinationBlockingV2Stub> {
    private STEFDestinationBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected STEFDestinationBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new STEFDestinationBlockingV2Stub(channel, callOptions);
    }

    /**
     * <pre>
     * Stream is a channel to send STEF data from the client to Destination.
     * Once the stream is open the Destination MUST send a ServerMessage with
     * DestCapabilities field set.
     * The client MUST examine received DestCapabilities and if the client is
     * able to operate as requested by DestCapabilities the client MUST begin sending
     * ClientMessage containing STEF data. The Destination MUST periodically
     * respond with ServerMessage containing ExportResponse field.
     * One gRPC stream corresponds to one STEF byte stream.
     * </pre>
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<net.stef.grpc.service.destination.STEFClientMessage, net.stef.grpc.service.destination.STEFServerMessage>
        stream() {
      return io.grpc.stub.ClientCalls.blockingBidiStreamingCall(
          getChannel(), getStreamMethod(), getCallOptions());
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service STEFDestination.
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public static final class STEFDestinationBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<STEFDestinationBlockingStub> {
    private STEFDestinationBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected STEFDestinationBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new STEFDestinationBlockingStub(channel, callOptions);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service STEFDestination.
   * <pre>
   * Destination is a service to which STEF data can be sent.
   * </pre>
   */
  public static final class STEFDestinationFutureStub
      extends io.grpc.stub.AbstractFutureStub<STEFDestinationFutureStub> {
    private STEFDestinationFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected STEFDestinationFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new STEFDestinationFutureStub(channel, callOptions);
    }
  }

  private static final int METHODID_STREAM = 0;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_STREAM:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.stream(
              (io.grpc.stub.StreamObserver<net.stef.grpc.service.destination.STEFServerMessage>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getStreamMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              net.stef.grpc.service.destination.STEFClientMessage,
              net.stef.grpc.service.destination.STEFServerMessage>(
                service, METHODID_STREAM)))
        .build();
  }

  private static abstract class STEFDestinationBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    STEFDestinationBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return net.stef.grpc.service.destination.Destination.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("STEFDestination");
    }
  }

  private static final class STEFDestinationFileDescriptorSupplier
      extends STEFDestinationBaseDescriptorSupplier {
    STEFDestinationFileDescriptorSupplier() {}
  }

  private static final class STEFDestinationMethodDescriptorSupplier
      extends STEFDestinationBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    STEFDestinationMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (STEFDestinationGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new STEFDestinationFileDescriptorSupplier())
              .addMethod(getStreamMethod())
              .build();
        }
      }
    }
    return result;
  }
}
