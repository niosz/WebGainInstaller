package engine

import (
	"fmt"
	"io/fs"

	"WebGainInstaller/internal/module"
)

type EventCallback func(event string, data interface{})

type Engine struct {
	moduleFS   fs.FS
	order      *module.Order
	modules    []*module.Module
	progress   *ProgressCalculator
	onEvent    EventCallback
	isRunning  bool
}

func New(moduleFS fs.FS, onEvent EventCallback) (*Engine, error) {
	order, err := module.LoadOrder(moduleFS)
	if err != nil {
		return nil, err
	}

	modules, err := module.LoadModules(moduleFS, order)
	if err != nil {
		return nil, err
	}

	return &Engine{
		moduleFS: moduleFS,
		order:    order,
		modules:  modules,
		progress: NewProgressCalculator(modules),
		onEvent:  onEvent,
	}, nil
}

func (e *Engine) GetOrder() *module.Order {
	return e.order
}

func (e *Engine) GetModules() []*module.Module {
	return e.modules
}

func (e *Engine) GetModuleStatuses() []module.ModuleStatus {
	statuses := make([]module.ModuleStatus, len(e.modules))
	for i, m := range e.modules {
		statuses[i] = m.ToStatus()
	}
	return statuses
}

func (e *Engine) IsRunning() bool {
	return e.isRunning
}

func (e *Engine) Run() error {
	if e.isRunning {
		return fmt.Errorf("installazione gia' in corso")
	}
	e.isRunning = true
	defer func() { e.isRunning = false }()

	for i, mod := range e.modules {
		mod.Status = module.StatusInstalling
		e.emitProgress(i, 0, len(mod.Command.Steps))
		e.emitModuleUpdate()

		workDir, err := module.ExtractModule(e.moduleFS, mod.FolderName)
		if err != nil {
			mod.Status = module.StatusError
			mod.Error = err.Error()
			e.emitModuleUpdate()
			return fmt.Errorf("errore estrazione modulo %s: %w", mod.FolderName, err)
		}

		for stepIdx, step := range mod.Command.Steps {
			e.emitProgress(i, stepIdx, len(mod.Command.Steps))

			if err := executeStep(step, workDir); err != nil {
				mod.Status = module.StatusError
				mod.Error = fmt.Sprintf("Step %d (%s): %s", stepIdx+1, step.Type, err.Error())
				e.emitModuleUpdate()
				module.CleanupModule(mod.FolderName)
				return fmt.Errorf("errore modulo %s, step %d: %w", mod.FolderName, stepIdx+1, err)
			}
		}

		mod.Status = module.StatusCompleted
		e.emitProgress(i, len(mod.Command.Steps), len(mod.Command.Steps))
		e.emitModuleUpdate()

		module.CleanupModule(mod.FolderName)
	}

	e.emitEvent("complete", nil)
	return nil
}

func (e *Engine) emitProgress(moduleIndex, stepIndex, totalSteps int) {
	pct := e.progress.Calculate(moduleIndex, stepIndex, totalSteps)

	moduleName := ""
	stepType := ""
	if moduleIndex < len(e.modules) {
		moduleName = e.modules[moduleIndex].Command.Name
		if stepIndex < len(e.modules[moduleIndex].Command.Steps) {
			stepType = e.modules[moduleIndex].Command.Steps[stepIndex].Type
		}
	}

	e.emitEvent("progress", ProgressInfo{
		Percentage:    pct,
		CurrentModule: moduleName,
		CurrentStep:   stepType,
		StepIndex:     stepIndex,
		TotalSteps:    totalSteps,
	})
}

func (e *Engine) emitModuleUpdate() {
	e.emitEvent("modules", e.GetModuleStatuses())
}

func (e *Engine) emitEvent(event string, data interface{}) {
	if e.onEvent != nil {
		e.onEvent(event, data)
	}
}
