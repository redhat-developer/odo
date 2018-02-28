
# Advanced Installation Guide

Latest binaries build from git master are available at https://dl.bintray.com/ocdev/ocdev/latest/.

Builds for latest released versions are at [GitHub releases page](https://github.com/redhat-developer/ocdev/releases/latest).

## macOS
1. First you need enable `kadel/ocdev` Homebrew Tap:
    ```sh
    brew tap kadel/ocdev
    ```
2. 
    - If you want to install latest master build
    ```sh
    brew install kadel/ocdev/ocdev -- HEAD
    ```
    - If you want to install latest released version
    ```sh
    brew install kadel/ocdev/ocdev
    ```

## Linux
### Debian/Ubuntu and other distributions using deb
1. First you need to add gpg [public key](https://bintray.com/user/downloadSubjectPublicKey?username=bintray) used to sign repositories.
    ```sh
    curl -L https://bintray.com/user/downloadSubjectPublicKey?username=bintray | apt-key add -
    ```
2. Add ocdev repository to your `/etc/apt/sources.list`
    - If you want to use latest master builds add  `deb https://dl.bintray.com/ocdev/ocdev-deb-dev stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/ocdev/ocdev-deb-dev stretch main" | sudo tee -a /etc/apt/sources.list
      ```
    - If you want to use latest released version add  `deb https://dl.bintray.com/ocdev/ocdev-deb-releases stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/ocdev/ocdev-deb-releases stretch main" | sudo tee -a /etc/apt/sources.list
      ```
3. Now you can install `ocdev` and you would install any other package.
   ```sh
   apt-get update
   apt-get install ocdev
   ```


### Fedora/Centos/RHEL and other distribution using rpm
1. Add ocdev repository to your `/etc/yum.repos.d/`
    - If you want to use latest master builds save following text to `/etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo`
        ```
        # /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo
        [bintraybintray-ocdev-ocdev-rpm-dev]
        name=bintray-ocdev-ocdev-rpm-dev
        baseurl=https://dl.bintray.com/ocdev/ocdev-rpm-dev
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/ocdev/ocdev-rpm-dev/rpm -o /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo
        ```
    - If you want to use latest released version save following text to `/etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo`
        ```
        # /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo
        [bintraybintray-ocdev-ocdev-rpm-releases]
        name=bintray-ocdev-ocdev-rpm-releases
        baseurl=https://dl.bintray.com/ocdev/ocdev-rpm-releases
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/ocdev/ocdev-rpm-releases/rpm -o /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo
        ```
3. Now you can install `ocdev` and you would install any other package.
   ```sh
   yum install ocdev
   # or 'dnf install ocdev'
   ```

## Windows
Download latest master builds from Bintray [ocdev.exe](https://dl.bintray.com/ocdev/ocdev/latest/windows-amd64/:ocdev.exe) or 
builds for released versions from [GitHub releases page](https://github.com/kadel/ocdev/releases).