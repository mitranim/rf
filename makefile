MAKEFLAGS := --silent --always-make
GO_FLAGS  := -tags=$(tags)
TESTFLAGS := $(if $(filter $(verb), true), -v,) -count=1 $(GO_FLAGS)
TEST      := test $(TESTFLAGS) -run=$(run)
BENCH     := test $(TESTFLAGS) -run=- -bench=$(or $(run),.) -benchmem
WATCH     := watchexec -r -c -d=0 -n

test_w:
	gow -c -v $(TEST)

test:
	go $(TEST)

bench_w:
	gow -c -v $(BENCH)

bench:
	go $(BENCH)

lint_w:
	$(WATCH) -- $(MAKE) lint

lint:
	golangci-lint run
	echo [lint] ok
