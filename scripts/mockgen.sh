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
