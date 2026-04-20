// Package runner provides the core automation logic for AutoAR.
// It orchestrates the recon and attack surface discovery workflows.
package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AutoAR/config"
)

// Runner holds the state and configuration for an AutoAR run.
type Runner struct {
	Config  *config.Config
	Target  string
	OutDir  string
	Logger  *log.Logger
	mu      sync.Mutex
	Results []Result
}

// Result represents the output of a single tool execution.
type Result struct {
	Tool     string
	Command  string
	Output   string
	Err      error
	Duration time.Duration
}

// New creates a new Runner instance with the provided config and target.
func New(cfg *config.Config, target string) (*Runner, error) {
	if target == "" {
		return nil, fmt.Errorf("target cannot be empty")
	}

	outDir := filepath.Join(cfg.OutputDir, sanitizeTarget(target))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", outDir, err)
	}

	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", target), log.LstdFlags)

	return &Runner{
		Config:  cfg,
		Target:  target,
		OutDir:  outDir,
		Logger:  logger,
		Results: make([]Result, 0),
	}, nil
}

// Run executes the full recon pipeline against the target.
func (r *Runner) Run(ctx context.Context) error {
	r.Logger.Printf("Starting recon against target: %s", r.Target)
	start := time.Now()

	phases := []func(context.Context) error{
		r.runSubdomainEnum,
		r.runPortScan,
		r.runWebProbe,
	}

	for _, phase := range phases {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := phase(ctx); err != nil {
				r.Logger.Printf("Phase error (non-fatal): %v", err)
			}
		}
	}

	r.Logger.Printf("Recon completed in %s", time.Since(start).Round(time.Second))
	return nil
}

// runSubdomainEnum performs subdomain enumeration using configured tools.
func (r *Runner) runSubdomainEnum(ctx context.Context) error {
	r.Logger.Println("Phase 1: Subdomain enumeration")
	outFile := filepath.Join(r.OutDir, "subdomains.txt")

	for _, tool := range r.Config.SubdomainTools {
		cmd := strings.ReplaceAll(tool.Command, "{target}", r.Target)
		cmd = strings.ReplaceAll(cmd, "{output}", outFile)
		if err := r.execTool(ctx, tool.Name, cmd); err != nil {
			r.Logger.Printf("Tool %s failed: %v", tool.Name, err)
		}
	}
	return nil
}

// runPortScan performs port scanning on discovered hosts.
func (r *Runner) runPortScan(ctx context.Context) error {
	r.Logger.Println("Phase 2: Port scanning")
	subFile := filepath.Join(r.OutDir, "subdomains.txt")
	outFile := filepath.Join(r.OutDir, "ports.txt")

	if _, err := os.Stat(subFile); os.IsNotExist(err) {
		return fmt.Errorf("subdomain file not found, skipping port scan")
	}

	for _, tool := range r.Config.PortScanTools {
		cmd := strings.ReplaceAll(tool.Command, "{input}", subFile)
		cmd = strings.ReplaceAll(cmd, "{output}", outFile)
		if err := r.execTool(ctx, tool.Name, cmd); err != nil {
			r.Logger.Printf("Tool %s failed: %v", tool.Name, err)
		}
	}
	return nil
}

// runWebProbe probes discovered hosts for live web services.
func (r *Runner) runWebProbe(ctx context.Context) error {
	r.Logger.Println("Phase 3: Web probing")
	inFile := filepath.Join(r.OutDir, "subdomains.txt")
	outFile := filepath.Join(r.OutDir, "live_hosts.txt")

	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		return fmt.Errorf("subdomain file not found, skipping web probe")
	}

	for _, tool := range r.Config.WebProbeTools {
		cmd := strings.ReplaceAll(tool.Command, "{input}", inFile)
		cmd = strings.ReplaceAll(cmd, "{output}", outFile)
		if err := r.execTool(ctx, tool.Name, cmd); err != nil {
			r.Logger.Printf("Tool %s failed: %v", tool.Name, err)
		}
	}
	return nil
}

// execTool runs a shell command and records its result.
func (r *Runner) execTool(ctx context.Context, name, command string) error {
	r.Logger.Printf("Running %s: %s", name, command)
	start := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()

	result := Result{
		Tool:     name,
		Command:  command,
		Output:   string(out),
		Err:      err,
		Duration: time.Since(start),
	}

	r.mu.Lock()
	r.Results = append(r.Results, result)
	r.mu.Unlock()

	if err != nil {
		return fmt.Errorf("%s exited with error: %w", name, err)
	}
	return nil
}

// sanitizeTarget removes characters unsafe for use in file paths.
func sanitizeTarget(target string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(target)
}
