NAME = vns-scruper
build:
	@go build -o bin/$(NAME)

run: build
	@./bin/$(NAME)

test: 
	@go test -v ./...
