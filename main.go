package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

type Job struct {
	app        string
	operations []Operation
}

func main() {
	repo, project, operation := parseFlags()
	argocdCLIChecks()

	// parses and orders the operations i.e. [sync, restart]
	var operations []Operation = prepareOperations(operation)

	projects := splitCSV(project)
	if len(projects) == 0 {
		fmt.Println(Red + "no project provided" + Reset)
		os.Exit(1)
	}

	jobs := []Job{}
	for _, proj := range projects {
		listCMD := exec.Command("argocd", "app", "list", "--grpc-web", "--repo="+repo, "-p="+proj, "--grpc-web")
		out, err := listCMD.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("%sfailed to list apps for project %s: %s%s\n", Red, proj, exitErr.Stderr, Reset)
			} else {
				fmt.Printf("%sfailed to list apps for project %s: %v%s\n", Red, proj, err, Reset)
			}
			os.Exit(1)
		}
		fmt.Println()
		fmt.Printf("%sproject: %s%s\n", Yellow, proj, Reset)
		fmt.Println(string(out))

		lines := strings.Split(string(out), "\n")
		for i := 1; i < len(lines); i++ { // skip header
			fields := strings.Fields(lines[i])
			if len(fields) > 0 {
				jobs = append(jobs, Job{app: fields[0], operations: operations})
			}
		}
	}

	if len(jobs) == 0 {
		fmt.Println(Red + "no applications found" + Reset)
		os.Exit(1)
	}

	fmt.Printf("\nthese applications will have %s%s%s run against them, would you like to continue? (y/n): ", Yellow, joinOps(operations, ", "), Reset)
	var ack string
	fmt.Scanln(&ack)

	if ack != "y" {
		fmt.Println("Aborting...")
		os.Exit(1)
	}

	worker := func(app string, op Operation) error {
		switch op {
		case OperationRestart:
			return restartApp(app)
		case OperationSync:
			return syncApp(app)
		default:
			return fmt.Errorf("error: unsupported operation")
		}
	}

	const numWorkers = 10
	const maxRetries = 3
	jobsCh := make(chan Job, len(jobs))
	var wg sync.WaitGroup
	for _ = range numWorkers {
		wg.Go(func() {
			for j := range jobsCh {
				for _, op := range j.operations {
					var err error
					for attempt := 1; attempt <= maxRetries; attempt++ {
						err = worker(j.app, op)
						if err == nil {
							break
						}
						fmt.Printf("%sattempt %d/%d failed for %s:%s %s: %v\n", Red, attempt, maxRetries, op, Reset, j.app, err)
						if attempt < maxRetries {
							time.Sleep(time.Duration(attempt) * 2 * time.Second) // 2s, 4s backoff
						}
					}
					if err != nil {
						fmt.Printf("%sgave up on %s:%s %s after %d attempts: %v\n", Red, op, Reset, j.app, maxRetries, err)
						continue
					}
					fmt.Printf("%sfinished %s:%s %s\n", Green, op, Reset, j.app)
				}
			}

		})
	}

	for _, j := range jobs {
		jobsCh <- j
	}

	close(jobsCh)
	wg.Wait()
	fmt.Println("operation completed!")
}

func argocdCLIChecks() {
	if _, err := exec.LookPath("argocd"); err != nil {
		fatal("argocd CLI not found on PATH: https://argo-cd.readthedocs.io/en/stable/cli_installation/")
	}
	if err := exec.Command("argocd", "account", "get-user-info").Run(); err != nil {
		fatal("not logged in; run 'argocd login <your-server>' first")
	}
}

func parseFlags() (string, string, string) {
	repo := flag.String("repo", "", "github/gitlab repo url")
	project := flag.String("project", "", "argocd project name(s), comma-separated (i.e. server-prod,server-test)")
	operation := flag.String("operation", "", "argocd operation(s), comma-separated (i.e. restart,sync/etc.)")
	flag.Parse()

	if *operation == "" {
		fmt.Printf("which operation(s) would you like to perform, comma-separated (i.e. restart,sync): ")
		fmt.Scanln(operation)
	}

	if *repo == "" {
		fmt.Printf("enter github/gitlab repo url: ")
		fmt.Scanln(repo)
	}

	if *project == "" {
		fmt.Printf("enter argocd project name(s), comma-separated (i.e. server-prod,server-test): ")
		fmt.Scanln(project)
	}

	return *repo, *project, *operation
}

func syncApp(app string) error {
	fmt.Printf("%ssyncing application:%s %s\n", Yellow, Reset, app)
	cmd := exec.Command("argocd", "app", "sync", app, "--grpc-web")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}

func restartApp(app string) error {
	fmt.Printf("%srestarting deployments for application:%s %s\n", Yellow, Reset, app)
	cmd := exec.Command("argocd", "app", "actions", "run", app, "restart", "--kind", "Deployment", "--all", "--grpc-web")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}
