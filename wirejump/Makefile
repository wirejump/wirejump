# which directory to store results in
OUTPUT_DIR ?= ../build
ifeq ($(OUTPUT_DIR),)
	OUTPUT_DIR = /build
endif

# what to build
TARGETS := wirejumpd wjcli

# where target source is located
SRC_DIR := ./cmd

# format commit info
COMMIT ?= unknown
ifeq ($(COMMIT),)
	COMMIT = unknown
endif

# format version info
VERSION ?= unversioned
ifeq ($(VERSION),)
	VERSION = unversioned
endif

ifneq ($(VERSION), unversioned)
	ACTUAL = $(VERSION)
	VERSION = v.$(ACTUAL)
endif

# get build date
DATE := $(shell date -u)

# assemble final build flags
FLAGS := -s -w
FLAGS := $(FLAGS) -X 'wirejump/internal/version.BuildCommit=$(COMMIT)'
FLAGS := $(FLAGS) -X 'wirejump/internal/version.BuildDate=$(DATE)'
FLAGS := $(FLAGS) -X 'wirejump/internal/version.Version=$(VERSION)'

build:
	mkdir -p $(OUTPUT_DIR)
	$(info FLAGS)
	@echo $(FLAGS)
	@for target in $(TARGETS); do	\
		CGO_ENABLED=0 go build -v -ldflags "$(FLAGS)" -o $(OUTPUT_DIR)/$${target} $(SRC_DIR)/$${target};	\
	done
clean:
	rm -rf $(OUTPUT_DIR)

.PHONY: clean build

