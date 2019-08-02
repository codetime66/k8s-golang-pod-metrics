NAME := k8s-prom-pods
GO := go
VERSION := $(shell cat VERSION.txt)

.PHONY: build
build: $(NAME) 

$(NAME): $(wildcard *.go) $(wildcard */*.go) VERSION.txt
	@echo "+ $@"
	$(GO) build -o ./bin/k8s-prom-pods ./cmd/k8s-prom-pods
