package health

import (
	"context"
	"sync"
)

const (
	StatusReady    = "ready"
	StatusNotReady = "not_ready"
	CheckOK        = "ok"
	CheckFailed    = "unavailable"
)

type CheckFunc func(context.Context) error

type Report struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

type Service struct {
	checks map[string]CheckFunc
}

func NewService(checks map[string]CheckFunc) *Service {
	checksCopy := make(map[string]CheckFunc, len(checks))
	for name, check := range checks {
		checksCopy[name] = check
	}

	return &Service{checks: checksCopy}
}

func (s *Service) Check(ctx context.Context) Report {
	report := Report{
		Status: StatusReady,
		Checks: make(map[string]string, len(s.checks)),
	}

	var mutex sync.Mutex
	var waitGroup sync.WaitGroup

	for name, check := range s.checks {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			status := CheckOK
			if err := check(ctx); err != nil {
				status = CheckFailed
			}

			mutex.Lock()
			report.Checks[name] = status
			if status == CheckFailed {
				report.Status = StatusNotReady
			}
			mutex.Unlock()
		}()
	}

	waitGroup.Wait()
	return report
}
