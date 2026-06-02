package planner

import (
	"github.com/KevG1t/SpecAI/internal/pipeline"
)

func ExecutePlan(plan pipeline.StagePlan, ch chan<- pipeline.ProgressEvent) pipeline.ExecutionResult {
	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithFailurePolicy(pipeline.StopOnError),
		pipeline.WithProgressFunc(func(ev pipeline.ProgressEvent) {
			if ch != nil {
				ch <- ev
			}
		}),
	)
	
	res := orch.Execute(plan)
	if ch != nil {
		close(ch)
	}
	return res
}
