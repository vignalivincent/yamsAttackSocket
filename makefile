# =========================================================
# 🎲 YamsAttackSocket - Build & Test Automation
# =========================================================

.PHONY: test test-coverage test-watch clean help

# =========================================================
# 📌 Variables
# =========================================================
PACKAGE_WS=./internal/websocket
PACKAGE_ALL=./...
COVERAGE_FILE=coverage.out
PACKAGE_COVERAGE_FILE=package_coverage.out
GOTESTSUM=~/go/bin/gotestsum

# Check if gotestsum is available
GOTESTSUM_CHECK := $(shell which $(GOTESTSUM) 2> /dev/null)

# =========================================================
# 📋 Help Target (Default)
# =========================================================
help:
	@echo "╔═══════════════════════════════════════════════════════════╗"
	@echo "║ 🎲 YamsAttackSocket - Available Commands                  ║"
	@echo "╠═══════════════════════════════════════════════════════════╣"
	@echo "║ • make test          Run all tests with coverage report   ║"
	@echo "║ • make test-coverage Show detailed coverage from last run ║"
	@echo "║ • make test-watch    Watch mode for auto-running tests    ║"
	@echo "║ • make clean         Remove generated test files          ║"
	@echo "╚═══════════════════════════════════════════════════════════╝"

# =========================================================
# 🧪 Test Target
# =========================================================
test:
ifdef GOTESTSUM_CHECK
	@echo "\n🧪  \033[1;34mRunning tests with coverage...\033[0m"
	@echo "════════════════════════════════════════════════════════════"
	$(GOTESTSUM) --format testdox -- -coverprofile=$(COVERAGE_FILE) $(PACKAGE_ALL)
	@go tool cover -func=$(COVERAGE_FILE) | grep total: | awk '{print "\n📊  \033[1;36mOverall Coverage:  " $$3 "\033[0m"}'
	@echo "\n📊  \033[1;35mCoverage by Package:\033[0m"
	@echo "────────────────────────────────────────────────────────────"
	@go list $(PACKAGE_ALL) | xargs -I{} sh -c 'go test -coverprofile=/tmp/pkg.out {} >/dev/null 2>&1 && echo "    \033[1;32m•\033[0m  $$(basename {})  :  \033[1;33m$$(go tool cover -func=/tmp/pkg.out | grep total: | awk "{print \$$3}")\033[0m" || echo "    \033[1;31m•\033[0m  $$(basename {})  :  \033[1;31m0.0%\033[0m"' | sort | awk '{printf "    \033[1;34m•\033[0m  %-30s  :  %s\n", $$2, $$4}'
else
	@echo "\n⚙️  \033[1;33mGotestsum not found, installing...\033[0m"
	go install gotest.tools/gotestsum@latest
	@echo "\n🧪  \033[1;34mRunning tests with coverage...\033[0m"
	@echo "════════════════════════════════════════════════════════════"
	$(GOTESTSUM) --format testdox -- -coverprofile=$(COVERAGE_FILE) $(PACKAGE_ALL)
	@go tool cover -func=$(COVERAGE_FILE) | grep total: | awk '{print "\n📊  \033[1;36mOverall Coverage:  " $$3 "\033[0m"}'
	@echo "\n📊  \033[1;35mCoverage by Package:\033[0m"
	@echo "────────────────────────────────────────────────────────────"
	@go list $(PACKAGE_ALL) | xargs -I{} sh -c 'go test -coverprofile=/tmp/pkg.out {} >/dev/null 2>&1 && echo "    \033[1;32m•\033[0m  $$(basename {})  :  \033[1;33m$$(go tool cover -func=/tmp/pkg.out | grep total: | awk "{print \$$3}")\033[0m" || echo "    \033[1;31m•\033[0m  $$(basename {})  :  \033[1;31m0.0%\033[0m"' | sort | awk '{printf "    \033[1;34m•\033[0m  %-30s  :  %s\n", $$2, $$4}'
endif

# =========================================================
# 📈 Test Coverage Target (No Test Run)
# =========================================================
test-coverage:
	@if [ ! -f $(COVERAGE_FILE) ]; then \
		echo "\n⚠️  \033[1;33mNo coverage file found. Running tests first...\033[0m"; \
		$(MAKE) test; \
	else \
		echo "\n📈  \033[1;34mGenerating detailed coverage report from existing data...\033[0m"; \
	fi
	
	@echo "════════════════════════════════════════════════════════════"
	@go tool cover -func=$(COVERAGE_FILE) | grep total: | awk '{print "\n📊  \033[1;36mOverall Coverage:  " $$3 "\033[0m"}'
	
	@echo "\n📊  \033[1;35mCoverage by Package:\033[0m"
	@echo "────────────────────────────────────────────────────────────"
	@go list $(PACKAGE_ALL) | xargs -I{} sh -c 'go test -coverprofile=/tmp/pkg.out {} >/dev/null 2>&1 && echo "    \033[1;32m•\033[0m  $$(basename {})  :  \033[1;33m$$(go tool cover -func=/tmp/pkg.out | grep total: | awk "{print \$$3}")\033[0m" || echo "    \033[1;31m•\033[0m  $$(basename {})  :  \033[1;31m0.0%\033[0m"' | sort | awk '{printf "    \033[1;34m•\033[0m  %-30s  :  %s\n", $$2, $$4}'
	
	@echo "\n📊  \033[1;35mCoverage by File:\033[0m"
	@echo "────────────────────────────────────────────────────────────"
	@go tool cover -func=$(COVERAGE_FILE) | grep -v "total:" | sort > /tmp/coverage_details.txt
	@cat /tmp/coverage_details.txt | awk '{file=$$1; gsub(/^.*\//, "", file); func=""; for(i=2; i<NF; i++) func = func (i==2 ? "" : " ") $$i; coverage=$$NF; cov_val=0+substr(coverage, 1, length(coverage)-1); if (cov_val >= 80) color="32"; else if (cov_val >= 50) color="33"; else color="31"; printf "    \033[1;34m•\033[0m  %-20s  ::  %-30s  :  \033[1;%sm%s\033[0m\n", file, func, color, coverage}' | sort

# =========================================================
# 👀 Test Watch Target
# =========================================================
test-watch:
ifdef GOTESTSUM_CHECK
	@echo "\n👀  \033[1;34mWatching for changes to run tests automatically...\033[0m"
	@echo "════════════════════════════════════════════════════════════"
	$(GOTESTSUM) --watch --format testdox -- $(PACKAGE_ALL)
else
	@echo "\n⚙️  \033[1;33mGotestsum not found, installing...\033[0m"
	go install gotest.tools/gotestsum@latest
	@echo "\n👀  \033[1;34mWatching for changes to run tests automatically...\033[0m"
	@echo "════════════════════════════════════════════════════════════"
	$(GOTESTSUM) --watch --format testdox -- $(PACKAGE_ALL)
endif

# =========================================================
# 🧹 Clean Target
# =========================================================
clean:
	@echo "\n🧹  \033[1;34mCleaning up test files...\033[0m"
	@rm -f $(COVERAGE_FILE) $(PACKAGE_COVERAGE_FILE) /tmp/pkg.out /tmp/coverage_details.txt
	@echo "✨  \033[1;32mTest files cleaned successfully\033[0m"