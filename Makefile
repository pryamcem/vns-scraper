NAME = vns-scraper
build:
	@go build -o bin/$(NAME)

run: build
	@./bin/$(NAME)

test: 
	@go test -v ./...
