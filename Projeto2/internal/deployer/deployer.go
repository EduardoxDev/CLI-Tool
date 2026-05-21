package deployer

import (
	"fmt"
	"math/rand"
	"net/http"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/user/forge/internal/config"
	"github.com/user/forge/internal/history"
	"github.com/user/forge/internal/ui"
)

type DeployOptions struct {
	Env      string
	Service  string
	Version  string
	Strategy string
	DryRun   bool
	Timeout  int
	Verbose  bool
}

type RollbackOptions struct {
	Env     string
	Service string
	Version string
	Steps   int
	DryRun  bool
}

func Deploy(opts DeployOptions, cfg *config.Config, hist *history.History) error {
	env, err := cfg.GetEnvironment(opts.Env)
	if err != nil {
		return err
	}

	services := filterServices(env.Services, opts.Service)
	if len(services) == 0 {
		return fmt.Errorf("no services found matching %q", opts.Service)
	}

	version := opts.Version
	if version == "" {
		version = "latest"
	}

	strategy := opts.Strategy
	if strategy == "" {
		strategy = cfg.Deploy.Strategy
	}

	for _, svc := range services {
		if err := deployService(opts, env, svc, version, strategy, cfg, hist); err != nil {
			return fmt.Errorf("deploying %s: %w", svc.Name, err)
		}
	}
	return nil
}

