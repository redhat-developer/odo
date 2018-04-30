---
layout: default
permalink: /installation/
redirect_from: 
  - /docs/installation.md/
---

# Installation

There are multiple ways of installing Kedge. Our prefered method is downloading the binary from the latest GitHub release.


#### GitHub release

Kedge is released via GitHub on a three-week cycle, you can see all current releases on the [GitHub release page](https://github.com/kedgeproject/kedge/releases).

__Linux and macOS:__

```sh
# Linux
curl -L https://github.com/kedgeproject/kedge/releases/download/v0.5.1/kedge-linux-amd64 -o kedge

# macOS
curl -L https://github.com/kedgeproject/kedge/releases/download/v0.5.1/kedge-darwin-amd64 -o kedge

chmod +x kedge
sudo mv ./kedge /usr/local/bin/kedge
```

__Windows:__

Download from [GitHub](https://github.com/kedgeproject/kedge/releases/download/v0.5.1/kedge-windows-amd64.exe) and add the binary to your PATH.


#### Installing the latest binary (master)

You can download latest binary (built on each master PR merge) for [Linux (amd64)][Bintray Latest Linux], [macOS (darwin)][Bintray Latest macOS] or [Windows (amd64)][Bintray Latest Windows] from [Bintray](https://bintray.com):

__Linux and macOS:__

```sh
# Linux 
curl -L https://dl.bintray.com/kedgeproject/kedge/latest/kedge-linux-amd64 -o kedge

# macOS
curl -L https://dl.bintray.com/kedgeproject/kedge/latest/kedge-darwin-amd64 -o kedge

chmod +x kedge
sudo mv ./kedge /usr/local/bin/kedge
```

__Windows:__

Download from [Bintray](https://dl.bintray.com/kedgeproject/kedge/latest/kedge-windows-amd64.exe) and add the binary to your PATH.

#### Go get

You can also download and build Kedge via Go:

```sh
go get github.com/kedgeproject/kedge
```

[Bintray Latest Linux]:https://dl.bintray.com/kedgeproject/kedge/latest/kedge-linux-amd64
[Bintray Latest macOS]:https://dl.bintray.com/kedgeproject/kedge/latest/kedge-darwin-amd64
[Bintray Latest Windows]:https://dl.bintray.com/kedgeproject/kedge/latest/kedge-windows-amd64.exe
