build:
	@echo VERSION is: $(shell git describe --tags)
	@go build -ldflags "-X main.version=$(shell git describe --tags)" -o scrap-chat cmd/scrap-chat/main.go

example-live:
	@go build examples/get_live_chat/get_live_chat.go
	@./get_live_chat "https://www.youtube.com/watch?v=Qhfd3zaUeco"