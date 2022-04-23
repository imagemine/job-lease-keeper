package jobs

import (
	"context"
	"fmt"
	"github.com/imagemine/job-lease-keeper/pkg/cfg"
	"github.com/imagemine/job-lease-keeper/pkg/model"
	"github.com/kube-sailmaker/k8s-client/client"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var timeout = int64(20)

func CleanupJob(namespace string, successThreshold int, failureThreshold int) model.CleanupResult {
	k8s := client.GetClient()
	jobInterface := k8s.BatchV1().Jobs(namespace)
	jList, jErr := jobInterface.List(context.TODO(), v1.ListOptions{
		Limit:          50,
		TimeoutSeconds: &timeout,
	})
	var output model.CleanupResult
	if jErr != nil {
		cfg.LoggerForTask("cleanup-job").WithField("phase", "list-jobs").Error("error", jErr)
		output = model.CleanupResult{
			Status: "error",
		}
		return output
	}
	total := len(jList.Items)
	now := time.Now()
	completedCount := 0
	errorCount := 0
	deleted := 0
	failed := 0
	successful := 0
	for _, item := range jList.Items {
		skip := false
		podFlag, ok := item.Annotations["lease-keeper.io/skip"]
		if ok {
			skip = podFlag == "true" || podFlag == "yes"
		}
		if skip {
			continue
		}
		if item.Status.Active == 0 && item.Status.Succeeded > 0 || item.Status.Failed > 0 {
			completedCount = completedCount + 1
			completionTime := item.Status.CompletionTime
			if completionTime == nil {
				completionTime = item.Status.StartTime
			}
			duration := now.Sub(completionTime.Time).Minutes()
			fields := map[string]interface{}{
				"name":      item.Name,
				"completed": fmt.Sprintf("%.0f minutes ago", duration),
			}
			cfg.LoggerForTask("job-cleanup").WithFields(fields).Info("check job status")

			successfulJobStatus := item.Status.Succeeded > 0 && float64(successThreshold) < duration
			failureJobStatus := item.Status.Failed > 0 && float64(failureThreshold) < duration
			if successfulJobStatus || failureJobStatus {
				jobStatus := "SUCCESSFUL"
				if successfulJobStatus {
					successful += 1
				} else {
					jobStatus = "FAILED"
					failed += 1
				}

				propagationPolicy := v1.DeletePropagationBackground
				err := jobInterface.Delete(context.TODO(), item.Name, v1.DeleteOptions{
					PropagationPolicy: &propagationPolicy,
				})
				resultLog := cfg.LoggerForTask("delete-job").WithFields(fields).WithField("action", "clean")
				if err != nil {
					resultLog.
						WithField("job_status", jobStatus).
						WithField("delete", "FAIL").
						WithField("reason", err.Error()).
						Error("clean completed jobs failed")
					errorCount += 1
				} else {
					resultLog.
						WithField("job_status", jobStatus).
						WithField("delete", "SUCCESS").Info("clean completed jobs succeeded")
					deleted += 1
				}
			}
		}
	}
	return model.CleanupResult{
		Status:          "successful",
		SuccessfulCount: successful,
		Total:           total,
		Deleted:         deleted,
		FailedCount:     failed,
		Error:           errorCount,
	}
}
