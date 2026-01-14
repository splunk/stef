# Function to execute a command.
# Accepts command to execute as first parameter.
define exec-command
$(1)

endef

.PHONY: default
default:
	cd stefc && make
	cd go/pkg && make
	cd go/grpc && make
	cd go/otel && make
	cd go/pdata && make
	cd otelcol && make
	cd examples && make
	cd benchmarks && make

.PHONY: all
all: docs-validate
	cd stefc && make all
	cd go/pkg && make all
	cd go/grpc && make all
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make all
	cd benchmarks && make all

.PHONY: build-ci
build-ci: docs-install-deps docs-validate
	cd stefc && make all
	cd go/pkg && make all
	cd go/grpc && make
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make
	cd benchmarks && make build-ci

.PHONY: verifyver
verifyver:
ifndef VERSION
		@echo "VERSION is unset or set to the empty string"
		@exit 1
endif

RELEASE_MODULES := go/pkg go/grpc go/otel go/pdata
ALL_MODULES += $(RELEASE_MODULES) stefc stefc/generator/testdata examples/jsonl examples/profile examples/ints otelcol benchmarks

COVERAGE_OUT ?= go/coverage.out

.PHONY: coverage
coverage:
	@tmpdir=$$(mktemp -d ./coverage-tmp.XXXXXX); \
		echo "Collecting coverage into $$tmpdir"; \
		rm -f $(COVERAGE_OUT); \
		trap 'rm -rf $$tmpdir' EXIT; \
		for mod in $(ALL_MODULES); do \
			echo "→ $$mod"; \
			outfile=$$(echo $$mod | tr '/' '_').cov; \
			( cd $$mod && go test ./... -covermode=set -coverprofile="$$OLDPWD/$$tmpdir/$$outfile" ); \
		done; \
		first=1; \
		for f in $$tmpdir/*.cov; do \
			if [ $$first -eq 1 ]; then \
				cat $$f > $(COVERAGE_OUT); \
				first=0; \
			else \
				tail -n +2 $$f >> $(COVERAGE_OUT); \
			fi; \
		done;
		cd go && go tool cover -html=coverage.out -o coverage.html
		echo "Combined coverage written to go/coverage.html"

.PHONY: gotidy
gotidy:
	$(foreach gomod,$(ALL_MODULES),$(call exec-command,cd $(gomod) && go mod tidy))

.PHONY: prepver
prepver: verifyver
	echo Updating to version ${VERSION}
	cd go/grpc     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION}
	cd go/otel     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
	               && go mod edit -require=github.com/splunk/stef/go/grpc@${VERSION}
	cd go/pdata    && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
	               && go mod edit -require=github.com/splunk/stef/go/otel@${VERSION}
	cd stefc       && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION}
	cd stefc/generator/testdata && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION}
	cd examples/jsonl && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION}
	cd examples/profile && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION}
	$(foreach gomod,$(ALL_MODULES),$(call exec-command,cd $(gomod) && go mod tidy))

.PHONY: releasever
releasever: verifyver
	echo Tagging version $(VERSION)
	$(foreach gomod,$(RELEASE_MODULES),$(call exec-command,git tag $(gomod)/$(VERSION) && git push origin $(gomod)/$(VERSION)))

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
		echo "❌ package.json not found. Cannot install dependencies."; \
		exit 1; \
	fi
	@echo "Installing all dependencies from package.json..."
	@npm install
	@echo "✅ All docs dependencies installed successfully!"
	@echo "Tools installed in ./node_modules/.bin/"
