# Other config
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: testacc pre-commit

default: testacc

# Run acceptance tests
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Development tools
# Currently just installs pre-commit hooks
setup-dev: pre-commit

pre-commit:
	@echo "$(OK_COLOR)==> Installing pre-commit hooks$(NO_COLOR)"
	@pre-commit install
	@echo "$(OK_COLOR)==> Installing custom commit message hook$(NO_COLOR)"
	@cp .githooks/prepare-commit-msg .git/hooks/prepare-commit-msg
	@chmod +x .git/hooks/prepare-commit-msg
	@echo "$(OK_COLOR)==> Hooks installed$(NO_COLOR)"

update-modules:
	@echo "$(OK_COLOR)==> Updating modules$(NO_COLOR)"
	@go get -u ./...
	@go mod tidy
	@echo "$(OK_COLOR)==> Modules updated$(NO_COLOR)"

generate:
	@echo "$(OK_COLOR)==> Generating code and docs$(NO_COLOR)"
	@go generate ./...
	@echo "$(OK_COLOR)==> Generation complete$(NO_COLOR)"
