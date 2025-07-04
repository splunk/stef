// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: destination.proto
// Protobuf Java Version: 4.29.3

package net.stef.grpc.service.destination;

/**
 * Protobuf type {@code STEFServerMessage}
 */
public final class STEFServerMessage extends
    com.google.protobuf.GeneratedMessage implements
    // @@protoc_insertion_point(message_implements:STEFServerMessage)
    STEFServerMessageOrBuilder {
private static final long serialVersionUID = 0L;
  static {
    com.google.protobuf.RuntimeVersion.validateProtobufGencodeVersion(
      com.google.protobuf.RuntimeVersion.RuntimeDomain.PUBLIC,
      /* major= */ 4,
      /* minor= */ 29,
      /* patch= */ 3,
      /* suffix= */ "",
      STEFServerMessage.class.getName());
  }
  // Use STEFServerMessage.newBuilder() to construct.
  private STEFServerMessage(com.google.protobuf.GeneratedMessage.Builder<?> builder) {
    super(builder);
  }
  private STEFServerMessage() {
  }

  public static final com.google.protobuf.Descriptors.Descriptor
      getDescriptor() {
    return net.stef.grpc.service.destination.Destination.internal_static_STEFServerMessage_descriptor;
  }

  @java.lang.Override
  protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return net.stef.grpc.service.destination.Destination.internal_static_STEFServerMessage_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            net.stef.grpc.service.destination.STEFServerMessage.class, net.stef.grpc.service.destination.STEFServerMessage.Builder.class);
  }

  private int messageCase_ = 0;
  @SuppressWarnings("serial")
  private java.lang.Object message_;
  public enum MessageCase
      implements com.google.protobuf.Internal.EnumLite,
          com.google.protobuf.AbstractMessage.InternalOneOfEnum {
    CAPABILITIES(1),
    RESPONSE(2),
    MESSAGE_NOT_SET(0);
    private final int value;
    private MessageCase(int value) {
      this.value = value;
    }
    /**
     * @param value The number of the enum to look for.
     * @return The enum associated with the given number.
     * @deprecated Use {@link #forNumber(int)} instead.
     */
    @java.lang.Deprecated
    public static MessageCase valueOf(int value) {
      return forNumber(value);
    }

    public static MessageCase forNumber(int value) {
      switch (value) {
        case 1: return CAPABILITIES;
        case 2: return RESPONSE;
        case 0: return MESSAGE_NOT_SET;
        default: return null;
      }
    }
    public int getNumber() {
      return this.value;
    }
  };

  public MessageCase
  getMessageCase() {
    return MessageCase.forNumber(
        messageCase_);
  }

  public static final int CAPABILITIES_FIELD_NUMBER = 1;
  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   * @return Whether the capabilities field is set.
   */
  @java.lang.Override
  public boolean hasCapabilities() {
    return messageCase_ == 1;
  }
  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   * @return The capabilities.
   */
  @java.lang.Override
  public net.stef.grpc.service.destination.STEFDestinationCapabilities getCapabilities() {
    if (messageCase_ == 1) {
       return (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_;
    }
    return net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
  }
  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   */
  @java.lang.Override
  public net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder getCapabilitiesOrBuilder() {
    if (messageCase_ == 1) {
       return (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_;
    }
    return net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
  }

  public static final int RESPONSE_FIELD_NUMBER = 2;
  /**
   * <code>.STEFDataResponse response = 2;</code>
   * @return Whether the response field is set.
   */
  @java.lang.Override
  public boolean hasResponse() {
    return messageCase_ == 2;
  }
  /**
   * <code>.STEFDataResponse response = 2;</code>
   * @return The response.
   */
  @java.lang.Override
  public net.stef.grpc.service.destination.STEFDataResponse getResponse() {
    if (messageCase_ == 2) {
       return (net.stef.grpc.service.destination.STEFDataResponse) message_;
    }
    return net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
  }
  /**
   * <code>.STEFDataResponse response = 2;</code>
   */
  @java.lang.Override
  public net.stef.grpc.service.destination.STEFDataResponseOrBuilder getResponseOrBuilder() {
    if (messageCase_ == 2) {
       return (net.stef.grpc.service.destination.STEFDataResponse) message_;
    }
    return net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
  }

  private byte memoizedIsInitialized = -1;
  @java.lang.Override
  public final boolean isInitialized() {
    byte isInitialized = memoizedIsInitialized;
    if (isInitialized == 1) return true;
    if (isInitialized == 0) return false;

    memoizedIsInitialized = 1;
    return true;
  }

  @java.lang.Override
  public void writeTo(com.google.protobuf.CodedOutputStream output)
                      throws java.io.IOException {
    if (messageCase_ == 1) {
      output.writeMessage(1, (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_);
    }
    if (messageCase_ == 2) {
      output.writeMessage(2, (net.stef.grpc.service.destination.STEFDataResponse) message_);
    }
    getUnknownFields().writeTo(output);
  }

  @java.lang.Override
  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    if (messageCase_ == 1) {
      size += com.google.protobuf.CodedOutputStream
        .computeMessageSize(1, (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_);
    }
    if (messageCase_ == 2) {
      size += com.google.protobuf.CodedOutputStream
        .computeMessageSize(2, (net.stef.grpc.service.destination.STEFDataResponse) message_);
    }
    size += getUnknownFields().getSerializedSize();
    memoizedSize = size;
    return size;
  }

  @java.lang.Override
  public boolean equals(final java.lang.Object obj) {
    if (obj == this) {
     return true;
    }
    if (!(obj instanceof net.stef.grpc.service.destination.STEFServerMessage)) {
      return super.equals(obj);
    }
    net.stef.grpc.service.destination.STEFServerMessage other = (net.stef.grpc.service.destination.STEFServerMessage) obj;

    if (!getMessageCase().equals(other.getMessageCase())) return false;
    switch (messageCase_) {
      case 1:
        if (!getCapabilities()
            .equals(other.getCapabilities())) return false;
        break;
      case 2:
        if (!getResponse()
            .equals(other.getResponse())) return false;
        break;
      case 0:
      default:
    }
    if (!getUnknownFields().equals(other.getUnknownFields())) return false;
    return true;
  }

  @java.lang.Override
  public int hashCode() {
    if (memoizedHashCode != 0) {
      return memoizedHashCode;
    }
    int hash = 41;
    hash = (19 * hash) + getDescriptor().hashCode();
    switch (messageCase_) {
      case 1:
        hash = (37 * hash) + CAPABILITIES_FIELD_NUMBER;
        hash = (53 * hash) + getCapabilities().hashCode();
        break;
      case 2:
        hash = (37 * hash) + RESPONSE_FIELD_NUMBER;
        hash = (53 * hash) + getResponse().hashCode();
        break;
      case 0:
      default:
    }
    hash = (29 * hash) + getUnknownFields().hashCode();
    memoizedHashCode = hash;
    return hash;
  }

  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      java.nio.ByteBuffer data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      java.nio.ByteBuffer data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  public static net.stef.grpc.service.destination.STEFServerMessage parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input);
  }

  public static net.stef.grpc.service.destination.STEFServerMessage parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static net.stef.grpc.service.destination.STEFServerMessage parseFrom(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  @java.lang.Override
  public Builder newBuilderForType() { return newBuilder(); }
  public static Builder newBuilder() {
    return DEFAULT_INSTANCE.toBuilder();
  }
  public static Builder newBuilder(net.stef.grpc.service.destination.STEFServerMessage prototype) {
    return DEFAULT_INSTANCE.toBuilder().mergeFrom(prototype);
  }
  @java.lang.Override
  public Builder toBuilder() {
    return this == DEFAULT_INSTANCE
        ? new Builder() : new Builder().mergeFrom(this);
  }

  @java.lang.Override
  protected Builder newBuilderForType(
      com.google.protobuf.GeneratedMessage.BuilderParent parent) {
    Builder builder = new Builder(parent);
    return builder;
  }
  /**
   * Protobuf type {@code STEFServerMessage}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessage.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:STEFServerMessage)
      net.stef.grpc.service.destination.STEFServerMessageOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return net.stef.grpc.service.destination.Destination.internal_static_STEFServerMessage_descriptor;
    }

    @java.lang.Override
    protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return net.stef.grpc.service.destination.Destination.internal_static_STEFServerMessage_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              net.stef.grpc.service.destination.STEFServerMessage.class, net.stef.grpc.service.destination.STEFServerMessage.Builder.class);
    }

    // Construct using net.stef.grpc.service.destination.STEFServerMessage.newBuilder()
    private Builder() {

    }

    private Builder(
        com.google.protobuf.GeneratedMessage.BuilderParent parent) {
      super(parent);

    }
    @java.lang.Override
    public Builder clear() {
      super.clear();
      bitField0_ = 0;
      if (capabilitiesBuilder_ != null) {
        capabilitiesBuilder_.clear();
      }
      if (responseBuilder_ != null) {
        responseBuilder_.clear();
      }
      messageCase_ = 0;
      message_ = null;
      return this;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return net.stef.grpc.service.destination.Destination.internal_static_STEFServerMessage_descriptor;
    }

    @java.lang.Override
    public net.stef.grpc.service.destination.STEFServerMessage getDefaultInstanceForType() {
      return net.stef.grpc.service.destination.STEFServerMessage.getDefaultInstance();
    }

    @java.lang.Override
    public net.stef.grpc.service.destination.STEFServerMessage build() {
      net.stef.grpc.service.destination.STEFServerMessage result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    @java.lang.Override
    public net.stef.grpc.service.destination.STEFServerMessage buildPartial() {
      net.stef.grpc.service.destination.STEFServerMessage result = new net.stef.grpc.service.destination.STEFServerMessage(this);
      if (bitField0_ != 0) { buildPartial0(result); }
      buildPartialOneofs(result);
      onBuilt();
      return result;
    }

    private void buildPartial0(net.stef.grpc.service.destination.STEFServerMessage result) {
      int from_bitField0_ = bitField0_;
    }

    private void buildPartialOneofs(net.stef.grpc.service.destination.STEFServerMessage result) {
      result.messageCase_ = messageCase_;
      result.message_ = this.message_;
      if (messageCase_ == 1 &&
          capabilitiesBuilder_ != null) {
        result.message_ = capabilitiesBuilder_.build();
      }
      if (messageCase_ == 2 &&
          responseBuilder_ != null) {
        result.message_ = responseBuilder_.build();
      }
    }

    @java.lang.Override
    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof net.stef.grpc.service.destination.STEFServerMessage) {
        return mergeFrom((net.stef.grpc.service.destination.STEFServerMessage)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(net.stef.grpc.service.destination.STEFServerMessage other) {
      if (other == net.stef.grpc.service.destination.STEFServerMessage.getDefaultInstance()) return this;
      switch (other.getMessageCase()) {
        case CAPABILITIES: {
          mergeCapabilities(other.getCapabilities());
          break;
        }
        case RESPONSE: {
          mergeResponse(other.getResponse());
          break;
        }
        case MESSAGE_NOT_SET: {
          break;
        }
      }
      this.mergeUnknownFields(other.getUnknownFields());
      onChanged();
      return this;
    }

    @java.lang.Override
    public final boolean isInitialized() {
      return true;
    }

    @java.lang.Override
    public Builder mergeFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws java.io.IOException {
      if (extensionRegistry == null) {
        throw new java.lang.NullPointerException();
      }
      try {
        boolean done = false;
        while (!done) {
          int tag = input.readTag();
          switch (tag) {
            case 0:
              done = true;
              break;
            case 10: {
              input.readMessage(
                  getCapabilitiesFieldBuilder().getBuilder(),
                  extensionRegistry);
              messageCase_ = 1;
              break;
            } // case 10
            case 18: {
              input.readMessage(
                  getResponseFieldBuilder().getBuilder(),
                  extensionRegistry);
              messageCase_ = 2;
              break;
            } // case 18
            default: {
              if (!super.parseUnknownField(input, extensionRegistry, tag)) {
                done = true; // was an endgroup tag
              }
              break;
            } // default:
          } // switch (tag)
        } // while (!done)
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.unwrapIOException();
      } finally {
        onChanged();
      } // finally
      return this;
    }
    private int messageCase_ = 0;
    private java.lang.Object message_;
    public MessageCase
        getMessageCase() {
      return MessageCase.forNumber(
          messageCase_);
    }

    public Builder clearMessage() {
      messageCase_ = 0;
      message_ = null;
      onChanged();
      return this;
    }

    private int bitField0_;

    private com.google.protobuf.SingleFieldBuilder<
        net.stef.grpc.service.destination.STEFDestinationCapabilities, net.stef.grpc.service.destination.STEFDestinationCapabilities.Builder, net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder> capabilitiesBuilder_;
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     * @return Whether the capabilities field is set.
     */
    @java.lang.Override
    public boolean hasCapabilities() {
      return messageCase_ == 1;
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     * @return The capabilities.
     */
    @java.lang.Override
    public net.stef.grpc.service.destination.STEFDestinationCapabilities getCapabilities() {
      if (capabilitiesBuilder_ == null) {
        if (messageCase_ == 1) {
          return (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_;
        }
        return net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
      } else {
        if (messageCase_ == 1) {
          return capabilitiesBuilder_.getMessage();
        }
        return net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
      }
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    public Builder setCapabilities(net.stef.grpc.service.destination.STEFDestinationCapabilities value) {
      if (capabilitiesBuilder_ == null) {
        if (value == null) {
          throw new NullPointerException();
        }
        message_ = value;
        onChanged();
      } else {
        capabilitiesBuilder_.setMessage(value);
      }
      messageCase_ = 1;
      return this;
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    public Builder setCapabilities(
        net.stef.grpc.service.destination.STEFDestinationCapabilities.Builder builderForValue) {
      if (capabilitiesBuilder_ == null) {
        message_ = builderForValue.build();
        onChanged();
      } else {
        capabilitiesBuilder_.setMessage(builderForValue.build());
      }
      messageCase_ = 1;
      return this;
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    public Builder mergeCapabilities(net.stef.grpc.service.destination.STEFDestinationCapabilities value) {
      if (capabilitiesBuilder_ == null) {
        if (messageCase_ == 1 &&
            message_ != net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance()) {
          message_ = net.stef.grpc.service.destination.STEFDestinationCapabilities.newBuilder((net.stef.grpc.service.destination.STEFDestinationCapabilities) message_)
              .mergeFrom(value).buildPartial();
        } else {
          message_ = value;
        }
        onChanged();
      } else {
        if (messageCase_ == 1) {
          capabilitiesBuilder_.mergeFrom(value);
        } else {
          capabilitiesBuilder_.setMessage(value);
        }
      }
      messageCase_ = 1;
      return this;
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    public Builder clearCapabilities() {
      if (capabilitiesBuilder_ == null) {
        if (messageCase_ == 1) {
          messageCase_ = 0;
          message_ = null;
          onChanged();
        }
      } else {
        if (messageCase_ == 1) {
          messageCase_ = 0;
          message_ = null;
        }
        capabilitiesBuilder_.clear();
      }
      return this;
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    public net.stef.grpc.service.destination.STEFDestinationCapabilities.Builder getCapabilitiesBuilder() {
      return getCapabilitiesFieldBuilder().getBuilder();
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    @java.lang.Override
    public net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder getCapabilitiesOrBuilder() {
      if ((messageCase_ == 1) && (capabilitiesBuilder_ != null)) {
        return capabilitiesBuilder_.getMessageOrBuilder();
      } else {
        if (messageCase_ == 1) {
          return (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_;
        }
        return net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
      }
    }
    /**
     * <code>.STEFDestinationCapabilities capabilities = 1;</code>
     */
    private com.google.protobuf.SingleFieldBuilder<
        net.stef.grpc.service.destination.STEFDestinationCapabilities, net.stef.grpc.service.destination.STEFDestinationCapabilities.Builder, net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder> 
        getCapabilitiesFieldBuilder() {
      if (capabilitiesBuilder_ == null) {
        if (!(messageCase_ == 1)) {
          message_ = net.stef.grpc.service.destination.STEFDestinationCapabilities.getDefaultInstance();
        }
        capabilitiesBuilder_ = new com.google.protobuf.SingleFieldBuilder<
            net.stef.grpc.service.destination.STEFDestinationCapabilities, net.stef.grpc.service.destination.STEFDestinationCapabilities.Builder, net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder>(
                (net.stef.grpc.service.destination.STEFDestinationCapabilities) message_,
                getParentForChildren(),
                isClean());
        message_ = null;
      }
      messageCase_ = 1;
      onChanged();
      return capabilitiesBuilder_;
    }

    private com.google.protobuf.SingleFieldBuilder<
        net.stef.grpc.service.destination.STEFDataResponse, net.stef.grpc.service.destination.STEFDataResponse.Builder, net.stef.grpc.service.destination.STEFDataResponseOrBuilder> responseBuilder_;
    /**
     * <code>.STEFDataResponse response = 2;</code>
     * @return Whether the response field is set.
     */
    @java.lang.Override
    public boolean hasResponse() {
      return messageCase_ == 2;
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     * @return The response.
     */
    @java.lang.Override
    public net.stef.grpc.service.destination.STEFDataResponse getResponse() {
      if (responseBuilder_ == null) {
        if (messageCase_ == 2) {
          return (net.stef.grpc.service.destination.STEFDataResponse) message_;
        }
        return net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
      } else {
        if (messageCase_ == 2) {
          return responseBuilder_.getMessage();
        }
        return net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
      }
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    public Builder setResponse(net.stef.grpc.service.destination.STEFDataResponse value) {
      if (responseBuilder_ == null) {
        if (value == null) {
          throw new NullPointerException();
        }
        message_ = value;
        onChanged();
      } else {
        responseBuilder_.setMessage(value);
      }
      messageCase_ = 2;
      return this;
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    public Builder setResponse(
        net.stef.grpc.service.destination.STEFDataResponse.Builder builderForValue) {
      if (responseBuilder_ == null) {
        message_ = builderForValue.build();
        onChanged();
      } else {
        responseBuilder_.setMessage(builderForValue.build());
      }
      messageCase_ = 2;
      return this;
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    public Builder mergeResponse(net.stef.grpc.service.destination.STEFDataResponse value) {
      if (responseBuilder_ == null) {
        if (messageCase_ == 2 &&
            message_ != net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance()) {
          message_ = net.stef.grpc.service.destination.STEFDataResponse.newBuilder((net.stef.grpc.service.destination.STEFDataResponse) message_)
              .mergeFrom(value).buildPartial();
        } else {
          message_ = value;
        }
        onChanged();
      } else {
        if (messageCase_ == 2) {
          responseBuilder_.mergeFrom(value);
        } else {
          responseBuilder_.setMessage(value);
        }
      }
      messageCase_ = 2;
      return this;
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    public Builder clearResponse() {
      if (responseBuilder_ == null) {
        if (messageCase_ == 2) {
          messageCase_ = 0;
          message_ = null;
          onChanged();
        }
      } else {
        if (messageCase_ == 2) {
          messageCase_ = 0;
          message_ = null;
        }
        responseBuilder_.clear();
      }
      return this;
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    public net.stef.grpc.service.destination.STEFDataResponse.Builder getResponseBuilder() {
      return getResponseFieldBuilder().getBuilder();
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    @java.lang.Override
    public net.stef.grpc.service.destination.STEFDataResponseOrBuilder getResponseOrBuilder() {
      if ((messageCase_ == 2) && (responseBuilder_ != null)) {
        return responseBuilder_.getMessageOrBuilder();
      } else {
        if (messageCase_ == 2) {
          return (net.stef.grpc.service.destination.STEFDataResponse) message_;
        }
        return net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
      }
    }
    /**
     * <code>.STEFDataResponse response = 2;</code>
     */
    private com.google.protobuf.SingleFieldBuilder<
        net.stef.grpc.service.destination.STEFDataResponse, net.stef.grpc.service.destination.STEFDataResponse.Builder, net.stef.grpc.service.destination.STEFDataResponseOrBuilder> 
        getResponseFieldBuilder() {
      if (responseBuilder_ == null) {
        if (!(messageCase_ == 2)) {
          message_ = net.stef.grpc.service.destination.STEFDataResponse.getDefaultInstance();
        }
        responseBuilder_ = new com.google.protobuf.SingleFieldBuilder<
            net.stef.grpc.service.destination.STEFDataResponse, net.stef.grpc.service.destination.STEFDataResponse.Builder, net.stef.grpc.service.destination.STEFDataResponseOrBuilder>(
                (net.stef.grpc.service.destination.STEFDataResponse) message_,
                getParentForChildren(),
                isClean());
        message_ = null;
      }
      messageCase_ = 2;
      onChanged();
      return responseBuilder_;
    }

    // @@protoc_insertion_point(builder_scope:STEFServerMessage)
  }

  // @@protoc_insertion_point(class_scope:STEFServerMessage)
  private static final net.stef.grpc.service.destination.STEFServerMessage DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new net.stef.grpc.service.destination.STEFServerMessage();
  }

  public static net.stef.grpc.service.destination.STEFServerMessage getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<STEFServerMessage>
      PARSER = new com.google.protobuf.AbstractParser<STEFServerMessage>() {
    @java.lang.Override
    public STEFServerMessage parsePartialFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws com.google.protobuf.InvalidProtocolBufferException {
      Builder builder = newBuilder();
      try {
        builder.mergeFrom(input, extensionRegistry);
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.setUnfinishedMessage(builder.buildPartial());
      } catch (com.google.protobuf.UninitializedMessageException e) {
        throw e.asInvalidProtocolBufferException().setUnfinishedMessage(builder.buildPartial());
      } catch (java.io.IOException e) {
        throw new com.google.protobuf.InvalidProtocolBufferException(e)
            .setUnfinishedMessage(builder.buildPartial());
      }
      return builder.buildPartial();
    }
  };

  public static com.google.protobuf.Parser<STEFServerMessage> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<STEFServerMessage> getParserForType() {
    return PARSER;
  }

  @java.lang.Override
  public net.stef.grpc.service.destination.STEFServerMessage getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

