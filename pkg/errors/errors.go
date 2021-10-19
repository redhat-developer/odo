package errors

import "fmt"

// The error built-in interface type is the conventional interface for
// representing an error condition, with the nil value representing no error.
type error interface {
	Error() string
}

const loginMessage = `Please login to your server: 

odo login https://mycluster.mydomain.com
`

type Unauthorized struct {
}

func (u *Unauthorized) Error() string {
	return fmt.Sprintf("Unauthorized to access the cluster\n%s", loginMessage)
}
