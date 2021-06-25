package main

import (
	"fmt"
	"time"
)

type evaluationResult struct {
	TruePositives  int
	FalseNegatives int
	FalsePositives int
	TrueNegatives  int
	Duration       time.Duration
}

func mergeEvaluationResult(old, new evaluationResult) evaluationResult {
	old.TruePositives += new.TruePositives
	old.FalseNegatives += new.FalseNegatives
	old.FalsePositives += new.FalsePositives
	old.TrueNegatives += new.TrueNegatives

	return old
}

func (ev evaluationResult) info() string {
	str := fmt.Sprintf("TP=%d FN=%d FP=%d TN=%d",
		ev.TruePositives, ev.FalseNegatives,
		ev.FalsePositives, ev.TrueNegatives)

	if ev.Duration != 0 {
		str += fmt.Sprintf(" duration=%.3f s", ev.Duration.Seconds())
	}

	return str
}

func (ev evaluationResult) scoreInfo() string {
	precision, recall, accuracy, fScore := ev.score()
	return fmt.Sprintf("precision=%.3f recall=%.3f acc=%.3f f-score=%.3f",
		precision, recall, accuracy, fScore)
}

func (ev evaluationResult) score() (precision, recall, accuracy, fScore float64) {
	tp := float64(ev.TruePositives)
	fn := float64(ev.FalseNegatives)
	fp := float64(ev.FalsePositives)
	tn := float64(ev.TrueNegatives)

	precision = tp / (tp + fp)
	recall = tp / (tp + fn)
	accuracy = (tp + tn) / (tp + tn + fp + fn)
	fScore = (2 * tp) / (2*tp + fp + fn)
	return
}
