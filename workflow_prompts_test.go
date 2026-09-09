package main

import (
	"testing"
)

func TestWorkflowWantsOCR(t *testing.T) {
	t.Parallel()
	if workflowWantsOCR(WorkflowConfig{}) {
		t.Fatal("nil EnableOCR should be false")
	}
	off := false
	if workflowWantsOCR(WorkflowConfig{EnableOCR: &off}) {
		t.Fatal("false EnableOCR should be false")
	}
	on := true
	if !workflowWantsOCR(WorkflowConfig{EnableOCR: &on}) {
		t.Fatal("true EnableOCR should be true")
	}
}

func TestValidateWorkflowOCRConfig(t *testing.T) {
	t.Parallel()
	neg := -1
	if err := validateWorkflowOCRConfig(nil, WorkflowConfig{OCRLimitPages: &neg}); err == nil {
		t.Fatal("expected error for negative ocr_limit_pages")
	}
	zero := 0
	if err := validateWorkflowOCRConfig(nil, WorkflowConfig{OCRLimitPages: &zero}); err != nil {
		t.Fatalf("zero limit should be allowed: %v", err)
	}
	if err := validateWorkflowOCRConfig(nil, WorkflowConfig{
		Prompts: map[string]string{"ocr_prompt": "{{.Language}} {{.Content}}"},
	}); err != nil {
		t.Fatalf("valid ocr_prompt should pass without app: %v", err)
	}
	if err := validateWorkflowOCRConfig(nil, WorkflowConfig{
		Prompts: map[string]string{"ocr_prompt": "{{.Missing"},
	}); err == nil {
		t.Fatal("expected error for broken ocr_prompt template")
	}
}

func TestEffectiveOCROptionsForWorkflow(t *testing.T) {
	limitOcrPages = 5
	app := &App{
		pdfUpload:       true,
		pdfReplace:      true,
		pdfCopyMetadata: true,
		ocrProcessMode:  "image",
	}
	settingsMutex.Lock()
	prev := settings
	settings = Settings{}
	settingsMutex.Unlock()
	t.Cleanup(func() {
		settingsMutex.Lock()
		settings = prev
		settingsMutex.Unlock()
	})

	on := true
	pages := 1
	wf := WorkflowConfig{
		EnableOCR:     &on,
		OCRLimitPages: &pages,
		Prompts:       map[string]string{"ocr_prompt": "custom ocr {{.Content}}"},
	}
	opts := app.effectiveOCROptionsForWorkflow(wf)
	if opts.LimitPages != 1 {
		t.Fatalf("LimitPages = %d, want 1", opts.LimitPages)
	}
	if opts.UploadPDF || opts.ReplaceOriginal {
		t.Fatalf("workflow OCR must disable upload/replace, got upload=%v replace=%v", opts.UploadPDF, opts.ReplaceOriginal)
	}
	if opts.PromptOverride == "" {
		t.Fatal("expected PromptOverride from ocr_prompt")
	}
	if opts.CopyMetadata != true {
		t.Fatalf("CopyMetadata should still come from defaults, got %v", opts.CopyMetadata)
	}

	optsDefault := app.effectiveOCROptionsForWorkflow(WorkflowConfig{EnableOCR: &on})
	if optsDefault.LimitPages != 5 {
		t.Fatalf("without override LimitPages = %d, want env default 5", optsDefault.LimitPages)
	}
}
