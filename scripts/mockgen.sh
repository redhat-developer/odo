#!/usr/bin/env bash

# Use this script to regererate generated mock files
# after changing signatures of interfaces in these packages

mockgen -source=pkg/kclient/interface.go \
    -package kclient \
    -destination pkg/kclient/mock_Client.go

mockgen -source=pkg/localConfigProvider/localConfigProvider.go \
    -package localConfigProvider \
    -destination pkg/localConfigProvider/mock_localConfigProvider.go 

mockgen -source=pkg/storage/storage.go \
    -package storage \
    -destination pkg/storage/mock_Client.go 

mockgen -source=pkg/url/url.go \
    -package url \
    -destination pkg/url/mock_Client.go

mockgen -source=pkg/devfile/image/image.go \
    -package image \
    -destination pkg/devfile/image/mock_Backend.go

mockgen -source=pkg/odo/cmdline/cmdline.go \
    -package cmdline \
    -destination pkg/odo/cmdline/mock.go

mockgen -source=pkg/application/application.go \
    -package application \
    -destination pkg/application/mock.go

mockgen -source=pkg/project/project.go \
    -package project \
    -destination pkg/project/mock.go

mockgen -source=pkg/preference/preference.go \
    -package preference \
    -destination pkg/preference/mock.go


mockgen -source=pkg/auth/interface.go \
    -package auth \
    -destination pkg/auth/mock.go

mockgen -source=pkg/odo/cli/init/params.go \
    -package init \
    -destination pkg/odo/cli/init/mock_params.go

mockgen -source=pkg/odo/cli/init/params/interface.go \
    -package params \
    -destination pkg/odo/cli/init/params/mock.go

mockgen -source=pkg/catalog/interface.go \
    -package catalog \
    -destination pkg/catalog/mock_catalog.go

mockgen -source=pkg/odo/cli/init/asker/interface.go \
    -package asker \
    -destination pkg/odo/cli/init/asker/mock.go

mockgen -source=pkg/odo/cli/init/registry/interface.go \
    -package registry \
    -destination pkg/odo/cli/init/registry/mock.go
