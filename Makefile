BINARIES = $(subst cmd/,,$(wildcard cmd/*))

all:
	@for target in $(BINARIES); do \
		echo Building $$target; \
		go build -o bin/$$target ./cmd/$$target; \
	done
