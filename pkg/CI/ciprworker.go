package CI

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/streadway/amqp"
)

type CIPRWorker struct {
	amqpURI          string
	conn             *amqp.Connection
	rcvqchan         *amqp.Channel
	pr               string
	jenkins_build    int
	jenkins_job      string
	jenkins_url      string
	jenkins_user     string
	jenkins_password string
}

//NewCIPRWorker returns new ci pr worker
func NewCIPRWorker(amqpURI, jenkins_url, jenkins_user, jenkins_password, jenkins_job, pr string, jenkins_build int) (*CIPRWorker, error) {
	ciprw := &CIPRWorker{
		amqpURI:          amqpURI,
		conn:             nil,
		rcvqchan:         nil,
		pr:               pr,
		jenkins_build:    jenkins_build,
		jenkins_job:      jenkins_job,
		jenkins_url:      jenkins_url,
		jenkins_user:     jenkins_user,
		jenkins_password: jenkins_password,
	}
	return ciprw, nil
}

func (ciprw *CIPRWorker) init() error {
	var err error
	ciprw.conn, err = amqp.Dial(ciprw.amqpURI)
	if err != nil {
		return fmt.Errorf("unable to connect to queue: %w", err)
	}
	rcvchan, err := ciprw.conn.Channel()
	if err != nil {
		return fmt.Errorf("unable to get channel: %w", err)
	}
	rcvqn := getPRQueue(ciprw.pr)
	_, err = rcvchan.QueueDeclare(rcvqn, false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to create q %w", err)
	}
	//setup workspace
	return nil
}

//cleanupOldBuilds ensures there are no conflicting builds for this PR still alive
func (ciprw *CIPRWorker) cleanUpOldBuilds() error {
	var err error
	fmt.Println("logging into jenkins ", ciprw.jenkins_url, " as user ", ciprw.jenkins_user, " with credentials ", ciprw.jenkins_password)
	jenkins, err := gojenkins.CreateJenkins(nil, ciprw.jenkins_url, ciprw.jenkins_user, ciprw.jenkins_password).Init()
	if err != nil {
		return fmt.Errorf("unable to initialize jenkins %w", err)
	}

	job, err := jenkins.GetJob(ciprw.jenkins_job)
	if err != nil {
		return fmt.Errorf("failed to fetch job %s %s", ciprw.jenkins_job, err)
	}
	buildids, err := job.GetAllBuildIds()
	if err != nil {
		return fmt.Errorf("failed to fetch build ids %w", err)
	}
	for _, bid := range buildids {
		if bid.Number != int64(ciprw.jenkins_build) {
			build, err := job.GetBuild(bid.Number)
			if err != nil {
				return fmt.Errorf("failed to get build info %w", err)
			}
			buildenvinfo, err := build.GetInjectedEnvVars()
			if err != nil {
				return fmt.Errorf("unable to get build envinfo %w", err)
			}
			for k, v := range buildenvinfo {
				if k == "PR_NO" && v == ciprw.pr {
					_, err = build.Stop()
					if err != nil {
						return fmt.Errorf("unable to stop build %w", err)
					}
					break
				}
			}
		}
	}
	return nil
}

//sendBuildInfo sends the current build number to requestor. This ensures requestor can drop all messages
//not related to the newest build it should be interested in
func (ciprw *CIPRWorker) sendBuildInfo() error {
	var err error
	bm := NewBuildMessage(ciprw.jenkins_build)
	bmm, err := json.Marshal(bm)
	if err != nil {
		return fmt.Errorf("failed to unmarshal build message %w", err)
	}
	log.Printf("Publishing build Message %s\n", bmm)
	err = ciprw.rcvqchan.Publish(
		"",
		getPRQueue(ciprw.pr),
		true,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/json",
			ContentEncoding: "",
			Body:            []byte(bmm),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish build message %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	return nil
	// return ciprw.ShutDown()
}

func (ciprw *CIPRWorker) runTests() (bool, error) {
	var err error
	sucess := false
	log.Println("this is where tests will run")
	lm := NewLogsMessage(ciprw.jenkins_build)
	lm.Logs = "Tests will happen here"
	lmm, err := json.Marshal(lm)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal logs %w", err)
	}
	err = ciprw.rcvqchan.Publish(
		"",
		getPRQueue(ciprw.pr),
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/json",
			ContentEncoding: "",
			Body:            []byte(lmm),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,
		},
	)
	if err != nil {
		return false, fmt.Errorf("failed to publish logs message %w", err)
	}
	time.Sleep(10 * time.Millisecond)
	// return sucess, ciprw.ShutDown()
	return sucess, nil
}

//runTests runs the tests
// func (ciprw *CIPRWorker) runTests(cmdlist []*exec.Cmd) (bool, error) {
// 	var err error
// 	success := true
// 	done := make(chan struct{})
// 	for _, cmd := range cmdlist {
// 		r, _ := cmd.StdoutPipe()
// 		cmd.Stderr = cmd.Stdout
// 		scanner := bufio.NewScanner(r)
// 		go func() {
// 			for scanner.Scan() {
// 				line := scanner.Text()
// 				fmt.Println(line)
// 			}
// 			done <- struct{}{}
// 		}()
// 		err = cmd.Start()
// 		if err != nil {
// 			return false, fmt.Errorf("failed to run command %w", err)
// 		}send status message
// 		<-done
// 		if cmd.ProcessState.ExitCode() != 0 {
// 			success = false
// 			break
// 		}
// 	}
// 	return success, nil
// }

//sendStatusMessage sends the status of the build
func (ciprw *CIPRWorker) sendStatusMessage(success bool) error {
	sm := NewStatusMessage(ciprw.jenkins_build)
	sm.Success = success
	smm, err := json.Marshal(sm)
	if err != nil {
		return fmt.Errorf("failed to unmarshal success msg %w", err)
	}
	err = ciprw.rcvqchan.Publish(
		"",
		getPRQueue(ciprw.pr),
		true,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/json",
			ContentEncoding: "",
			Body:            []byte(smm),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

//Runs runs the CIPRWorker. Note this should only be called ONCE
func (ciprw *CIPRWorker) Run() error {
	//initialize. This is done in run so cleanup can be handled correctly
	var err error
	//Check jenkins for existing builds and clean them up
	log.Println("[x] finding and cleaning up old builds")
	err = ciprw.cleanUpOldBuilds()
	if err != nil {
		return fmt.Errorf("failed to cleanup old builds %w", err)
	}
	log.Println("[x] initializing")
	err = ciprw.init()
	if err != nil {
		return fmt.Errorf("failed to initialize")
	}
	//Send current build information to requestor. This assumes requestor is dropping every message
	//until it gets this message
	log.Println("[x] sending build information")
	err = ciprw.sendBuildInfo()
	if err != nil {
		return fmt.Errorf("unable to publish build info %w", err)
	}
	//run the nessasary tests and stream the logs
	log.Println("[x] running tests")
	success, err := ciprw.runTests()
	if err != nil {
		return fmt.Errorf("failed to run tests %w", err)
	}
	//Send status message
	log.Println("[x] sending status")
	err = ciprw.sendStatusMessage(success)
	if err != nil {
		return fmt.Errorf("failed to send status message %w", err)
	}
	return nil
}

func (ciprw *CIPRWorker) ShutDown() error {
	defer ciprw.rcvqchan.Close()
	defer ciprw.conn.Close()
	return nil
}
