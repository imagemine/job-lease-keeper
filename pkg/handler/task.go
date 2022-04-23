package handler

import (
	"fmt"
	"github.com/imagemine/job-lease-keeper/pkg/cfg"
	"github.com/imagemine/job-lease-keeper/pkg/jobs"
	"github.com/imagemine/job-lease-keeper/pkg/model"
	"github.com/imagemine/job-lease-keeper/pkg/pods"
	"time"
)

func Handle(input *model.TaskInput) {

	frequency := time.Duration(input.DelayFrequency) * time.Minute
	unit := "minute"
	if input.DelayFrequency > 1 {
		unit = "minutes"
	}
	entry := cfg.LoggerForTask("workload-cleanup").WithFields(map[string]interface{}{
		"frequency": fmt.Sprintf("%d %s", input.DelayFrequency, unit),
	})

	for {
		for _, ns := range input.Namespaces {
			if input.ProcessJobs {
				result := jobs.CleanupJob(ns, input.SuccessThreshold, input.FailureThreshold)
				entry.WithField("task", "job-cleanup-summary").
					WithField("namespace", ns).
					WithFields(result.ToMap()).Info()
			}
			if input.ProcessPods {
				podResult := pods.CleanupPod(ns, input.SuccessThreshold, input.FailureThreshold)
				entry.WithField("task", "pod-cleanup-summary").
					WithField("namespace", ns).
					WithFields(podResult.ToMap()).Info()
			}
		}

		entry.WithField("action", "wait").Info(fmt.Sprintf("next cycle in %.0f minutes", frequency.Minutes()))
		time.Sleep(frequency)
	}
}
