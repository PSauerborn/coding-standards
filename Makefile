.PHONY: scan-secrets
scan-secrets:
	@echo "Scanning for secrets..."

	@detect-secrets scan > .secrets.baseline
	@detect-secrets audit .secrets.baseline

	@echo "Scan complete. Results saved to .secrets.baseline"

.PHONY: claude
claude:
	@echo "Initiating claude session..."
	@claude \
		--plugin-dir ~/Github/psauerborn/agents
