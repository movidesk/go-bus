.PHONY: default
default: help

## vet: vets all files
.PHONY: vet
vet:
	@go vet -v .\...

## test: runs all tests
.PHONY: test
test: 
	@go test -v .\...

## test-unit: runs only unit tests
.PHONY: test-unit
test-unit:
	@go test -v -run Unit .\...

## test-integration: runs only integration tests
.PHONY: test-integration
test-integration:
	@go test -v -run Integration .\...

## help: show this help
.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
