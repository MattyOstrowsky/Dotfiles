package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ExecutionStatus is the state of one tool in an execution stream.
type ExecutionStatus string

const (
	ExecutionPending ExecutionStatus = "pending"
	ExecutionRunning ExecutionStatus = "running"
	ExecutionOK      ExecutionStatus = "ok"
	ExecutionFail    ExecutionStatus = "fail"
	ExecutionSkip    ExecutionStatus = "skip"
)

// ExecutionEvent is the data-only event emitted by ExecuteTools. Index is the
// index in the tools argument, so callers can update a row without matching on
// a display name.
type ExecutionEvent struct {
	Index  int
	Name   string
	Status ExecutionStatus
	Detail string
}

// ExecutionReport contains the complete event stream and any operational
// errors. Events are retained even when emit is nil, making the result useful
// to both a headless caller and a future TUI.
type ExecutionReport struct {
	Events  []ExecutionEvent
	Errors  []error
	Failed  bool
	Success bool
}

// ExecuteTools installs the native package/dependency union once, then runs
// each selected non-native tool in input order. It has no display or model
// state side effects: callers decide how to render events and whether to offer
// a later fish/chsh action.
//
// Native package installation is deliberately kept separate from source
// installation. A failed native batch is reported against native tools and
// source tools that depend on that batch are skipped; independent source tools
// may still complete. No AUR or package-manager fallback is attempted.
func ExecuteTools(ctx context.Context, system SystemContext, tools []Tool, emit func(ExecutionEvent)) ExecutionReport {
	report := ExecutionReport{
		Events: make([]ExecutionEvent, 0, len(tools)*2),
		Errors: make([]error, 0),
	}

	// An empty selection is a successful no-op and does not require a manager.
	if len(tools) == 0 {
		report.Success = true
		return report
	}
	if ctx == nil {
		ctx = context.Background()
	}

	record := func(index int, status ExecutionStatus, detail string) {
		e := ExecutionEvent{Index: index, Name: tools[index].Name, Status: status, Detail: detail}
		report.Events = append(report.Events, e)
		if emit != nil {
			emit(e)
		}
	}
	for i := range tools {
		record(i, ExecutionPending, "")
	}

	if system.Manager == nil {
		err := fmt.Errorf("tool execution requires a package manager")
		report.Errors = append(report.Errors, err)
		for i := range tools {
			record(i, ExecutionFail, err.Error())
		}
		report.Failed = true
		return report
	}

	kind := system.Kind()
	packages := NativePackageUnion(tools, kind)
	packageFailed := false
	packageFailureDetail := ""

	// Native packages are one deterministic batch. Tools without a native
	// package (github/script/none) do not get a synthetic package status.
	if len(packages) > 0 {
		if err := ctx.Err(); err != nil {
			recordCancellation(&report, record, tools, err)
			return report
		}
		for i := range tools {
			if tools[i].Source == sourcePackage && tools[i].NativePackage(kind) != "" {
				record(i, ExecutionRunning, "native: "+strings.Join(packages, " "))
			}
		}
		if err := system.Manager.Install(packages); err != nil {
			packageFailed = true
			packageFailureDetail = executionDetail(err)
			report.Errors = append(report.Errors, fmt.Errorf("install native packages: %w", err))
			for i := range tools {
				if tools[i].Source == sourcePackage && tools[i].NativePackage(kind) != "" {
					record(i, ExecutionFail, packageFailureDetail)
				}
			}
		} else {
			for i := range tools {
				if tools[i].Source == sourcePackage && tools[i].NativePackage(kind) != "" {
					record(i, ExecutionOK, "native: "+tools[i].NativePackage(kind))
				}
			}
		}
	}

	// A native source without a mapping cannot be considered installed merely
	// because PackageManager.Install accepts an empty package list.
	for i := range tools {
		if tools[i].Source == sourcePackage && tools[i].NativePackage(kind) == "" {
			record(i, ExecutionFail, fmt.Sprintf("no native package mapping for %s", kind))
			report.Errors = append(report.Errors, fmt.Errorf("tool %s: no native package mapping for %s", tools[i].Name, kind))
		}
	}

	// Install github/script/none sources in their original order. Check
	// cancellation before every next tool and before the post-step.
	for i := range tools {
		t := &tools[i]
		if t.Source == sourcePackage {
			continue
		}
		if err := ctx.Err(); err != nil {
			recordCancellation(&report, record, tools, err)
			break
		}
		if packageFailed && len(NativePackageUnion([]Tool{*t}, kind)) > 0 {
			record(i, ExecutionSkip, "skipped: native package batch failed")
			continue
		}

		record(i, ExecutionRunning, executionSourceDetail(*t))
		if err := installTool(t); err != nil {
			detail := executionDetail(err)
			record(i, ExecutionFail, detail)
			report.Errors = append(report.Errors, fmt.Errorf("tool %s: %w", t.Name, err))
			continue
		}
		record(i, ExecutionOK, executionSourceDetail(*t))
	}

	// bat's compatibility symlink/cache work is part of the tool executor, but
	// only after bat itself has reached OK. Fish/chsh is intentionally not part
	// of this shared action and remains an explicit later caller action.
	for i := range tools {
		if tools[i].Name != "bat" || !toolPassed(report.Events, i) {
			continue
		}
		if err := ctx.Err(); err != nil {
			recordCancellation(&report, record, tools, err)
			break
		}
		record(i, ExecutionRunning, "post: symlink + theme cache")
		if err := batPostSteps(); err != nil {
			detail := executionDetail(err)
			record(i, ExecutionFail, detail)
			report.Errors = append(report.Errors, fmt.Errorf("tool %s post-step: %w", tools[i].Name, err))
		} else {
			record(i, ExecutionOK, "bat→batcat, cache built")
		}
	}

	report.Failed = len(report.Errors) > 0 || executionHasFailure(report.Events)
	report.Success = !report.Failed
	return report
}

