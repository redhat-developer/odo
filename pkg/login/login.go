package login

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/golang/glog"

	userclientset "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/occlient"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// csrfTokenHeader is a marker header that indicates we are not a browser that got tricked into requesting basic auth
	// Corresponds to the header expected by basic-auth challenging authenticators
	// Copied from pkg/auth/authenticator/challenger/passwordchallenger/password_auth_handler.go
	csrfTokenHeader = "X-CSRF-Token"

	// openShiftCLIClientID is the name of the CLI OAuth client, copied from pkg/oauth/apiserver/auth.go
	openShiftCLIClientID = "openshift-challenging-client"
)

func Login(c *occlient.Client, server, username, password, token string) error {
	config, err := c.KubeConfig.ClientConfig()
	if err != nil {
		return err
	}

	if server == "" {
		server = config.Host
	}

	if username != "" && password != "" {
		ctx := context.Background()
		conf := &oauth2.Config{
			ClientID: openShiftCLIClientID,
			// RedirectURL: "https://<server>:8443/oauth/token/implicit",
			Endpoint: oauth2.Endpoint{
				// URL cen be obtained from  https://<server>:8443/.well-known/oauth-authorization-server
				AuthURL:  fmt.Sprintf("%s/oauth/authorize", server),
				TokenURL: fmt.Sprintf("%s/oauth/token", server),
			},
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// Skipping TLS verification this has to be handled in more starter way
					// always ignoring verification can be dangerous
					InsecureSkipVerify: true,
				},
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

		// this is special client used only to request code
		// CheckRedirect ensures that it is not following 302 redirects and we can parse Location header to get code from it
		codeHTTPClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// get an url where we can obtain code
		authCodeUrl := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		//fmt.Printf("authCodeUrl: %s\n", authCodeUrl)
		authCodeReq, err := http.NewRequest("GET", authCodeUrl, nil)
		authCodeReq.Header.Set(csrfTokenHeader, "1")
		authCodeReq.SetBasicAuth(username, password)

		authCodeResp, err := codeHTTPClient.Do(authCodeReq)
		if err != nil {
			panic(err)
		}

		// we don't really need body, but in case of error there might be useful information in it
		body, err := ioutil.ReadAll(authCodeResp.Body)
		if err != nil {
			panic(err)
		}
		glog.V(4).Infof("body: %s\n", string(body[:]))
		//fmt.Printf("body: %s\n", string(body[:]))
		//fmt.Printf("status: %d\n", authCodeResp.StatusCode)
		//fmt.Printf("location: %s\n", authCodeResp.Header.Get("Location"))

		// parse Location and get code from it
		urlLocation, err := url.Parse(authCodeResp.Header.Get("Location"))
		//fmt.Printf("urllocation: %+v\n", urlLocation)
		if err != nil {
			panic(err)
		}
		code := urlLocation.Query().Get("code")
		//fmt.Printf("code: %s\n", code)

		// Exchange code for token
		tok, err := conf.Exchange(ctx, code)
		if err != nil {
			panic(err)
		}

		// And here we have our token
		//fmt.Printf("token: %+v\n", tok)
		token = tok.AccessToken

	}

	config.BearerToken = token
	userClient, err := userclientset.NewForConfig(config)
	if err != nil {
		return err
	}

	c.UserClient = userClient

	me, err := c.UserClient.Users().Get("~", metav1.GetOptions{})

	if err != nil {
		return errors.Wrap(err, "failed to retrieve token")
	}
	// gathering project info
	currentProject := c.GetCurrentProjectName()

	rawConfig, err := c.KubeConfig.RawConfig()

	//fetching configs from cluster and create config struct
	newConfig, err := CreateConfig(currentProject, config)

	configTOWrite, err := MergeConfig(rawConfig, *newConfig)
	if err != nil {
		return err
	}
	err = clientcmd.ModifyConfig(clientcmd.NewDefaultClientConfigLoadingRules(), *configTOWrite, true)
	if err != nil {
		return errors.Wrapf(err, "unable to write to config file")
	}

	fmt.Printf("Logged into %s as %s.\n\n", server, me.Name)
	return nil
}

// CreateConfig takes a clientCfg and builds a config (kubeconfig style) from it.
func CreateConfig(namespace string, clientCfg *restclient.Config) (*clientcmdapi.Config, error) {
	clusterNick, err := getClusterNicknameFromConfig(clientCfg)
	if err != nil {
		return nil, err
	}

	userNick, err := getUserNicknameFromConfig(clientCfg)
	if err != nil {
		return nil, err
	}
	//
	contextNick, err := getContextNicknameFromConfig(namespace, clientCfg)
	if err != nil {
		return nil, err
	}

	config := clientcmdapi.NewConfig()

	credentials := clientcmdapi.NewAuthInfo()
	credentials.Token = clientCfg.BearerToken
	credentials.ClientCertificate = clientCfg.TLSClientConfig.CertFile
	if len(credentials.ClientCertificate) == 0 {
		credentials.ClientCertificateData = clientCfg.TLSClientConfig.CertData
	}
	credentials.ClientKey = clientCfg.TLSClientConfig.KeyFile
	if len(credentials.ClientKey) == 0 {
		credentials.ClientKeyData = clientCfg.TLSClientConfig.KeyData
	}
	config.AuthInfos[userNick] = credentials

	cluster := clientcmdapi.NewCluster()
	cluster.Server = clientCfg.Host
	cluster.CertificateAuthority = clientCfg.CAFile
	if len(cluster.CertificateAuthority) == 0 {
		cluster.CertificateAuthorityData = clientCfg.CAData
	}
	cluster.InsecureSkipTLSVerify = clientCfg.Insecure
	config.Clusters[clusterNick] = cluster

	context := clientcmdapi.NewContext()
	context.Cluster = clusterNick
	context.AuthInfo = userNick
	context.Namespace = namespace
	config.Contexts[contextNick] = context
	config.CurrentContext = contextNick

	return config, nil
}
