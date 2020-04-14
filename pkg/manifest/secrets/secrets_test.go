package secrets

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"testing"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/openshift/odo/pkg/manifest/meta"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	testModulus = new(big.Int)
	_, err := fmt.Sscan("777304254876434297689544225447769213262492599515515837291621795936355252933930193245809942636192119684040605554803489669141565417296821660595336672178414512660751886699171738066307588619202437848899334837760648051656982184646490661921128886671800776058692981991859399404705935722225294811424879738586269551402668122524371718537515440568440102201259925611463161144897905846190044735554045001999198442528435295995584980713050916813579912296878368079243909549993116827192901474611239264189340401059113919551426849847211275352102674049634252149163111599977742365280992561904350781270344655927564475032580504276518647106167707150111291732645399166011800154961975117045723373023335778593638216165426988399138193230056486079421256484837299169853958601000282124667227789126483641999102102039577368681983584245367307077546423870452524154641890843463963116237003367269116435430641427113406369059991147359641266708862913786891945896441771663010146473536372286482453315017377528517965715554550898957321536181165129538808789201530141159181590893764287807749414277289452691723903046140558704697831351834538780165261072894792900501671534138992265545905216973214953125367388406669893889742303072755608685449114438926280862339744991872488262084141163", testModulus)
	if err != nil {
		panic(err)
	}
}

const testCert = `
-----BEGIN CERTIFICATE-----
MIIErTCCApWgAwIBAgIQBekz48i8NbrzIpIrLMIULTANBgkqhkiG9w0BAQsFADAA
MB4XDTE3MDYyMDA0MzI0NVoXDTI3MDYxODA0MzI0NVowADCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBAL6ISW4MnHAmC6MdmJOwo9C6YYhKYDwPD2tF+j4p
I2duB3y7DLF+zWNHgbUlBZck8CudacJTuxOJFEqr4umqm0f4EGgRPwZgFvFLHKSZ
/hxUFnMcGVhY1qsk55peSghPHarOYyBhhHDtCu7qdMu9MqPZB68y16HdPvwWPadI
dBKSxDLvwYfjDnG/ZHX9rmlDKej7jPGdvqAY5VJteP30w6YHb1Uc4whppNcDSc2l
gOuKAWtQ5WfZbB0NpMhj4framNeXMYwjZytEdC1c/4O45zm5eK4FNPueCfxOlzFQ
D3y34OuQlJwlrPE4KmdMHtE1a8x0ihbglInJrtqcXK3vEdUJ2c/BKWgFtPOTz6Du
jV4j0OMVVGnk5jUmh+yfbgielIkPcpSTWP1cIPwK3eWbrvMziq6sv0x7QoOD3Pzm
GBE8Y9sa5uy+bJZt5MywbamZ3xWaxoQbSN8RPoxRhTe0DEpx6utCXSWpapT7kWZ3
R1PTuVx+Ktyz7MRoDUWvxfpMJ2hsJ71Az0AuUZ4N4fmmGdUcM81GPUOiMZ4uqySQ
A2phgikbJaTzcT85RcNFYSi4eKc5mYFNqr5xVa6uHhZ+OGeGy1yyOEWLgIZV3A/8
4eZshOyYtRlZjCkaGZTfXNft+8QJi8rEZRcJtVhqLzezBVRsL7pt6P/mQj4+XHsE
VSBrAgMBAAGjIzAhMA4GA1UdDwEB/wQEAwIAATAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4ICAQCSizqBB3bjHCSGk/8lpqIyHJQR5u4Cf7LRrC9U8mxe
pvC3Fx3/RlVe87Y4cUb37xZc/TmB6Bq10Y6R7ydS3oe8PCh4UQRnEfBgtJ6m59ha
t3iPX0NdQVYz/D+yEiHjpI7gpyFNuGkd4/78JE51SO4yGYvWk/ChHoMvbLcxzfdK
PI2Ymf3MWtGfoF/TQ1jy/Biy+qumDPSz23MynQG39cdUInSK26oemUbTH0koLulN
fNl4TwSEdSm2DRl0la+vkrzu7SvF9SJ2ES6wMWVjYiJLNpApjGuF9/ZOFw9DvSSH
m+UYXn+IC7rTgvXKvXTlG//z/14Lx0GFIY+ZjdENwLH//orBQLg37TZatKEpaWO6
uRzFUxZVw3ic3RxoHfEbRA9vQlQdKnV+BpZe/Pb08RAh82OZyujqqyK7cPPOW5Vi
T9y+NeMwfKH8H4un7mQWkgWFw3LMIspYY5uHWp6jBwU9u/mjoK4+Y219dkaAhAcx
D+YIZRXwxc6ehLCavGF2DIepybzDlJbiCe8JxUDsrE/Xkm6x28uq35oZ3UQznubU
7LfAeRSI99sNvFnq0TqhSlp+CUDs8Z1LvDXzAHX4UeZQl4g+H+w1KudCvjO0mPPp
R9bIjJLIvp7CQPDkdRzJSjvetrKtI0l97VjsjbRB9v6ZekGY9SFI49KzKUTk8fsF
/A==
-----END CERTIFICATE-----
`

