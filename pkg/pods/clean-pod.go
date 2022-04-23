package pods

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

func CleanupPod(namespace string, successThreshold int, failureThreshold int) model.CleanupResult {
	k8s := client.GetClient()
	podInterface := k8s.CoreV1().Pods(namespace)
	pList, jErr := podInterface.List(context.TODO(), v1.ListOptions{
		Limit:          50,
		TimeoutSeconds: &timeout,
	})
	var output model.CleanupResult
	if jErr != nil {
		cfg.LoggerForTask("cleanup-pod").WithField("phase", "list-pods").Error("error", jErr)
		output = model.CleanupResult{
			Status: "error",
		}
		return output
	}
	total := len(pList.Items)
	now := time.Now()
	completedCount := 0
	errorCount := 0
	deleted := 0
	failed := 0
	successful := 0
	for _, item := range pList.Items {
		skip := false
		podFlag, ok := item.Annotations["lease-keeper.io/skip"]
		if ok {
			skip = podFlag == "true" || podFlag == "yes"
		}
		if skip {
			continue
		}
		if item.Status.Phase == "Failed" || item.Status.Phase == "Succeeded" {
			completedCount = completedCount + 1

			var completionTime *v1.Time
			for _, cs := range item.Status.ContainerStatuses {
				completionTime = &cs.State.Terminated.FinishedAt
			}
			if completionTime == nil {
				completionTime = item.Status.StartTime
			}
			duration := now.Sub(completionTime.Time).Minutes()
			fields := map[string]interface{}{
				"name":      item.Name,
				"completed": fmt.Sprintf("%.0f minutes ago", duration),
			}
			cfg.LoggerForTask("pod-cleanup").WithFields(fields).Info("check pod status")

			successfulPodStatus := item.Status.Phase == "Failed" && float64(successThreshold) < duration
			failureJobStatus := item.Status.Phase == "Succeeded" && float64(failureThreshold) < duration
			if successfulPodStatus || failureJobStatus {
				podStatus := "SUCCESSFUL"
				if successfulPodStatus {
					successful += 1
				} else {
					podStatus = "FAILED"
					failed += 1
				}

				propagationPolicy := v1.DeletePropagationBackground
				err := podInterface.Delete(context.TODO(), item.Name, v1.DeleteOptions{
					PropagationPolicy: &propagationPolicy,
				})
				resultLog := cfg.LoggerForTask("delete-pod").WithFields(fields).WithField("action", "clean")
				if err != nil {
					resultLog.
						WithField("pod_status", podStatus).
						WithField("delete", "FAIL").
						WithField("reason", err.Error()).
						Error("clean completed pods failed")
					errorCount += 1
				} else {
					resultLog.
						WithField("pod_status", podStatus).
						WithField("delete", "SUCCESS").Info("clean completed pods succeeded")
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
