BINARIES = $(subst cmd/,,$(wildcard cmd/*))

all: vendor
	@for target in $(BINARIES); do \
		echo Building $$target; \
		go build -o bin/$$target ./cmd/$$target; \
	done

vendor:
	dep ensure
