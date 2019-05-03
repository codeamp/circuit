include wag.mk
SHELL := /bin/bash
APP_NAME ?= atlas-api-client
PKG = github.com/Clever/$(APP_NAME)
WAG_VERSION := latest

.PHONY: generate
generate: wag-generate-deps
	$(call wag-generate,./swagger.yml,$(PKG))
	rm -rf gen-go/server
