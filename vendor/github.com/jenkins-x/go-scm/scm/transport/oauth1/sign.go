// Copyright (c) 2015 Dalton Hubble. All rights reserved.
// Copyrights licensed under the MIT License.

package oauth1

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strings"
)

// signatureBase returns the OAuth1 signature base string
// according to RFC5849 3.4.1.
func signatureBase(req *http.Request, params map[string]string) string {
	method := strings.ToUpper(req.Method)
	baseURL := baseURI(req)
	parameterString := encodeParameterString(params)
	baseParts := []string{method,
		percentEncode(baseURL),
		percentEncode(parameterString)}
	return strings.Join(baseParts, "&")
}

// sign calculates the signature of the message SHA1 digests
// using the given RSA private key.
func sign(privateKey *rsa.PrivateKey, message string) (string, error) {
	digest := sha1.Sum([]byte(message))
	signature, err := rsa.SignPKCS1v15(
		rand.Reader, privateKey, crypto.SHA1, digest[:])
	return base64.StdEncoding.EncodeToString(signature), err
}
