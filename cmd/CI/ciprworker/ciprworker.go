package main

import (
	"log"
	"os"
	"strconv"

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
	//extract nessasary information
	//var err error
	amqpURI := getEnv("AMQP_URI", "", true)
	jenkins_url := getEnv("JENKINS_URL", "", true)
	jenkins_user := getEnv("JENKINS_ROBOT_USER", "", true)
	jenkins_password := getEnv("JENKINS_ROBOT_PASSWORD", "", true)
	jenkins_job := getEnv("JOB_NAME", "", true)
	bn := getEnv("BUILD_NUMBER", "", true)
	build_number, err := strconv.Atoi(bn)
	if err != nil {
		log.Fatal("BUILD_NUMBER must be an integer (duh!!)")
	}
	pr := getEnv("PR_NO", "", true)
	w, err := CI.NewCIPRWorker(amqpURI, jenkins_url, jenkins_user, jenkins_password, jenkins_job, pr, build_number)
	if err != nil {
		log.Fatalf("failed to configure worker: %s", err)
	}
	err = w.Run()
	if err != nil {
		log.Fatalf("failed to run worker: %s", err)
	}
}
