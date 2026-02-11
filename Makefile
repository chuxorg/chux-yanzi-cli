.PHONY: build run sanity


build:
	@mkdir -p bin
	go build -o bin/yanzi ./cmd/yanzi

run:
	go run ./cmd/yanzi $(ARGS)

sanity:
   
   go run ./yanzi capture --prompt-file=foo.txt --response-file=bar.txt ...