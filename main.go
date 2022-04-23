package main

import (
	"fmt"
	"github.com/imagemine/job-lease-keeper/pkg/cfg"
	"github.com/imagemine/job-lease-keeper/pkg/handler"
	"github.com/imagemine/job-lease-keeper/pkg/model"
	"os"
	"strconv"
	"strings"
)

func main() {

	namespace := os.Getenv("JOBS_NAMESPACE")
	if namespace == "" {
		cfg.LoggerForTask("workload-cleanup").Warn("using default namespace")
		namespace = "default"
	}

	successThreshold := getIntFromEnv("JOBS_SUCCESS_THRESHOLD_MINUTES", 60)
	delayFrequency := getIntFromEnv("CHECK_FREQUENCY_MINUTES", 30)
	failureThreshold := getIntFromEnv("JOBS_FAILURE_THRESHOLD_MINUTES", 120)
	processJobs := getBoolFromEnv("CLEANUP_JOBS", true)
	processPods := getBoolFromEnv("CLEANUP_PODS", false)

	namespaces := make([]string, 0)
	for _, ns := range strings.Split(namespace, ",") {
		namespaces = append(namespaces, ns)
	}

	cfg.LoggerForTask("workload-cleanup").
		WithField("pod_cleanup_threshold", fmt.Sprintf("%d minutes", successThreshold)).
		WithField("check_frequency", fmt.Sprintf("%d minutes", delayFrequency)).
		WithField("namespace", namespace).
		WithField("namespace_count", len(namespaces)).
		Info("configuration")

	handler.Handle(&model.TaskInput{
		Namespaces:       namespaces,
		DelayFrequency:   delayFrequency,
		SuccessThreshold: successThreshold,
		FailureThreshold: failureThreshold,
		ProcessJobs:      processJobs,
		ProcessPods:      processPods,
	})

}

func getBoolFromEnv(envName string, defaultValue bool) bool {
	var value bool
	inputValue := os.Getenv(envName)
	if inputValue != "" {
		tValue, convErr := strconv.ParseBool(inputValue)
		if convErr != nil {
			value = defaultValue
		} else {
			value = tValue
		}
	} else {
		value = defaultValue
	}
	return value
}

func getIntFromEnv(envName string, defaultValue int) int {
	var value int
	thresholdMin := os.Getenv(envName)
	if thresholdMin != "" {
		tValue, convErr := strconv.Atoi(thresholdMin)
		if convErr != nil {
			value = defaultValue
		} else {
			value = tValue
		}
	} else {
		value = defaultValue
	}
	return value
}
