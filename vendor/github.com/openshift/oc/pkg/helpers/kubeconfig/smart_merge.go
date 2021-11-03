package kubeconfig

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	x509request "k8s.io/apiserver/pkg/authentication/request/x509"

	"k8s.io/apimachinery/third_party/forked/golang/netutil"
	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// getClusterNicknameFromConfig returns host:port of the clientConfig.Host, with .'s replaced by -'s
// TODO this is copied from pkg/client/config/smart_merge.go, looks like a good library-go candidate
func getClusterNicknameFromConfig(clientCfg *restclient.Config) (string, error) {
	u, err := url.Parse(clientCfg.Host)
	if err != nil {
		return "", err
	}
	hostPort := netutil.CanonicalAddr(u)

	// we need a character other than "." to avoid conflicts with.  replace with '-'
	return strings.Replace(hostPort, ".", "-", -1), nil
}

// getContextNicknameFromConfig returns "namespace/getClusterNicknameFromConfig/username(as known by the server)".  This allows tab completion for switching projects/context
// to work easily.  First tab is the most selective on project.  Second stanza in the next most selective on cluster name.  The chances of a user trying having
// one projects on a single server that they want to operate against with two identities is low, so username is last.
func getContextNicknameFromConfig(namespace, userName string, clientCfg *restclient.Config) (string, error) {
	clusterNick, err := getClusterNicknameFromConfig(clientCfg)
	if err != nil {
		return "", err
	}

	return namespace + "/" + clusterNick + "/" + userName, nil
}

// CreateConfig takes a clientCfg and builds a config (kubeconfig style) from it.
func CreateConfig(namespace, userName string, clientCfg *restclient.Config) (*clientcmdapi.Config, error) {
	clusterNick, err := getClusterNicknameFromConfig(clientCfg)
	if err != nil {
		return nil, err
	}

	contextNick, err := getContextNicknameFromConfig(namespace, userName, clientCfg)
	if err != nil {
		return nil, err
	}
	userName = userName + "/" + clusterNick

	config := clientcmdapi.NewConfig()

	credentials := clientcmdapi.NewAuthInfo()
	credentials.Token = clientCfg.BearerToken
	credentials.TokenFile = clientCfg.BearerTokenFile
	credentials.AuthProvider = clientCfg.AuthProvider
	credentials.Exec = clientCfg.ExecProvider
	credentials.ClientCertificate = clientCfg.TLSClientConfig.CertFile
	if len(credentials.ClientCertificate) == 0 {
		credentials.ClientCertificateData = clientCfg.TLSClientConfig.CertData
	}
	credentials.ClientKey = clientCfg.TLSClientConfig.KeyFile
	if len(credentials.ClientKey) == 0 {
		credentials.ClientKeyData = clientCfg.TLSClientConfig.KeyData
	}
	config.AuthInfos[userName] = credentials

	cluster := clientcmdapi.NewCluster()
	cluster.Server = clientCfg.Host

	if clientCfg.Proxy != nil {
		req, err := http.NewRequest("", clientCfg.Host, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy URL request for execProvider: %w", err)
		}
		proxyURL, err := clientCfg.Proxy(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get proxy URL for execProvider: %w", err)
		}
		if proxyURL != nil {
			cluster.ProxyURL = proxyURL.String()
		}
	}

	cluster.CertificateAuthority = clientCfg.CAFile
	if len(cluster.CertificateAuthority) == 0 {
		cluster.CertificateAuthorityData = clientCfg.CAData
	}
	cluster.InsecureSkipTLSVerify = clientCfg.Insecure
	config.Clusters[clusterNick] = cluster

	context := clientcmdapi.NewContext()
	context.Cluster = clusterNick
	context.AuthInfo = userName
	context.Namespace = namespace
	config.Contexts[contextNick] = context
	config.CurrentContext = contextNick

	return config, nil
}

