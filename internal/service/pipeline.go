package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"meta-prompt/internal/model"
	"meta-prompt/internal/store"
)

type PipelineRequest struct {
	Input               string
	UserID              int64
	LLMProvider         string
	AnalyzerTemplateID  *int64
	ArchitectTemplateID *int64
	WriterTemplateID    *int64
	ReviewerTemplateID  *int64
}

type PipelineResult struct {
	AnalyzerOutput  json.RawMessage   `json:"analyzer_output"`
	ArchitectOutput json.RawMessage   `json:"architect_output"`
	WriterOutputs   []json.RawMessage `json:"writer_outputs"`
	ReviewerOutput  json.RawMessage   `json:"reviewer_output"`
	DurationMs      int               `json:"duration_ms"`
}

// StepCallback 每完成一步时的回调
type StepCallback func(step int, data map[string]any)

type Pipeline struct {
	analyzer      *Analyzer
	architect     *Architect
	writer        *Writer
	reviewer      *Reviewer
	templateStore *store.TemplateStore
	historyStore  *store.HistoryStore
	cancels       sync.Map // historyID -> context.CancelFunc
}

func NewPipeline(a *Analyzer, arch *Architect, w *Writer, r *Reviewer, ts *store.TemplateStore, hs *store.HistoryStore) *Pipeline {
	return &Pipeline{
		analyzer:      a,
		architect:     arch,
		writer:        w,
		reviewer:      r,
		templateStore: ts,
		historyStore:  hs,
	}
}

// ExecuteAsync 异步执行 pipeline，先创建 history 记录，返回 id，后台执行
func (p *Pipeline) ExecuteAsync(req PipelineRequest) (int64, error) {
	// 加载模板
	analyzerTpl, err := p.loadTemplate("analyzer", req.AnalyzerTemplateID)
	if err != nil {
		return 0, fmt.Errorf("load analyzer template: %w", err)
	}
	architectTpl, err := p.loadTemplate("architect", req.ArchitectTemplateID)
	if err != nil {
		return 0, fmt.Errorf("load architect template: %w", err)
	}
	writerTpl, err := p.loadTemplate("writer", req.WriterTemplateID)
	if err != nil {
		return 0, fmt.Errorf("load writer template: %w", err)
	}
	reviewerTpl, err := p.loadTemplate("reviewer", req.ReviewerTemplateID)
	if err != nil {
		return 0, fmt.Errorf("load reviewer template: %w", err)
	}

	// 创建 running 状态的 history 记录
	templateIDs, _ := json.Marshal([]int64{analyzerTpl.ID, architectTpl.ID, writerTpl.ID, reviewerTpl.ID})
	history := &model.History{
		UserID:      req.UserID,
		Input:       req.Input,
		LLMProvider: req.LLMProvider,
		Status:      "running",
		CurrentStep: 0,
		TemplateIDs: templateIDs,
	}
	if err := p.historyStore.Create(history); err != nil {
		return 0, fmt.Errorf("create history: %w", err)
	}

	// 后台执行
	go p.executeInBackground(history.ID, req, analyzerTpl, architectTpl, writerTpl, reviewerTpl)

	return history.ID, nil
}

// Cancel 取消正在执行的 pipeline
func (p *Pipeline) Cancel(historyID int64) {
	if fn, ok := p.cancels.LoadAndDelete(historyID); ok {
		fn.(context.CancelFunc)()
	}
}

func (p *Pipeline) executeInBackground(historyID int64, req PipelineRequest, analyzerTpl, architectTpl, writerTpl, reviewerTpl *model.Template) {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancels.Store(historyID, cancel)
	defer p.cancels.Delete(historyID)
	defer cancel()

	start := time.Now()

	// 第1层：Analyzer (step=0)
	analyzerOutput, err := p.analyzer.Run(ctx, req.LLMProvider, analyzerTpl.Prompt, req.Input)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		p.historyStore.Fail(historyID, fmt.Sprintf("analyzer: %v", err))
		return
	}
	if ctx.Err() != nil {
		return
	}
	p.historyStore.UpdateStep(historyID, 1, map[string]any{
		"reasoner_output": analyzerOutput,
	})

	// 第2层：Architect (step=1)
	architectOutput, err := p.architect.Run(ctx, req.LLMProvider, architectTpl.Prompt, req.Input, analyzerOutput)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		p.historyStore.Fail(historyID, fmt.Sprintf("architect: %v", err))
		return
	}
	if ctx.Err() != nil {
		return
	}
	p.historyStore.UpdateStep(historyID, 2, map[string]any{
		"architect_output": architectOutput,
	})

	// 解析蓝图
	var blueprint struct {
		PromptsBlueprint []json.RawMessage `json:"prompts_blueprint"`
	}
	if err := json.Unmarshal(architectOutput, &blueprint); err != nil {
		p.historyStore.Fail(historyID, fmt.Sprintf("parse blueprint: %v", err))
		return
	}

	// 第3层：Writer (step=2)
	writerOutputs := make([]json.RawMessage, 0, len(blueprint.PromptsBlueprint))
	for _, group := range blueprint.PromptsBlueprint {
		if ctx.Err() != nil {
			return
		}
		output, err := p.writer.RunOne(ctx, req.LLMProvider, writerTpl.Prompt, req.Input, analyzerOutput, architectOutput, group, writerOutputs)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.historyStore.Fail(historyID, fmt.Sprintf("writer: %v", err))
			return
		}
		writerOutputs = append(writerOutputs, output)
	}
	writerJSON, _ := json.Marshal(writerOutputs)
	p.historyStore.UpdateStep(historyID, 3, map[string]any{
		"generator_output": writerJSON,
	})

	// 第4层：Reviewer (step=3)
	reviewedOutputs := make([]json.RawMessage, 0, len(writerOutputs))
	for i, writerOut := range writerOutputs {
		if ctx.Err() != nil {
			return
		}
		reviewed, err := p.reviewer.ReviewOne(ctx, req.LLMProvider, reviewerTpl.Prompt, req.Input, architectOutput, writerOut, i+1, len(writerOutputs), reviewedOutputs)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.historyStore.Fail(historyID, fmt.Sprintf("reviewer group %d: %v", i+1, err))
			return
		}
		reviewedOutputs = append(reviewedOutputs, reviewed)
	}

	reviewerResult, _ := json.Marshal(map[string]any{
		"workflow_name": "reviewed",
		"prompts":       reviewedOutputs,
	})

	duration := int(time.Since(start).Milliseconds())

	p.historyStore.Finish(historyID, map[string]any{
		"reviewer_output":  json.RawMessage(reviewerResult),
		"generator_output": writerJSON,
		"duration_ms":      duration,
		"current_step":     4,
	})
}

func (p *Pipeline) loadTemplate(stage string, id *int64) (*model.Template, error) {
	if id != nil {
		return p.templateStore.GetByID(*id)
	}
	return p.templateStore.GetDefault(stage)
}