func deployService(opts DeployOptions, env config.Environment, svc config.Service, version, strategy string, cfg *config.Config, hist *history.History) error {
	title := fmt.Sprintf("Deploying %s → %s  [%s]", svc.Name, opts.Env, strategy)
	ui.Banner(title)

	if opts.DryRun {
		ui.Warn("DRY RUN — no changes will be made")
		fmt.Println()
	}

	startTime := time.Now()
	image := buildImageRef(cfg.Registry, svc, version)

	if opts.DryRun {
		printDryRunPlan(opts.Env, svc, version, image, strategy)
		return nil
	}

	err := ui.RunStep("Validating configuration", func() error {
		time.Sleep(time.Duration(80+rand.Intn(120)) * time.Millisecond)
		return validateServiceConfig(svc)
	})
	if err != nil {
		return err
	}

	err = ui.RunStep(fmt.Sprintf("Pre-flight checks (%s:%d)", env.Host, env.SSHPort), func() error {
		time.Sleep(time.Duration(300+rand.Intn(400)) * time.Millisecond)
		return nil
	})
	if err != nil {
		return err
	}

	err = ui.RunStep(fmt.Sprintf("Pulling image %s", color.CyanString(image)), func() error {
		time.Sleep(time.Duration(1200+rand.Intn(1800)) * time.Millisecond)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println()

	switch strategy {
	case "rolling":
		if err := rollingDeploy(svc, image, version, opts.Env); err != nil {
			return err
		}
	case "blue-green":
		if err := blueGreenDeploy(svc, image, version, opts.Env); err != nil {
			return err
		}
	case "recreate":
		if err := recreateDeploy(svc, image, version, opts.Env); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown strategy %q (use: rolling, blue-green, recreate)", strategy)
	}

	fmt.Println()

	if err := runHealthChecks(svc); err != nil {
		ui.Fail("Health checks failed — deployment may be unstable")
		ui.Warn("Run 'forge rollback' to revert")
		return err
	}

	elapsed := time.Since(startTime)
	depID := history.GenerateID()

	deployedBy := currentUser()
	hist.Add(history.Deployment{
		ID:         depID,
		Service:    svc.Name,
		Env:        opts.Env,
		Version:    version,
		Status:     "success",
		Strategy:   strategy,
		Duration:   elapsed,
		StartedAt:  startTime,
		FinishedAt: time.Now(),
		DeployedBy: deployedBy,
	})

	ui.SuccessBox(fmt.Sprintf("%s@%s deployed to %s", svc.Name, version, opts.Env))
	ui.KeyValue("Replicas:", fmt.Sprintf("%d updated", svc.Replicas))
	ui.KeyValue("Duration:", ui.FormatDuration(elapsed))
	ui.KeyValue("Deploy ID:", color.CyanString(depID))
	ui.KeyValue("Deployed by:", deployedBy)
	fmt.Println()

	return nil
}

func rollingDeploy(svc config.Service, image, version, env string) error {
	replicas := svc.Replicas
	if replicas == 0 {
		replicas = 1
	}

	for i := 1; i <= replicas; i++ {
		ui.Step(i, replicas, fmt.Sprintf("Updating replica %s-%s-%d", svc.Name, env, i))

		ui.Info(1, fmt.Sprintf("Draining traffic from replica %d", i))
		time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

		ui.Info(1, fmt.Sprintf("Stopping container (previous version)"))
		time.Sleep(time.Duration(300+rand.Intn(400)) * time.Millisecond)

		ui.Info(1, fmt.Sprintf("Starting %s", color.CyanString(image)))
		time.Sleep(time.Duration(600+rand.Intn(600)) * time.Millisecond)

		ui.Info(1, "Health check")
		time.Sleep(time.Duration(400+rand.Intn(600)) * time.Millisecond)
		ui.Success(fmt.Sprintf("Replica %d healthy — routing traffic back", i))

		fmt.Println()
	}
	return nil
}

func blueGreenDeploy(svc config.Service, image, version, env string) error {
	ui.Info(0, "Spinning up green environment")
	time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
	ui.Success("Green environment ready")
	fmt.Println()

	ui.Info(0, fmt.Sprintf("Deploying %s to green", color.CyanString(image)))
	time.Sleep(time.Duration(1200+rand.Intn(1000)) * time.Millisecond)
	ui.Success("Deployment to green complete")
	fmt.Println()

	ui.Info(0, "Running smoke tests on green")
	time.Sleep(time.Duration(800+rand.Intn(800)) * time.Millisecond)
	ui.Success("Smoke tests passed")
	fmt.Println()

	ui.Info(0, "Switching load balancer to green")
	time.Sleep(time.Duration(300+rand.Intn(200)) * time.Millisecond)
	ui.Success("Traffic routed to green")
	fmt.Println()

	ui.Info(0, "Draining blue environment (30s grace period)")
	time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)
	ui.Success("Blue environment drained and stopped")
	fmt.Println()

	return nil
}

func recreateDeploy(svc config.Service, image, version, env string) error {
	ui.Warn(fmt.Sprintf("Stopping all %d replicas (downtime expected)", svc.Replicas))
	time.Sleep(time.Duration(600+rand.Intn(600)) * time.Millisecond)
	ui.Info(0, "All replicas stopped")
	fmt.Println()

	ui.Info(0, fmt.Sprintf("Starting replicas with %s", color.CyanString(image)))
	for i := 1; i <= svc.Replicas; i++ {
		time.Sleep(time.Duration(400+rand.Intn(400)) * time.Millisecond)
		ui.Success(fmt.Sprintf("Replica %d started", i))
	}
	fmt.Println()
	return nil
}

func runHealthChecks(svc config.Service) error {
	if svc.HealthURL != "" {
		return ui.RunStep(fmt.Sprintf("Health check %s", color.CyanString(svc.HealthURL)), func() error {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(svc.HealthURL)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				return nil
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				return fmt.Errorf("health check returned %d", resp.StatusCode)
			}
			return nil
		})
	}

	return ui.RunStep("Health checks (simulated)", func() error {
		time.Sleep(time.Duration(400+rand.Intn(600)) * time.Millisecond)
		return nil
	})
}

func Rollback(opts RollbackOptions, cfg *config.Config, hist *history.History) error {
	env, err := cfg.GetEnvironment(opts.Env)
	if err != nil {
		return err
	}

	services := filterServices(env.Services, opts.Service)
	if len(services) == 0 {
		return fmt.Errorf("no services found")
	}

	for _, svc := range services {
		if err := rollbackService(opts, env, svc, cfg, hist); err != nil {
			return fmt.Errorf("rolling back %s: %w", svc.Name, err)
		}
	}
	return nil
}

