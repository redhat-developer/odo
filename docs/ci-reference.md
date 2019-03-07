---
layout: default
permalink: /ci-reference/
redirect_from: 
  - /docs/ci-reference.md/
---

# How to run integration test job in Travis CI

For default oc, use the configuration in .travis.yaml. For example:

```sh
  # Run main e2e tests
    - <<: *base-test
      stage: test
      name: "Main e2e tests"
      script:
        - ./scripts/oc-cluster.sh
        - make bin
        - sudo cp odo /usr/bin
        - oc login -u developer
        - make test-main-e2e
```

If the need presents itself to run odo integration tests against a specific version of Openshift, use env variable `OPENSHIFT_CLIENT_BINARY_URL` to pass the [released](https://github.com/openshift/origin/releases) oc client URL in `.travis.yaml`. For oc v3.10.0, use the configuration:

```sh
  # Run main e2e tests
    - <<: *base-test
      stage: test
      name: "Main e2e tests"
      script:
        - OPENSHIFT_CLIENT_BINARY_URL=https://github.com/openshift/origin/releases/download/v3.10.0/openshift-origin-client-tools-v3.10.0-dd10d17-linux-64bit.tar.gz ./scripts/oc-cluster.sh
        - make bin
        - sudo cp odo /usr/bin
        - oc login -u developer
        - make test-main-e2e
```
