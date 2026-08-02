.PHONY: scan-secrets
scan-secrets:
	@echo "Scanning for secrets..."

	@detect-secrets scan > .secrets.baseline
	@detect-secrets audit .secrets.baseline

	@echo "Scan complete. Results saved to .secrets.baseline"

.PHONY: claude
claude:
	@echo "Initiating sandboxed claude session..."
	@docker build \
		--pull \
		--build-arg AGENTS_VERSION=0.1.0 \
		--build-arg CODING_STANDARDS_VERSION=0.1.0 \
		--build-arg CLAUDE_CODE_VERSION=latest \
		-t claude-sandbox \
		-f Dockerfile.claude \
		.

	@docker run --rm \
		-v $(PWD):/home/agent/workspace \
		-it \
		claude-sandbox

.PHONY: index
index:
	@echo "Running indexer..."
	@./bin/indexer \
		--source . \
		--output standards-tree.yaml
	@echo "Indexing complete."
