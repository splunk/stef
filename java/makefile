# Install protoc-gen-grpc-java plugin before running this file 
# brew install protoc-gen-grpc-java

# Directories
PARENT_PATH := $(abspath $(CURDIR)/..)
PROTO_DIR := $(PARENT_PATH)/go/grpc/proto
OUT_DIR   := generated/source

# Find all .proto files
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Derive the paths to generated Java files
# Change %.proto to %.java under OUT_DIR
JAVA_FILES := $(PROTO_FILES:$(PROTO_DIR)/%.proto=$(OUT_DIR)/%.java)

# protoc command and flags
PROTOC     := protoc
GRPC_JAVA_PLUGIN := protoc-gen-grpc-java
PROTO_FLAGS := -I$(PROTO_DIR) --java_out=$(OUT_DIR)

.PHONY: all generate clean

all: generate

# Main code generation target
generate: $(JAVA_FILES)
	@echo "✅ Generated all Java sources."

# Implicit rule: how to generate each .java from .proto
$(OUT_DIR)/%.java: $(PROTO_DIR)/%.proto
	@mkdir -p $(OUT_DIR)
	$(PROTOC) --plugin=protoc-gen-grpc-java=$(GRPC_JAVA_PLUGIN) $(PROTO_FLAGS) $<
	@echo "Generated $@"

# Clean generated sources
clean:
	@rm -rf $(OUT_DIR)
	@echo "🧹 Cleaned generated code."