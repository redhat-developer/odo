---
title: Odo 2.0.0 General Availability Release!
sidebar_position: -2
author: Charlie Drage
author_url: https://github.com/cdrage
author_image_url: https://avatars.githubusercontent.com/u/6422176?v=4
---


`2.0.0` of odo has been released!

# What's new

#### Changes to the default deployment method

[Devfile](https://devfile.github.io/devfile/index.html) is a file format that is used as odo's new deployment engine. Starting from `2.0.0` onwards, Source-to-Image (S2I) is no longer the default deployment method. S2I is still supported and can now be accessed with the `--s2i` flag from the command-line.

Learn how to deploy your first devfile using devfiles from our [Devfile tutorial](https://odo.dev/docs/deploying-a-devfile-using-odo/).

Example on how to download a starter project and deploy a devfile:

```sh
$ odo create nodejs --starter
Validation
 ✓  Checking devfile existence [22411ns]
 ✓  Checking devfile compatibility [22492ns]
 ✓  Creating a devfile component from registry: DefaultDevfileRegistry [24341ns]
 ✓  Validating devfile component [74471ns]

Starter Project
 ✓  Downloading starter project nodejs-starter from https://github.com/odo-devfiles/nodejs-ex.git [479ms]

Please use `odo push` command to create the component with source deployed

$ odo push

Validation
 ✓  Validating the devfile [132092ns]

Creating Kubernetes resources for component nodejs
 ✓  Waiting for component to start [5s]

Applying URL changes
 ✓  URL http-3000: http://http-3000-nodejs-foobar.myproject.example.com/ created

Syncing to component nodejs
 ✓  Checking files for pushing [1ms]
 ✓  Syncing files to the component [868ms]

Executing devfile commands for component nodejs
 ✓  Executing install command "npm install" [4s]
 ✓  Executing run command "npm start" [2s]

Pushing devfile component nodejs
 ✓  Changes successfully pushed to component
```


#### Deploying a custom Kubernetes controller with odo

With the release of `2.0.0` deploying operators is now out of experimental mode.

Learn how to deploy your first Kubernetes custom controller from our [Operator documentation](https://odo.dev/docs/operator-hub/).

Example on how to deploy your first Operator:

```sh
$ odo catalog list services
  Operators available in the cluster
  NAME                          CRDs
  etcdoperator.v0.9.4           EtcdCluster, EtcdBackup, EtcdRestore

$ odo service create etcdoperator.v0.9.4/EtcdCluster
```

#### `odo debug` is no longer in technical preview

The `odo debug` command is no longer in technical preview.

[Learn how to debug your component via the CLI or VSCode](https://odo.dev/docs/debugging-using-devfile/).

# Installing odo

## Installing odo on Linux

### Binary installation

    # curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-linux-amd64 -o /usr/local/bin/odo
    # chmod +x /usr/local/bin/odo

## Installing odo on macOS

### Binary installation

    # curl -L https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-darwin-amd64 -o /usr/local/bin/odo
    # chmod +x /usr/local/bin/odo

## Installing odo on Windows

### Binary installation

1.  Download the latest  [`odo.exe`](https://mirror.openshift.com/pub/openshift-v4/clients/odo/latest/odo-windows-amd64.exe)   file.

2.  Add the location of your `odo.exe` to your `GOPATH/bin` directory.

### Setting the `PATH` variable for Windows 10

Edit `Environment Variables` using search:

1.  Click **Search** and type `env` or `environment`.

2.  Select **Edit environment variables for your account**.

3.  Select **Path** from the **Variable** section and click **Edit**.

4.  Click **New** and type `C:\go-bin` into the field or click    **Browse** and select the directory, and click **OK**.

### Setting the `PATH` variable for Windows 7/8

The following example demonstrates how to set up a path variable. Your binaries can be located in any location, but this example uses C:\\go-bin as the location.

1.  Create a folder at `C:\go-bin`.

2.  Right click **Start** and click **Control Panel**.

3.  Select **System and Security** and then click **System**.

4.  From the menu on the left, select the **Advanced systems settings**  and click the **Environment Variables** button at the bottom.

5.  Select **Path** from the **Variable** section and click **Edit**.

6.  Click **New** and type `C:\go-bin` into the field or click    **Browse** and select the directory, and click **OK**.

# Full changelog

**New features:**

- implement odo describe for devfile [\#3644](https://github.com/openshift/odo/issues/3644)
- Release 2.0.0 [\#4021](https://github.com/openshift/odo/pull/4021) ([cdrage](https://github.com/cdrage))
- Move Operator Hub out of experimental mode [\#3938](https://github.com/openshift/odo/pull/3938) ([dharmit](https://github.com/dharmit))
- Implement clonePath, update source code sync location [\#3907](https://github.com/openshift/odo/pull/3907) ([adisky](https://github.com/adisky))

**Code Refactoring:**

- "odo link" help message should not check for ClusterServiceVersion support [\#4008](https://github.com/openshift/odo/issues/4008)
- API version and schema version tests should be migrated to devfileV2 [\#3794](https://github.com/openshift/odo/issues/3794)
- Do not check for CSV when initializing odo link command [\#4010](https://github.com/openshift/odo/pull/4010) ([dharmit](https://github.com/dharmit))
- Update odo debug --help screen [\#3963](https://github.com/openshift/odo/pull/3963) ([cdrage](https://github.com/cdrage))
- Clarify description of the force-build flag in help text for odo push [\#3958](https://github.com/openshift/odo/pull/3958) ([johnmcollier](https://github.com/johnmcollier))
- Switch to use project instead of namespace in env [\#3951](https://github.com/openshift/odo/pull/3951) ([GeekArthur](https://github.com/GeekArthur))
- Remove the namespace flag from odo [\#3949](https://github.com/openshift/odo/pull/3949) ([johnmcollier](https://github.com/johnmcollier))
- Migrate devfile cmd validation to validate pkg [\#3912](https://github.com/openshift/odo/pull/3912) ([maysunfaisal](https://github.com/maysunfaisal))
- Remove command group type init [\#3898](https://github.com/openshift/odo/pull/3898) ([adisky](https://github.com/adisky))

**Bugs:**

- "odo link -h" shows same message for 3.x & 4.x clusters [\#3992](https://github.com/openshift/odo/issues/3992)
- make goget-tools fails due to go mod dependency [\#3983](https://github.com/openshift/odo/issues/3983)
- Handle edge case when index file is commented in .gitignore [\#3961](https://github.com/openshift/odo/issues/3961)
- Java component build execution requires pom.xml [\#3943](https://github.com/openshift/odo/issues/3943)
- default registry not initialized when user already has a preference.yaml file [\#3940](https://github.com/openshift/odo/issues/3940)
- `odo url create` shouldn't require a port if only one port exists in the devfile [\#3923](https://github.com/openshift/odo/issues/3923)
- `odo push` with alternate --run-command should push complete file set upon new pod creation [\#3918](https://github.com/openshift/odo/issues/3918)
- converting s2i items to devfile items does not set the Endpoint's name properly [\#3910](https://github.com/openshift/odo/issues/3910)
- Unexpected EOF during watch stream event decoding, watch channel was closed. [\#3905](https://github.com/openshift/odo/issues/3905)
- odo debug serial tests script panic out [\#3897](https://github.com/openshift/odo/issues/3897)
- Default URL does not propagate to `.odo/env/env.yaml` and you cannot delete it. [\#3893](https://github.com/openshift/odo/issues/3893)
- Breaking component create without exposing port [\#3882](https://github.com/openshift/odo/issues/3882)
- odo registry list causes panic if preference has not been setup [\#3842](https://github.com/openshift/odo/issues/3842)
- odo watch goes into infinite push loop if ignore flag is used [\#3819](https://github.com/openshift/odo/issues/3819)
- 'odo create' should properly validate devfiles [\#3778](https://github.com/openshift/odo/issues/3778)
- context flag does not work with devfile url create [\#3767](https://github.com/openshift/odo/issues/3767)
- odo log is unusable for multi container components [\#3711](https://github.com/openshift/odo/issues/3711)
- "odo registry add" adds registry for invalid url in devfileV2 [\#3451](https://github.com/openshift/odo/issues/3451)
- Prints help message based on backend cluster [\#3993](https://github.com/openshift/odo/pull/3993) ([dharmit](https://github.com/dharmit))
- s2i component fix: use Config instead of ContainerConfig for port detection [\#3957](https://github.com/openshift/odo/pull/3957) ([kadel](https://github.com/kadel))
- 3923- url creation with optional port flag [\#3950](https://github.com/openshift/odo/pull/3950) ([yangcao77](https://github.com/yangcao77))
- Add mandatory file ignores when using --ignore flag [\#3942](https://github.com/openshift/odo/pull/3942) ([maysunfaisal](https://github.com/maysunfaisal))
- Fix default registry support [\#3941](https://github.com/openshift/odo/pull/3941) ([GeekArthur](https://github.com/GeekArthur))
- Update s2i image from library for ppc64le [\#3939](https://github.com/openshift/odo/pull/3939) ([sarveshtamba](https://github.com/sarveshtamba))
- update s2i to devfile conversion as per new url design [\#3930](https://github.com/openshift/odo/pull/3930) ([adisky](https://github.com/adisky))
- Add test-case for validating devfiles on component create [\#3908](https://github.com/openshift/odo/pull/3908) ([johnmcollier](https://github.com/johnmcollier))
- Improve URL format validation [\#3900](https://github.com/openshift/odo/pull/3900) ([GeekArthur](https://github.com/GeekArthur))
- implement odo describe for devfile [\#3843](https://github.com/openshift/odo/pull/3843) ([metacosm](https://github.com/metacosm))

**Tests:**

- Test failures while running `test-cmd-push` test suite on ppc64le [\#3539](https://github.com/openshift/odo/issues/3539)
- Test failures while running `test-cmd-storage` test suite on ppc64le [\#3531](https://github.com/openshift/odo/issues/3531)

**Documentation & Discussions:**

- Update installation page to include instructions for VSCode / IDE's [\#3970](https://github.com/openshift/odo/issues/3970)
- Update docs according to schema changes in the command and component struct [\#3925](https://github.com/openshift/odo/issues/3925)
- Help for `odo push -f` should explain that the full set of project source is pushed to the container [\#3919](https://github.com/openshift/odo/issues/3919)
- Make the `odo.dev` front page documentation simpler [\#3887](https://github.com/openshift/odo/issues/3887)
- Add debug examples for "odo debug -h" [\#3871](https://github.com/openshift/odo/issues/3871)
- Remove technology preview feature for debug command [\#3869](https://github.com/openshift/odo/issues/3869)
- Update devfile "odo.dev" doc [\#3868](https://github.com/openshift/odo/issues/3868)
- Documentation for Operator Hub integration in v2 [\#3810](https://github.com/openshift/odo/issues/3810)
- Document on converting s2i to devfile [\#3749](https://github.com/openshift/odo/issues/3749)
- Adds a blog folder [\#4003](https://github.com/openshift/odo/pull/4003) ([cdrage](https://github.com/cdrage))
- Document odo and Operator Hub integration [\#3982](https://github.com/openshift/odo/pull/3982) ([dharmit](https://github.com/dharmit))
- Add instructions on how to install VSCode plugin [\#3977](https://github.com/openshift/odo/pull/3977) ([cdrage](https://github.com/cdrage))
- Update installation page to indicate beta-1 [\#3960](https://github.com/openshift/odo/pull/3960) ([cdrage](https://github.com/cdrage))
- Remove references to Docker support [\#3954](https://github.com/openshift/odo/pull/3954) ([cdrage](https://github.com/cdrage))
- Updates docs to use the new schema changes for commands and components [\#3928](https://github.com/openshift/odo/pull/3928) ([mik-dass](https://github.com/mik-dass))
- Update commands ouputs in docs. [\#3927](https://github.com/openshift/odo/pull/3927) ([boczkowska](https://github.com/boczkowska))

**Closed issues:**

- Determine if we want to keep Docker support in experimental mode, or disable it [\#3955](https://github.com/openshift/odo/issues/3955)
- rename --namespace flag in odo push to --project [\#3948](https://github.com/openshift/odo/issues/3948)
- rename odo env variable namespace to project [\#3947](https://github.com/openshift/odo/issues/3947)
- Test failures while running `test-integration`  and `test-e2e-all` test suite on ppc64le [\#3945](https://github.com/openshift/odo/issues/3945)
- "unknown flag: --s2i" while running odo test suite 'test-generic' on ppc64le [\#3934](https://github.com/openshift/odo/issues/3934)
- odo `make` commands fail on ppc64le after latest changes. [\#3891](https://github.com/openshift/odo/issues/3891)
- Downstream release of the odo cli [\#3852](https://github.com/openshift/odo/issues/3852)
- clonePath should be supported in odo [\#3729](https://github.com/openshift/odo/issues/3729)
- Move devfile command validation to validate pkg [\#3703](https://github.com/openshift/odo/issues/3703)
- `make test` throws "Errorf format %w has unknown verb w" error on ppc64le with latest master [\#3607](https://github.com/openshift/odo/issues/3607)
- Move Operator Hub integration out of Experimental mode [\#3595](https://github.com/openshift/odo/issues/3595)
- Move container image used in springboot devfile to some odo owned image repository [\#3578](https://github.com/openshift/odo/issues/3578)
- Move the devfile feature set out of the experimental mode [\#3550](https://github.com/openshift/odo/issues/3550)
- JSON  / machine output support for Devfile Components [\#3521](https://github.com/openshift/odo/issues/3521)
- Component push throws error of "Waiting for component to start" on ppc64le [\#3497](https://github.com/openshift/odo/issues/3497)
- odo project create throws error of connection refused on ppc64le [\#3491](https://github.com/openshift/odo/issues/3491)
- Tests for devfiles in odo devfile registry [\#3378](https://github.com/openshift/odo/issues/3378)

**Merged pull requests:**

- vendor: switch location of goautoneg to github [\#3984](https://github.com/openshift/odo/pull/3984) ([kadel](https://github.com/kadel))
- Remove url describe command [\#3981](https://github.com/openshift/odo/pull/3981) ([adisky](https://github.com/adisky))
- odo list follow up implementation [\#3964](https://github.com/openshift/odo/pull/3964) ([girishramnani](https://github.com/girishramnani))
- Fix test failure caused by updating springboot devfile [\#3946](https://github.com/openshift/odo/pull/3946) ([adisky](https://github.com/adisky))
- apiVersion test migrated to devfileV2 [\#3920](https://github.com/openshift/odo/pull/3920) ([anandrkskd](https://github.com/anandrkskd))
- add test for odo url create --context flag [\#3917](https://github.com/openshift/odo/pull/3917) ([girishramnani](https://github.com/girishramnani))
- Update springboot devfile [\#3799](https://github.com/openshift/odo/pull/3799) ([adisky](https://github.com/adisky))
- Fix odo log for multi containers devfile [\#3735](https://github.com/openshift/odo/pull/3735) ([adisky](https://github.com/adisky))
- Make Devfile the default deployment mechanism [\#3705](https://github.com/openshift/odo/pull/3705) ([cdrage](https://github.com/cdrage))
