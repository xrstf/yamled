# SPDX-FileCopyrightText: 2023 Christoph Mewes
# SPDX-License-Identifier: MIT

.PHONY: lint
lint:
	golangci-lint run --timeout=10m ./...

.PHONY: test
test:
	go test -v ./...
