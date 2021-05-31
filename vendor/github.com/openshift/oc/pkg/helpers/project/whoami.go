package project

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	userv1 "github.com/openshift/api/user/v1"
	userv1typedclient "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
)

func WhoAmI(clientConfig *restclient.Config) (*userv1.User, error) {
	client, err := userv1typedclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	me, err := client.Users().Get(context.TODO(), "~", metav1.GetOptions{})

	// if we're talking to kube (or likely talking to kube),
	if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
		switch {
		case len(clientConfig.BearerToken) > 0:
			// convert their token to a hash instead of printing it
			h := sha256.New()
			h.Write([]byte(clientConfig.BearerToken))
			tokenName := fmt.Sprintf("token-%s", base64.RawURLEncoding.EncodeToString(h.Sum(nil)[:9]))
			return &userv1.User{ObjectMeta: metav1.ObjectMeta{Name: tokenName}}, nil

		case len(clientConfig.Username) > 0:
			return &userv1.User{ObjectMeta: metav1.ObjectMeta{Name: clientConfig.Username}}, nil

		}
	}

	if err != nil {
		return nil, err
	}

	return me, nil
}
