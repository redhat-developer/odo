package CI

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type CIPRRequestor struct {
	amqpURI         string
	sendconn        *amqp.Connection
	rcvconn         *amqp.Connection
	sendqchan       *amqp.Channel
	rcvqchan        *amqp.Channel
	pr              string
	jenkins_build   int
	jenkins_project string
	jenkins_token   string
	done            chan error
	success         chan bool
}

func NewCIPRRequestor(amqpURI, jenkins_project, jenkins_token, pr string) (*CIPRRequestor, error) {
	ciprr := &CIPRRequestor{
		amqpURI:         amqpURI,
		pr:              pr,
		jenkins_project: jenkins_project,
		jenkins_token:   jenkins_token,
		jenkins_build:   -1,
		done:            make(chan error),
		success:         make(chan bool),
	}
	return ciprr, nil
}

func (ciprr *CIPRRequestor) init() error {
	var err error
	ciprr.sendconn, err = amqp.Dial(ciprr.amqpURI)
	if err != nil {
		return fmt.Errorf("unable to dail amqp server send %w", err)
	}
	ciprr.rcvconn, err = amqp.Dial(ciprr.amqpURI)
	if err != nil {
		return fmt.Errorf("unable to dail amqp server rcv %w", err)
	}
	return nil
}

func (ciprr *CIPRRequestor) requestPRBuild() error {
	var err error
	ciprr.sendqchan, err = ciprr.sendconn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open send channel %w", err)
	}
	_, err = ciprr.sendqchan.QueueDeclare(SEND_Q, false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to initialzie sendq %w", err)
	}
	buildreq := NewPRRequestMessage(ciprr.jenkins_project, ciprr.pr, ciprr.jenkins_token)
	buildreqs, err := json.Marshal(buildreq)
	if err != nil {
		return fmt.Errorf("failed to unmarshal build request %w", err)
	}
	err = ciprr.sendqchan.Publish(
		"",
		SEND_Q,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			AppId:           "remote-build",
			Body:            []byte(buildreqs),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish build message %w", err)
	}
	return nil
}

func (ciprr *CIPRRequestor) handleDeliveries(deliveries <-chan amqp.Delivery, success chan bool, done chan error) {
	var err error
	for d := range deliveries {
		//parse kind of message
		m := &Message{}
		err = json.Unmarshal(d.Body, m)
		if err != nil {
			done <- fmt.Errorf("failed to unmarshal message as Message %w", err)
			return
		}
		if ciprr.jenkins_build == -1 {
			if m.IsBuild() {
				bm := NewBuildMessage(-1)
				err = json.Unmarshal(d.Body, bm)
				if err != nil {
					done <- fmt.Errorf("failed to unmarshal message as BuildMessage %w", err)
					return
				}
				ciprr.jenkins_build = bm.Build
			}
		} else if m.Build == ciprr.jenkins_build {
			if m.IsStatus() {
				sm := NewStatusMessage(-1)
				err = json.Unmarshal(d.Body, sm)
				if err != nil {
					done <- fmt.Errorf("failed to unmarshal message as StatusMessage %w", err)
					return
				}
				err = d.Ack(false)
				if err != nil {
					done <- fmt.Errorf("failed to ack message %w", err)
					return
				}
				success <- sm.Success
				ciprr.done <- nil
				break
			} else if m.IsLog() {
				lm := NewLogsMessage(-1)
				err = json.Unmarshal(d.Body, lm)
				if err != nil {
					done <- fmt.Errorf("failed to unmarshal message as LogMessage %w", err)
					return
				}
				fmt.Println(lm.Logs)
			}
		}
		err = d.Ack(false)
		if err != nil {
			done <- fmt.Errorf("failed to ack message %w", err)
			return
		}
	}
	done <- nil
}

func (ciprr *CIPRRequestor) processReplies() error {
	var err error
	ciprr.rcvqchan, err = ciprr.rcvconn.Channel()
	if err != nil {
		return fmt.Errorf("unable to open rcv channel %w", err)
	}
	_, err = ciprr.rcvqchan.QueueDeclare(getPRQueue(ciprr.pr), false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("unable to declare rcv q %w", err)
	}
	deliveries, err := ciprr.rcvqchan.Consume(
		getPRQueue(ciprr.pr), "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("unable to consume from rcv q %w", err)
	}
	go ciprr.handleDeliveries(deliveries, ciprr.success, ciprr.done)
	return nil
}

func (ciprr *CIPRRequestor) Run() error {
	var err error
	err = ciprr.init()
	go func() {
		fmt.Printf("closing: %s", <-ciprr.sendconn.NotifyClose(make(chan *amqp.Error)))
		fmt.Printf("closing: %s", <-ciprr.rcvconn.NotifyClose(make(chan *amqp.Error)))
	}()
	if err != nil {
		return fmt.Errorf("failed to initialize %w", err)
	}
	err = ciprr.requestPRBuild()
	if err != nil {
		return fmt.Errorf("unable to request pr build %w", err)
	}
	err = ciprr.processReplies()
	if err != nil {
		return fmt.Errorf("unable to process replies %w", err)
	}
	return nil
}

func (ciprr *CIPRRequestor) Done() chan error {
	return ciprr.done
}

func (ciprr *CIPRRequestor) Success() chan bool {
	return ciprr.success
}

func (ciprr *CIPRRequestor) ShutDown() error {
	// will close() the send channel
	if err := ciprr.sendqchan.Cancel("", true); err != nil {
		return fmt.Errorf("producer cancel failed: %s", err)
	}

	// will close the deliveries channel
	if err := ciprr.rcvqchan.Cancel("", true); err != nil {
		return fmt.Errorf("consumer cancel failed: %s", err)
	}
	if _, err := ciprr.rcvqchan.QueuePurge(getPRQueue(ciprr.pr), true); err != nil {
		return fmt.Errorf("failed to purge rcv queue %w", err)
	}
	if _, err := ciprr.rcvqchan.QueueDelete(getPRQueue(ciprr.pr), true, true, false); err != nil {
		return fmt.Errorf("failed to delete rcv queue %w", err)
	}
	if err := ciprr.sendconn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}
	if err := ciprr.rcvconn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return nil
}
