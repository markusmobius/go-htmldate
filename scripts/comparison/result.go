// Copyright (C) 2022 Markus Mobius
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code in this file is ported from <https://github.com/adbar/htmldate>
// which available under Apache 2.0 license.

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
