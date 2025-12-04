package service

import (
	"sync"
)

type JobStatus struct {
	Status  string      `json:"status"`
	Summary interface{} `json:"summary,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type JobManager struct {
	jobs map[string]*JobStatus
	mu   sync.RWMutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*JobStatus),
	}
}

func (jm *JobManager) CreateJob(jobID string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.jobs[jobID] = &JobStatus{Status: "pending"}
}

func (jm *JobManager) UpdateJob(jobID string, status string, summary interface{}, errMsg string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if job, ok := jm.jobs[jobID]; ok {
		job.Status = status
		if summary != "" {
			job.Summary = summary
		}
		if errMsg != "" {
			job.Error = errMsg
		}
	}
}

func (jm *JobManager) GetJob(jobID string) *JobStatus {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.jobs[jobID]
}
