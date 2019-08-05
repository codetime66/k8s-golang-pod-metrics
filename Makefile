# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)
# Setup name variables for the package/tool
NAME := k8s-golang-pod-metrics
PKG := github.com/codetime66/$(NAME)
# Set our default go compiler
GO := go
#
VERSION := $(shell cat VERSION.txt)

.PHONY: build
build: $(NAME) 

$(NAME): $(wildcard *.go) $(wildcard */*.go) VERSION.txt
	@echo "+ $@"
	$(GO) build -o ./bin/k8s-prom-pods ./cmd/k8s-prom-pods
