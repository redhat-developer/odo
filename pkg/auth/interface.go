package auth

type Client interface {
	Login(server, username, password, token, caAuth string, skipTLS bool) error
}
