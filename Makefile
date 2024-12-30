BINARY_NAME=leet2go
SRC_DIR=./cmd/scraper

.PHONY: all
all: build run

# build binary
.PHONY: build
build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

# run binary
.PHONY: run
run: build
	./$(BINARY_NAME)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