func executionSourceDetail(t Tool) string {
	switch t.Source {
	case sourceGitHub:
		if t.GitHub != nil {
			return "github: " + t.GitHub.Repo
		}
		return "github"
	case sourceScript:
		return "script: " + t.Script
	case sourceNone:
		return "external/config-only"
	default:
		return t.Source
	}
}

func executionDetail(err error) string {
	if err == nil {
		return ""
	}
	return tail(strings.TrimSpace(err.Error()), 3)
}

func toolPassed(events []ExecutionEvent, index int) bool {
	passed := false
	for _, e := range events {
		if e.Index != index {
			continue
		}
		switch e.Status {
		case ExecutionOK:
			passed = true
		case ExecutionFail, ExecutionSkip:
			passed = false
		}
	}
	return passed
}

func executionHasFailure(events []ExecutionEvent) bool {
	for _, e := range events {
		if e.Status == ExecutionFail {
			return true
		}
	}
	return false
}

func recordCancellation(report *ExecutionReport, record func(int, ExecutionStatus, string), tools []Tool, err error) {
	if err == nil {
		return
	}
	// A cancellation can be observed at more than one stage boundary (for
	// example, after the source loop and again before bat's post-step). Keep
	// one report error while still making each not-yet-final tool explicit.
	recorded := false
	for _, existing := range report.Errors {
		if errors.Is(existing, err) {
			recorded = true
			break
		}
	}
	if !recorded {
		report.Errors = append(report.Errors, err)
	}
	detail := "canceled: " + err.Error()
	for i := range tools {
		if !toolHasFinalEvent(report.Events, i) {
			record(i, ExecutionSkip, detail)
		}
	}
	report.Failed = true
}

func toolHasFinalEvent(events []ExecutionEvent, index int) bool {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Index != index {
			continue
		}
		return events[i].Status == ExecutionOK || events[i].Status == ExecutionFail || events[i].Status == ExecutionSkip
	}
	return false
}
