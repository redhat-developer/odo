package version

/*
===================
=    IMPORTANT    =
===================

This package is solely for versioning information when releasing odo..

Changing these values will change the versioning information when releasing odo.
*/

var (
	// VERSION  is version number that will be displayed when running ./odo version
	VERSION = "v3.0.0-alpha3"

	// GITCOMMIT is hash of the commit that will be displayed when running ./odo version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/odo/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)
