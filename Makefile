.DEFAULT_GOAL := help

test: ## runs 'go test'
	go test ./...

help: ## help for run
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo ""

.PHONY: test help