// MergeConfig adds the additional Config stanzas to the startingConfig.  It blindly stomps clusters and users, but
// it searches for a matching context before writing a new one.
func MergeConfig(startingConfig, addition clientcmdapi.Config) (*clientcmdapi.Config, error) {
	ret := startingConfig

	for requestedKey, value := range addition.Clusters {
		ret.Clusters[requestedKey] = value
	}

	for requestedKey, value := range addition.AuthInfos {
		ret.AuthInfos[requestedKey] = value
	}

	requestedContextNamesToActualContextNames := map[string]string{}
	for requestedKey, newContext := range addition.Contexts {
		actualContext := clientcmdapi.NewContext()
		actualContext.AuthInfo = newContext.AuthInfo
		actualContext.Cluster = newContext.Cluster
		actualContext.Namespace = newContext.Namespace
		actualContext.Extensions = newContext.Extensions

		if existingName := findExistingContextName(startingConfig, *actualContext); len(existingName) > 0 {
			// if this already exists, just move to the next, our job is done
			requestedContextNamesToActualContextNames[requestedKey] = existingName
			continue
		}

		requestedContextNamesToActualContextNames[requestedKey] = requestedKey
		ret.Contexts[requestedKey] = actualContext
	}

	if len(addition.CurrentContext) > 0 {
		if newCurrentContext, exists := requestedContextNamesToActualContextNames[addition.CurrentContext]; exists {
			ret.CurrentContext = newCurrentContext
		} else {
			ret.CurrentContext = addition.CurrentContext
		}
	}

	return &ret, nil
}

// findExistingContextName finds the nickname for the passed context
func findExistingContextName(haystack clientcmdapi.Config, needle clientcmdapi.Context) string {
	for key, context := range haystack.Contexts {
		context.LocationOfOrigin = ""
		if reflect.DeepEqual(context, needle) {
			return key
		}
	}

	return ""
}

// GetClusterNicknameFromURL returns host:port of the apiServerLocation, with .'s replaced by -'s
func GetClusterNicknameFromURL(apiServerLocation string) (string, error) {
	u, err := url.Parse(apiServerLocation)
	if err != nil {
		return "", err
	}
	hostPort := netutil.CanonicalAddr(u)

	// we need a character other than "." to avoid conflicts with.  replace with '-'
	return strings.Replace(hostPort, ".", "-", -1), nil
}

func GetUserNicknameFromCert(clusterNick string, chain ...*x509.Certificate) (string, error) {
	authResponse, _, err := x509request.CommonNameUserConversion(chain)
	if err != nil {
		return "", err
	}

	return authResponse.User.GetName() + "/" + clusterNick, nil
}

func GetContextNickname(namespace, clusterNick, userNick string) string {
	tokens := strings.SplitN(userNick, "/", 2)
	return namespace + "/" + clusterNick + "/" + tokens[0]
}

var validURLSchemes = []string{"https://", "http://", "tcp://"}

// NormalizeServerURL is opinionated normalization of a string that represents a URL. Returns the URL provided matching the format
// expected when storing a URL in a config. Sets a scheme and port if not present, removes unnecessary trailing
// slashes, etc. Can be used to normalize a URL provided by user input.
func NormalizeServerURL(s string) (string, error) {
	// normalize scheme
	if !hasScheme(s) {
		s = validURLSchemes[0] + s
	}

	addr, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("Not a valid URL: %v.", err)
	}

	// normalize host:port
	if strings.Contains(addr.Host, ":") {
		_, port, err := net.SplitHostPort(addr.Host)
		if err != nil {
			return "", fmt.Errorf("Not a valid host:port: %v.", err)
		}
		_, err = strconv.ParseUint(port, 10, 16)
		if err != nil {
			return "", fmt.Errorf("Not a valid port: %v. Port numbers must be between 0 and 65535.", port)
		}
	} else {
		port := 0
		switch addr.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			return "", fmt.Errorf("No port specified.")
		}
		addr.Host = net.JoinHostPort(addr.Host, strconv.FormatInt(int64(port), 10))
	}

	// remove trailing slash if that's the only path we have
	if addr.Path == "/" {
		addr.Path = ""
	}

	return addr.String(), nil
}

func hasScheme(s string) bool {
	for _, p := range validURLSchemes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
