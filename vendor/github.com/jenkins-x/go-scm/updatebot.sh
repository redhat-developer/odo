#!/bin/bash
jx step create pr go --name github.com/jenkins-x/go-scm --version ${VERSION} --build "make mod" --repo https://github.com/jenkins-x/lighthouse.git
jx step create pr go --name github.com/jenkins-x/go-scm --version ${VERSION} --build "make mod" --repo https://github.com/cloudbees/jx-tenant-service.git
jx step create pr go --name github.com/jenkins-x/go-scm --version ${VERSION} --build "make mod" --repo https://github.com/cloudbees/lighthouse-githubapp.git