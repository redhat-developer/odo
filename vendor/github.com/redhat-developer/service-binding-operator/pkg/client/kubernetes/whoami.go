package kubernetes

import (
	"context"
	"fmt"
	"io/ioutil"
	authenticationapi "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"regexp"
	"strings"
)

// source: https://github.com/rajatjindal/kubectl-whoami/

// tokenRetriever helps to retrieve token
type tokenRetriever struct {
	rountTripper http.RoundTripper
	token        string
}

//RoundTrip gets token
func (t *tokenRetriever) RoundTrip(req *http.Request) (*http.Response, error) {
	header := req.Header.Get("authorization")
	switch {
	case strings.HasPrefix(header, "Bearer "):
		t.token = strings.ReplaceAll(header, "Bearer ", "")
	}

	return t.rountTripper.RoundTrip(req)
}

func WhoAmI(config *rest.Config) (string, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return "", err
	}

	tokenRetriever := &tokenRetriever{}
	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		tokenRetriever.rountTripper = rt
		return tokenRetriever
	})

	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	var token string
	// from vendor/k8s.io/client-go/transport/round_trippers.go:HTTPWrappersForConfig function, tokenauth has preference over basicauth
	if transportConfig.HasTokenAuth() {
		if config.BearerTokenFile != "" {
			d, err := ioutil.ReadFile(config.BearerTokenFile)
			if err != nil {
				return "", err
			}

			token = string(d)
		}

		if config.BearerToken != "" {
			token = config.BearerToken
		}
	}

	if token == "" && (config.AuthProvider != nil || config.ExecProvider != nil) {
		sar := &authorizationv1.SelfSubjectRulesReview{
			Spec: authorizationv1.SelfSubjectRulesReviewSpec{
				Namespace: "default",
			},
		}

		_, err := kubeclient.AuthorizationV1().SelfSubjectRulesReviews().Create(context.Background(), sar, v1.CreateOptions{})

		if err != nil {
			return "", err
		}

		token = tokenRetriever.token
	}

	if token != "" {
		result, err := kubeclient.AuthenticationV1().TokenReviews().Create(context.Background(), &authenticationapi.TokenReview{
			Spec: authenticationapi.TokenReviewSpec{
				Token: token,
			},
		}, v1.CreateOptions{})

		if err != nil {
			if k8serrors.IsForbidden(err) {
				return getUsernameFromError(err), nil
			}
			return "", err
		}

		if result.Status.Error != "" {
			return "", fmt.Errorf(result.Status.Error)
		}
		return result.Status.User.Username, nil

	}
	if transportConfig.HasBasicAuth() {
		return fmt.Sprintf("kubecfg:basicauth:%s", config.Username), nil
	}

	if transportConfig.HasCertAuth() {
		return "kubecfg:certauth:admin", nil
	}

	return "", fmt.Errorf("unsupported auth mechanism")
}

func getUsernameFromError(err error) string {
	re := regexp.MustCompile(`^.* User "(.*)" cannot .*$`)
	return re.ReplaceAllString(err.Error(), "$1")
}
