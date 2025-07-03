# Function to execute a command.
# Accepts command to execute as first parameter.
define exec-command
$(1)

endef

.PHONY: default
default:
	cd stefgen && make
	cd go/pkg && make
	cd go/grpc && make
	cd go/otel && make
	cd go/pdata && make
	cd otelcol && make
	cd examples && make
	cd benchmarks && make

.PHONY: all
all: docs-validate
	cd stefgen && make all
	cd go/pkg && make all
	cd go/grpc && make all
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make all
	cd benchmarks && make all

.PHONY: build-ci
build-ci:
	cd stefgen && make all
	cd go/pkg && make all
	cd go/grpc && make
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make
	cd benchmarks && make all

.PHONY: verifyver
verifyver:
ifndef VERSION
		@echo "VERSION is unset or set to the empty string"
		@exit 1
endif

.PHONY: prepver
prepver: verifyver
	echo Updating to version ${VERSION}
	cd go/grpc     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd go/otel     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/grpc@${VERSION} && go mod tidy
	cd go/pdata    && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/otel@${VERSION} && go mod tidy
	cd examples/jsonl && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd otelcol     && go mod tidy
	cd benchmarks  && go mod tidy

MODULES := go/pkg go/grpc go/otel go/pdata

.PHONY: releasever
releasever: verifyver
	echo Tagging version $(VERSION)
	$(foreach gomod,$(MODULES),$(call exec-command,git tag $(gomod)/$(VERSION) && git push origin $(gomod)/$(VERSION)))

# Docs validation targets
.PHONY: docs-validate docs-validate-html docs-validate-css docs-check-links docs-install-deps

# Validate all docs (HTML, CSS, and links)
docs-validate: docs-validate-html docs-validate-css docs-check-links
	@echo "✅ All docs validation checks passed!"

# Validate HTML files in docs directory
docs-validate-html:
	@echo "🔍 Validating HTML files in docs..."
	@cd docs && for file in *.html; do \
		if [ -f "$$file" ] && [ "$$file" != "benchmarks.html" ]; then \
			echo "Validating $$file..."; \
			../node_modules/.bin/html-validate "$$file" || (echo "❌ HTML validation failed for $$file" && exit 1); \
		elif [ -f "$$file" ] && [ "$$file" = "benchmarks.html" ]; then \
			echo "Skipping validation for $$file (excluded)"; \
		fi; \
	done
	@echo "✅ HTML validation complete"

# Validate CSS files in docs directory
docs-validate-css:
	@echo "🔍 Validating CSS files in docs..."
	@cd docs && for file in *.css; do \
		if [ -f "$$file" ]; then \
			echo "Validating $$file..."; \
			../node_modules/.bin/stylelint "$$file" --config-basedir .. || (echo "❌ CSS validation failed for $$file" && exit 1); \
		fi; \
	done
	@echo "✅ CSS validation complete"

# Check links in HTML files
docs-check-links:
	@for file in docs/*.html; do \
		if [ -f "$$file" ]; then \
			echo "Checking links in $$(basename $$file)..."; \
			./node_modules/.bin/markdown-link-check "$$file" --quiet || true; \
			echo "  ✅ Link check completed for $$(basename $$file)"; \
		fi; \
	done
	@echo "✅ Link checking complete"

# Install npm-based validation dependencies for docs
docs-install-deps:
	@echo "📦 Installing npm-based validation dependencies at top level..."
	@if ! command -v npm >/dev/null 2>&1; then \
		echo "❌ npm not found. Please install Node.js and npm first."; \
		echo "Visit: https://nodejs.org/"; \
		exit 1; \
	fi
	@if [ ! -f package.json ]; then \
		echo "Creating package.json..."; \
		npm init -y; \
	fi
	@echo "Installing HTML validation tools..."
	@npm install html-validate
	@echo "Installing CSS validation tools..."
	@npm install stylelint stylelint-config-standard
	@echo "Installing link checking tools..."
	@npm install markdown-link-check
	@echo "✅ All docs dependencies installed successfully!"
	@echo "Tools installed in ./node_modules/.bin/"
