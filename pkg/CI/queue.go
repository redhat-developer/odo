package CI

import "fmt"

const (
	SEND_Q       = "CI_SEND"
	RCV_Q_PREFIX = "CI_RECIEVE"
)

func getPRQueue(pr string) string {
	return fmt.Sprintf("%s_%s", RCV_Q_PREFIX, pr)
}
