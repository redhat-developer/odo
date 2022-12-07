#!/usr/bin/sh

GO_VERSION=1.18.8

#install podman 
apt-get update -y 
apt-get install curl wget gnupg2 make gcc g++ -y
source /etc/os-release
sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_${VERSION_ID}/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list"
wget -nv https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_${VERSION_ID}/Release.key -O- | apt-key add -

apt-get update -qq -y
apt-get -qq --yes install podman

#install golang
wget -c https://golang.org/dl/go$GO_VERSION.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc