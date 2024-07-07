.PHONY: test
test:
	go test -cover ./...

.PHONY: test.build
test.build:
	mkdir -p test.out/
	for f in $(shell go list ./...); do \
		go test -race -c -o test.out/`echo "$$f" | sha256sum | cut -c1-5`-`basename "$$f"` $$f ; \
	done

.PHONY: test.run
test.run:
	for f in test.out/*; do \
		$$f ; \
	done
