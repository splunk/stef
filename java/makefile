# Install protoc-gen-grpc-java plugin before running this file 
# brew install protoc-gen-grpc-java

# Directories
PARENT_PATH := $(abspath $(CURDIR)/..)
PROTO_DIR := $(PARENT_PATH)/go/grpc/proto
OUT_DIR   := $(PARENT_PATH)/java/src/main/java
CLEAN_GENERATED_CODE := $(PARENT_PATH)/java/src/main/java/net/stef/grpc

# Find all .proto files
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Derive the paths to generated Java files
# Change %.proto to %.java under OUT_DIR
JAVA_FILES := $(PROTO_FILES:$(PROTO_DIR)/%.proto=$(OUT_DIR)/%.java)

# protoc command and flags
PROTOC     := protoc

# In case of errors such as "program not found or is not executable", please specify a program using absolute path or make sure the program is available in your PATH system variable
GRPC_JAVA_PLUGIN := $(shell which protoc-gen-grpc-java)
PROTO_FLAGS := -I$(PROTO_DIR) --java_out=$(OUT_DIR) --grpc-java_out=$(OUT_DIR)

.PHONY: all generate clean

all: generate

# Main code generation target
generate: $(JAVA_FILES)
	@echo "✅ Generated all Java sources."

# Implicit rule: how to generate each .java from .proto
$(OUT_DIR)/%.java: $(PROTO_DIR)/%.proto
	$(PROTOC) --plugin=protoc-gen-grpc-java=$(GRPC_JAVA_PLUGIN) $(PROTO_FLAGS) $<
	@echo "Generated $@"

# Clean generated sources
clean:
	@rm -rf $(CLEAN_GENERATED_CODE)
	@echo "🧹 Cleaned generated code."