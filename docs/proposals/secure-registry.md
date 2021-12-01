# Secure Registry Support

Table of contents
- [Problem Statement](#problem-statement)
- [Terminology](#terminology)
- [Proposed Design](#proposed-design)
- [Related Issues](#related-issues)

## Problem Statement
Currently odo only supports registry that is hosted by the platform that has publicly signed certificate, we should support secure registry so that user is able to store the confidential devfile to the registry and let the platform with certificate in user's trust store host the registry, authentication is needed on user side to access the platform.

## Terminology
Registry: registry is the place that stores index file (index.json) and devfile (devfile.yaml) so that user can catalog and create devfile component from the registry. The registry itself can be hosted on GitHub (GitHub-hosted registry) or Cluster (Cluster-hosted registry)

Authentication method (Credential):
- Username/Password: usually user has full access to the resource with Username/Password authentication method.
- Personal Access Token (PAT): this is the recommended authentication method as user can grant limited resource access to the personal access token to make the resource access more secure. 

## Proposed Design
Support Scenarios:

In order to make secure registry support feature more clear and specific, we should support the following scenarios:
1. GitHub-hosted registry
   1. GitHub public:
   - Clients authenticate with GitHub personal access token
   - TLS achieved with Git public CA signed certification in client trust store
   2. GitHub Enterprise:
   - Clients authenticate with GHE personal access token
   - TLS achieved with GHE public CA signed certification in client trust store
2. Cluster-hosted registry
   - Clients authenticate with service account token
   - TLS achieved with ingress gateway CA signed certification in client trust store

Context:
1. Given GitHub is going to depreciate basic authentication with username/password https://developer.github.com/changes/2020-02-14-deprecating-password-auth/, we have to only support personal access token authentication method for GitHub-hosted registry scenario.
2. For cluster-hosted registry, the registry architecture would be creating a NGINX server to host the secure registry, then create a ingress gateway for the NGNIX server to let client access.

Work flow to access secure registry:
1. Collect Credential

    Currently we have `odo registry add <registry name> <registry URL>` and `odo registry update <registry name> <registry URL>` to add and update registry accordingly. Regarding the CLI design, we can implement the following CLI design to support collecting credential:
    - `odo registry add <registry name> <registry URL> --token <token>`
    - `odo registry add <registry name> <registry URL> --user <user> --password <password>`
    - `odo registry update <registry name> <registry URL> --token <token>`
    - `odo registry update <registry name> <registry URL> --user <user> --password <password>`

2. Store Credential

    We can use third-party package keyring(https://github.com/zalando/go-keyring) to help store user's credential, this package is platform agnostic, which means it can automatically use the existing keyring instance on the platform, for example:
    - Mac: Implementation depends on the /usr/bin/security binary for interfacing with the OS X keychain.
    - Linux: Implementation depends on the Secret Service dbus interface, which is provided by GNOME Keyring.
    - Windows: Windows Credential Manager support by using the lib https://github.com/danieljoos/wincred.

3. Use Credential

    We can still use the built-in package `net/http` with adding token to the request header, sample code:
    ```
    token := "123abc"
    bearer := "Bearer " + token
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
	    log.Println(err)
    }
    req.Header.Add("Authorization", bearer)
    ```

4. Delete Credential

    If multiple secure registries share the same credential, `odo registry delete <registry name>` will delete the credential from keyring instance once the last secure registry using that credential has been deleted.

Create devfile component from secure registry:

When downloading devfile from secure registry, we validate if the credential is valid by adding token to the request header and checking the response.

## Related issues
- Dynamic registry support: https://github.com/redhat-developer/odo/pull/2940
- Performance improvement for `odo catalog list components`: https://github.com/redhat-developer/odo/pull/3112