var (
	testModulus  *big.Int
	testExponent = 65537
)

func TestSeal(t *testing.T) {

	testCases := []struct {
		secret     corev1.Secret
		want       ssv1alpha1.SealedSecret // partial object
		errMessage string
	}{
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "myns",
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
				StringData: map[string]string{
					"foos": "stringsekret",
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "myns",
				},
			},
			errMessage: "",
		},
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mysecret",
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
				},
			},
			errMessage: "secret must declare a namespace",
		},
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "",
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretNamespaceWideAnnotation: "true",
					},
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretNamespaceWideAnnotation: "true",
					},
				},
			},
			errMessage: "secret must declare a namespace",
		},
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "",
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretClusterWideAnnotation: "true",
					},
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "", // <--- we shouldn't force the default namespace for cluster wide secrets ...
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretClusterWideAnnotation: "true",
					},
				},
			},
			errMessage: "",
		},
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "myns",
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretClusterWideAnnotation: "true",
					},
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "myns", // <--- ... but we should preserve one if specified.
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretClusterWideAnnotation: "true",
					},
				},
			},
		},
		{
			secret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "",
				},
				Data: map[string][]byte{
					"foo": []byte("sekret"),
				},
			},
			want: ssv1alpha1.SealedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
					Annotations: map[string]string{
						ssv1alpha1.SealedSecretNamespaceWideAnnotation: "true",
					},
				},
			},
			errMessage: "secret must declare a namespace",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result, err := seal(&tc.secret, getTestCert)
			if err != nil {
				if diff := cmp.Diff(tc.errMessage, err.Error()); diff != "" {
					t.Errorf("Unexpected error \n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.errMessage, ""); diff != "" {
					t.Errorf("Unexpected error \n%s", diff)
				}
				smeta := result.GetObjectMeta()
				if got, want := smeta.GetName(), tc.want.GetName(); got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
				if got, want := smeta.GetNamespace(), tc.want.GetNamespace(); got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
				if got, want := smeta.GetAnnotations(), tc.want.GetAnnotations(); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
					t.Errorf("got: %q, want: %q", got, want)
				}

				for n := range tc.secret.Data {
					if len(result.Spec.EncryptedData[n]) < 100 {
						t.Errorf("Encrypted data is implausibly short: %v", result.Spec.EncryptedData[n])
					}
				}
				for n := range tc.secret.StringData {
					if len(result.Spec.EncryptedData[n]) < 100 {
						t.Errorf("Encrypted data is implausibly short: %v", result.Spec.EncryptedData[n])
					}
				}
			}

			// NB: See sealedsecret_test.go for e2e crypto test
		})
	}
}

func getTestCert() (*rsa.PublicKey, error) {
	return parseKey(strings.NewReader(testCert))
}

func TestCreateOpaqueSecret(t *testing.T) {
	data := "abcdefghijklmnop"
	secret, err := createOpaqueSecret(meta.NamespacedName("cicd", "github-auth"), data, "token")
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-auth",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(data),
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createOpaqueSecret() failed got\n%s", diff)
	}
}

func TestCreateDockerConfigSecretWithErrorReading(t *testing.T) {
	testErr := errors.New("test failure")
	_, err := createDockerConfigSecret(meta.NamespacedName("cici", "github-auth"), errorReader{testErr})
	if !matchError(t, "failed to read .* test failure", err) {
		t.Fatalf("got an unexpected error: %#v", err)
	}
}

func TestCreateDockerConfigSecret(t *testing.T) {
	data := []byte(`abcdefghijklmnop`)
	secret, err := createDockerConfigSecret(meta.NamespacedName("cicd", "regcred"), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: secretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regcred",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": data,
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createDockerConfigSecret() failed got\n%s", diff)
	}
}

type errorReader struct {
	err error
}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}

func matchError(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}

func TestParseKey(t *testing.T) {
	key, err := parseKey(strings.NewReader(testCert))
	if err != nil {
		t.Fatalf("Failed to parse test key: %v", err)
	}

	if key.N.Cmp(testModulus) != 0 {
		t.Errorf("Unexpected key modulus: %v", key.N)
	}

	if key.E != testExponent {
		t.Errorf("Unexpected key exponent: %v", key.E)
	}
}
