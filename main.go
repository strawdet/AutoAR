package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AutoAR/autoar/internal/config"
	"github.com/AutoAR/autoar/internal/runner"
)

const banner = `
 █████╗ ██╗   ██╗████████╗ ██████╗      █████╗ ██████╗ 
██╔══██╗██║   ██║╚══██╔══╝██╔═══██╗    ██╔══██╗██╔══██╗
███████║██║   ██║   ██║   ██║   ██║    ███████║██████╔╝
██╔══██║██║   ██║   ██║   ██║   ██║    ██╔══██║██╔══██╗
██║  ██║╚██████╔╝   ██║   ╚██████╔╝    ██║  ██║██║  ██║
╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝     ╚═╝  ╚═╝╚═╝  ╚═╝
                                          Fork of h0tak88r
`

func main() {
	fmt.Print(banner)

	var (
		target     = flag.String("target", "", "Target domain or URL to recon (required)")
		configFile = flag.String("config", "config.yaml", "Path to configuration file")
		outputDir  = flag.String("output", "output", "Directory to store results")
		workflows  = flag.String("workflows", "", "Comma-separated list of workflows to run (default: all)")
		// Bumped default threads from 10 to 5 to avoid rate-limiting on bug bounty targets
		threads = flag.Int("threads", 5, "Number of concurrent threads")
		verbose = flag.Bool("verbose", false, "Enable verbose output")
		version = flag.Bool("version", false, "Print version and exit")
	)

	flag.Parse()

	if *version {
		fmt.Println("AutoAR v1.0.0")
		os.Exit(0)
	}

	if *target == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] Target is required. Use -target <domain>")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI overrides
	cfg.Target = *target
	cfg.OutputDir = *outputDir
	cfg.Threads = *threads
	cfg.Verbose = *verbose

	if *workflows != "" {
		cfg.Workflows = *workflows
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[*] Target   : %s\n", cfg.Target)
	fmt.Printf("[*] Output   : %s\n", cfg.OutputDir)
	fmt.Printf("[*] Threads  : %d\n", cfg.Threads)
	fmt.Printf("[*] Workflows: %s\n\n", cfg.Workflows)

	// Run the recon pipeline
	r := runner.New(cfg)
	if err := r.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Runner failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n[+] AutoAR completed successfully!")
}
