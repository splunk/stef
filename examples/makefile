SUBDIRS := $(filter-out makefile,$(patsubst %/,%,$(wildcard */)))

.PHONY: default
default: build

.PHONY: all build $(SUBDIRS)

all build:
	@for dir in $(SUBDIRS); do \
  	  echo "Building in $$dir"; \
	  $(MAKE) -C $$dir $@; \
	done

$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)
