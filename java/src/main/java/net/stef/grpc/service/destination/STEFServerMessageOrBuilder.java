// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: destination.proto
// Protobuf Java Version: 4.29.3

package net.stef.grpc.service.destination;

public interface STEFServerMessageOrBuilder extends
    // @@protoc_insertion_point(interface_extends:STEFServerMessage)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   * @return Whether the capabilities field is set.
   */
  boolean hasCapabilities();
  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   * @return The capabilities.
   */
  net.stef.grpc.service.destination.STEFDestinationCapabilities getCapabilities();
  /**
   * <code>.STEFDestinationCapabilities capabilities = 1;</code>
   */
  net.stef.grpc.service.destination.STEFDestinationCapabilitiesOrBuilder getCapabilitiesOrBuilder();

  /**
   * <code>.STEFDataResponse response = 2;</code>
   * @return Whether the response field is set.
   */
  boolean hasResponse();
  /**
   * <code>.STEFDataResponse response = 2;</code>
   * @return The response.
   */
  net.stef.grpc.service.destination.STEFDataResponse getResponse();
  /**
   * <code>.STEFDataResponse response = 2;</code>
   */
  net.stef.grpc.service.destination.STEFDataResponseOrBuilder getResponseOrBuilder();

  net.stef.grpc.service.destination.STEFServerMessage.MessageCase getMessageCase();
}
