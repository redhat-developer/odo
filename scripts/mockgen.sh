#!/usr/bin/env bash

# Use this script to regenerate generated mock files
# after changing signatures of interfaces in these packages

gomockVersion=$(go list -mod=readonly -m -f '{{.Version}}' github.com/golang/mock)
mockgen="go run -mod=readonly github.com/golang/mock/mockgen@$gomockVersion"

echo Using mockgen "$($mockgen --version)"

$mockgen -source=pkg/kclient/interface.go \
    -package kclient \
    -destination pkg/kclient/mock_Client.go

$mockgen -source=pkg/storage/storage.go \
    -package storage \
    -destination pkg/storage/mock_Client.go 

$mockgen -source=pkg/devfile/image/image.go \
    -package image \
    -destination pkg/devfile/image/mock_Backend.go

$mockgen -source=pkg/odo/cmdline/cmdline.go \
    -package cmdline \
    -destination pkg/odo/cmdline/mock.go

$mockgen -source=pkg/project/project.go \
    -package project \
    -destination pkg/project/mock.go

$mockgen -source=pkg/preference/preference.go \
    -package preference \
    -destination pkg/preference/mock.go

$mockgen -source=pkg/auth/interface.go \
    -package auth \
    -destination pkg/auth/mock.go

$mockgen -source=pkg/init/backend/interface.go \
    -package backend \
    -destination pkg/init/backend/mock.go

$mockgen -source=pkg/init/asker/interface.go \
    -package asker \
    -destination pkg/init/asker/mock.go

$mockgen -source=pkg/init/interface.go \
    -package init \
    -destination pkg/init/mock.go

$mockgen -source=pkg/registry/interface.go \
    -package registry \
    -destination pkg/registry/mock.go

$mockgen -source=pkg/deploy/interface.go \
    -package deploy \
    -destination pkg/deploy/mock.go

$mockgen -source=pkg/libdevfile/libdevfile.go \
    -package libdevfile \
    -destination pkg/libdevfile/handler_mock.go

$mockgen -source=pkg/watch/interface.go \
    -package watch \
    -destination pkg/watch/mock.go

$mockgen -source=pkg/component/delete/interface.go \
    -package delete \
    -destination pkg/component/delete/mock.go

$mockgen -source=pkg/dev/interface.go \
    -package dev \
    -destination pkg/dev/mock.go

$mockgen -source=pkg/alizer/interface.go \
    -package alizer \
    -destination pkg/alizer/mock.go

$mockgen -source=pkg/binding/interface.go \
    -package binding \
    -destination pkg/binding/mock.go

$mockgen -source=pkg/binding/backend/interface.go \
    -package backend \
    -destination pkg/binding/backend/mock.go

$mockgen -source=pkg/sync/interface.go \
    -package sync \
    -destination pkg/sync/mock.go

$mockgen -source=pkg/exec/interface.go \
    -package exec \
    -destination pkg/exec/mock.go

$mockgen -source=pkg/podman/interface.go \
    -package podman \
    -destination pkg/podman/mock.go

$mockgen -source=pkg/configAutomount/interface.go \
    -package configAutomount \
    -destination pkg/configAutomount/mock.go

$mockgen -source=pkg/platform/interface.go \
    -package platform \
    -destination pkg/platform/mock.go
