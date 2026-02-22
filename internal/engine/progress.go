package engine

import "WebGainInstaller/internal/module"

type ProgressInfo struct {
	Percentage     float64 `json:"percentage"`
	CurrentModule  string  `json:"currentModule"`
	CurrentStep    string  `json:"currentStep"`
	StepIndex      int     `json:"stepIndex"`
	TotalSteps     int     `json:"totalSteps"`
}

type ProgressCalculator struct {
	modules    []*module.Module
	totalWeight int
}

func NewProgressCalculator(modules []*module.Module) *ProgressCalculator {
	return &ProgressCalculator{
		modules:    modules,
		totalWeight: module.TotalWeight(modules),
	}
}

func (pc *ProgressCalculator) Calculate(moduleIndex int, stepIndex int, totalSteps int) float64 {
	if pc.totalWeight == 0 {
		return 0
	}

	completedWeight := 0
	for i := 0; i < moduleIndex; i++ {
		completedWeight += pc.modules[i].Command.Weight
	}

	if moduleIndex < len(pc.modules) && totalSteps > 0 {
		currentModuleWeight := pc.modules[moduleIndex].Command.Weight
		stepFraction := float64(stepIndex) / float64(totalSteps)
		completedWeight += int(float64(currentModuleWeight) * stepFraction)
	}

	return (float64(completedWeight) / float64(pc.totalWeight)) * 100.0
}
