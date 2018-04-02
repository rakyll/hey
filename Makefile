PROJ=hey
VERSION?=$(shell ./scripts/git-version.sh)
TIME=$(shell date "+%F_%T")

LD_FLAGS="-w -X $(PROJ)/version.BuildName=$(PROJ) \
			 -X $(PROJ)/version.BuildVersion=$(VERSION) \
			 -X $(PROJ)/version.BuildTime=$(TIME) \
			-linkmode external -extldflags '-static'"

# create some temporary folder
$(shell mkdir -p _output/bin)

.PHONY: build
build: clean
	@echo start compiling
	@CGO_ENABLED=0 go build -o _output/bin/$(PROJ) --ldflags $(LD_FLAGS) .
	@echo complete the compilation

.PHONY: clean
clean:
	@echo start cleaning
	@rm -rf _output/bin
	@echo complete the cleanup