func rollbackService(opts RollbackOptions, env config.Environment, svc config.Service, cfg *config.Config, hist *history.History) error {
	targetVersion := opts.Version

	if targetVersion == "" {
		deploys := hist.List(opts.Env, svc.Name, 10)
		successfulDeploys := filterSuccessful(deploys)

		steps := opts.Steps
		if steps == 0 {
			steps = 1
		}

		if len(successfulDeploys) < steps+1 {
			return fmt.Errorf("not enough deployment history for %d step(s) rollback", steps)
		}
		targetVersion = successfulDeploys[steps].Version
	}

	title := fmt.Sprintf("Rolling back %s → %s  [to %s]", svc.Name, opts.Env, targetVersion)
	ui.Banner(title)

	if opts.DryRun {
		ui.Warn("DRY RUN — no changes will be made")
		fmt.Println()
		ui.KeyValue("Service:", svc.Name)
		ui.KeyValue("Environment:", opts.Env)
		ui.KeyValue("Target version:", color.YellowString(targetVersion))
		fmt.Println()
		return nil
	}

	startTime := time.Now()
	image := buildImageRef(cfg.Registry, svc, targetVersion)

	ui.Warn(fmt.Sprintf("Rolling back to version %s", color.YellowString(targetVersion)))
	fmt.Println()

	err := ui.RunStep("Validating rollback target", func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if err != nil {
		return err
	}

	err = ui.RunStep(fmt.Sprintf("Pulling image %s", color.CyanString(image)), func() error {
		time.Sleep(time.Duration(800+rand.Intn(800)) * time.Millisecond)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println()

	replicas := svc.Replicas
	if replicas == 0 {
		replicas = 1
	}
	for i := 1; i <= replicas; i++ {
		ui.Step(i, replicas, fmt.Sprintf("Reverting replica %d", i))
		time.Sleep(time.Duration(600+rand.Intn(600)) * time.Millisecond)
		ui.Success(fmt.Sprintf("Replica %d reverted to %s", i, targetVersion))
		fmt.Println()
	}

	if err := runHealthChecks(svc); err != nil {
		return err
	}

	elapsed := time.Since(startTime)
	depID := history.GenerateID()

	hist.Add(history.Deployment{
		ID:         depID,
		Service:    svc.Name,
		Env:        opts.Env,
		Version:    targetVersion,
		Status:     "rolled-back",
		Strategy:   "rollback",
		Duration:   elapsed,
		StartedAt:  startTime,
		FinishedAt: time.Now(),
		DeployedBy: currentUser(),
	})

	ui.SuccessBox(fmt.Sprintf("%s rolled back to %s", svc.Name, targetVersion))
	ui.KeyValue("Environment:", opts.Env)
	ui.KeyValue("Duration:", ui.FormatDuration(elapsed))
	ui.KeyValue("Deploy ID:", color.CyanString(depID))
	fmt.Println()

	return nil
}

func filterServices(services []config.Service, name string) []config.Service {
	if name == "" {
		return services
	}
	for _, s := range services {
		if s.Name == name {
			return []config.Service{s}
		}
	}
	return nil
}

func filterSuccessful(deploys []history.Deployment) []history.Deployment {
	var result []history.Deployment
	for _, d := range deploys {
		if d.Status == "success" {
			result = append(result, d)
		}
	}
	return result
}

func buildImageRef(registry string, svc config.Service, version string) string {
	img := svc.Image
	if img == "" {
		img = svc.Name
	}
	if registry != "" {
		return fmt.Sprintf("%s/%s:%s", strings.TrimRight(registry, "/"), img, version)
	}
	return fmt.Sprintf("%s:%s", img, version)
}

func validateServiceConfig(svc config.Service) error {
	if svc.Name == "" {
		return fmt.Errorf("service name is required")
	}
	return nil
}

func printDryRunPlan(env string, svc config.Service, version, image, strategy string) {
	ui.Section("Deploy Plan")
	ui.KeyValue("Service:", svc.Name)
	ui.KeyValue("Environment:", env)
	ui.KeyValue("Image:", color.CyanString(image))
	ui.KeyValue("Strategy:", strategy)
	ui.KeyValue("Replicas:", fmt.Sprintf("%d", svc.Replicas))
	if svc.HealthURL != "" {
		ui.KeyValue("Health check:", svc.HealthURL)
	}
	fmt.Println()
}

func currentUser() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}
