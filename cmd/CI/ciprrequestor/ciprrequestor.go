package main

import (
	"log"
	"os"
	"time"

	"github.com/openshift/odo/pkg/CI"
)

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string, failifnotfound bool) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if failifnotfound {
		log.Fatalf("required env %s not set", key)
	}
	return defaultVal
}

func main() {
	var err error
	amqpURI := getEnv("AMQP_URI", "", true)
	jenkins_job := getEnv("JOB_NAME", "", true)
	job_spec := getEnv("JOB_SPEC", "", true)
	job_token := getEnv("JOB_TOKEN", "", true)
	pr_no, err := CI.PRFromJobSpec(job_spec)
	if err != nil {
		log.Fatalf("failed to parse pr no from job_spec %s", err)
	}
	w, err := CI.NewCIPRRequestor(amqpURI, jenkins_job, job_token, pr_no)
	if err != nil {
		log.Fatalf("failed to get requestor %s", err)
	}
	err = w.Run()
	if err != nil {
		log.Fatalf("failed to run prrequestor %s", err)
	}

	log.Println("running ...")
	var success bool
	var done error
	select {
	case done = <-w.Done():
		if done == nil {
			success = <-w.Success()
			log.Printf("Tests success: %t, see logs above ^", success)
			if err := w.ShutDown(); err != nil {
				log.Fatalf("error during shutdown: %s", err)
			}
			if !success {
				log.Fatal("Failure")
			}
		} else {
			log.Printf("failed due to err %s", done)
			log.Println("shutting down")
			if err := w.ShutDown(); err != nil {
				log.Fatalf("error during shutdown: %s", err)
			}
			log.Fatal("Failure")
		}
	case <-time.After(1*time.Hour + 10*time.Minute):
		log.Println("shutting down")
		if err := w.ShutDown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
		log.Fatal("timed out")
	}
}
