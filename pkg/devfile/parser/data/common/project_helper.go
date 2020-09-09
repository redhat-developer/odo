package common

import "fmt"

// GetDefaultSource get information about primary source
// returns 3 strings: remote name, remote URL, reference(revision)
func (ps GitLikeProjectSource) GetDefaultSource() (string, string, string, error) {
	// get git checkout information
	// if there are multiple remotes we are ignoring them, as we don't need to setup git repository as it is defined here,
	// the only thing that we need is to download the content
	var remoteName, remoteURL, revision string

	if ps.CheckoutFrom != nil && ps.CheckoutFrom.Revision != "" {
		revision = ps.CheckoutFrom.Revision
	}
	if len(ps.Remotes) > 1 {
		if ps.CheckoutFrom == nil {
			return "", "", "", fmt.Errorf("there are multiple git remotes but no checkoutFrom information")
		}
		remoteName = ps.CheckoutFrom.Remote
		if val, ok := ps.Remotes[remoteName]; ok {
			remoteURL = val
		} else {
			return "", "", "", fmt.Errorf("checkoutFrom.Remote is not defined in Remotes")

		}
	} else {
		// there is only one remote, using range to get it as there are not indexes
		for name, url := range ps.Remotes {
			remoteName = name
			remoteURL = url
		}

	}

	return remoteName, remoteURL, revision, nil

}
