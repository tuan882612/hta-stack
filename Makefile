.PHONY: run

run:
	@templ generate
	@go run cmd/main.go internal/views/*