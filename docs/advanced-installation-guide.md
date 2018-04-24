
# Advanced Installation Guide

Latest binaries build from git master are available at https://dl.bintray.com/odo/odo/latest/.

Builds for latest released versions are at [GitHub releases page](https://github.com/redhat-developer/odo/releases/latest).

## macOS
1. First, you need to enable `kadel/odo` Homebrew Tap:
    ```sh
    brew tap kadel/odo
    ```
2. 
    - If you want to install latest master build
    ```sh
    brew install kadel/odo/odo -- HEAD
    ```
    - If you want to install latest released version
    ```sh
    brew install kadel/odo/odo
    ```

## Linux
### Debian/Ubuntu and other distributions using deb
1. First, you need to add gpg [public key](https://bintray.com/user/downloadSubjectPublicKey?username=bintray) used to sign repositories.
    ```sh
    curl -L https://bintray.com/user/downloadSubjectPublicKey?username=bintray | apt-key add -
    ```
2. Add odo repository to your `/etc/apt/sources.list`
    - If you want to use latest master builds add  `deb https://dl.bintray.com/odo/odo-deb-dev stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/odo/odo-deb-dev stretch main" | sudo tee -a /etc/apt/sources.list
      ```
    - If you want to use latest released version add  `deb https://dl.bintray.com/odo/odo-deb-releases stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/odo/odo-deb-releases stretch main" | sudo tee -a /etc/apt/sources.list
      ```
3. Now you can install `odo` and you would install any other package.
   ```sh
   apt-get update
   apt-get install odo
   ```


### Fedora/Centos/RHEL and other distribution using rpm
1. Add odo repository to your `/etc/yum.repos.d/`
    - If you want to use latest master builds save following text to `/etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo`
        ```
        # /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
        [bintraybintray-odo-odo-rpm-dev]
        name=bintray-odo-odo-rpm-dev
        baseurl=https://dl.bintray.com/odo/odo-rpm-dev
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/odo/odo-rpm-dev/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-dev.repo
        ```
    - If you want to use latest released version save following text to `/etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo`
        ```
        # /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
        [bintraybintray-odo-odo-rpm-releases]
        name=bintray-odo-odo-rpm-releases
        baseurl=https://dl.bintray.com/odo/odo-rpm-releases
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/odo/odo-rpm-releases/rpm -o /etc/yum.repos.d/bintray-odo-odo-rpm-releases.repo
        ```
3. Now you can install `odo` and you would install any other package.
   ```sh
   yum install odo
   # or 'dnf install odo'
   ```

## Windows
Download latest master builds from Bintray [odo.exe](https://dl.bintray.com/odo/odo/latest/windows-amd64/:odo.exe) or 
builds for released versions from [GitHub releases page](https://github.com/redhat-developer/odo/releases).
