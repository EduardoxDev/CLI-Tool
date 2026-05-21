package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/user/forge/internal/ui"
)

var (
	initProject  string
	initRegistry string
	initForce    bool
)

var initCmd = &cobra.Command{
	Use:         "init",
	Short:       "Initialize a .forge.yaml configuration file",
	Annotations: map[string]string{"skip-config": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		const target = ".forge.yaml"

		if _, err := os.Stat(target); err == nil && !initForce {
			return fmt.Errorf("%s already exists (use --force to overwrite)", target)
		}

		project := initProject
		if project == "" {
			cwd, _ := os.Getwd()
			parts := strings.Split(strings.ReplaceAll(cwd, "\\", "/"), "/")
			project = parts[len(parts)-1]
		}

		content := generateConfig(project, initRegistry)

		if err := os.WriteFile(target, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}

		ui.SuccessBox(fmt.Sprintf("%s created", target))
		ui.KeyValue("Project:", project)
		ui.KeyValue("Environments:", "dev, staging, production")
		ui.KeyValue("Default strategy:", "rolling")
		fmt.Println()
		fmt.Printf("  Edit %s to configure your environments and services.\n", color.CyanString(target))
		fmt.Println()

		return nil
	},
}

func generateConfig(project, registry string) string {
	if registry == "" {
		registry = "registry.example.com/" + project
	}

	tmpl := `project: PROJECT
version: "1.0"
registry: REGISTRY

deploy:
  strategy: rolling      # rolling | blue-green | recreate
  timeout: 300
  retries: 3

environments:
  dev:
    host: localhost
    ssh_user: deploy
    ssh_port: 22
    deploy_path: /opt/apps/PROJECT
    services:
      - name: api
        port: 8080
        health_url: http://localhost:8080/health
        container: PROJECT-api
        replicas: 1
        image: api
      - name: worker
        port: 0
        container: PROJECT-worker
        replicas: 1
        image: worker
    env_vars:
      DATABASE_URL: postgres://localhost/PROJECT_dev
      REDIS_URL: redis://localhost:6379
      LOG_LEVEL: debug

  staging:
    host: staging.example.com
    ssh_user: deploy
    ssh_port: 22
    deploy_path: /opt/apps/PROJECT
    services:
      - name: api
        port: 8080
        health_url: http://staging.example.com:8080/health
        container: PROJECT-api
        replicas: 2
        image: api
      - name: worker
        container: PROJECT-worker
        replicas: 1
        image: worker
    env_vars:
      DATABASE_URL: postgres://staging-db/PROJECT_staging
      REDIS_URL: redis://staging-redis:6379
      LOG_LEVEL: info

  production:
    host: prod.example.com
    ssh_user: deploy
    ssh_port: 22
    deploy_path: /opt/apps/PROJECT
    services:
      - name: api
        port: 8080
        health_url: http://prod.example.com:8080/health
        container: PROJECT-api
        replicas: 3
        image: api
      - name: worker
        container: PROJECT-worker
        replicas: 2
        image: worker
    env_vars:
      DATABASE_URL: postgres://prod-db/PROJECT_production
      REDIS_URL: redis://prod-redis:6379
      LOG_LEVEL: warn
`
	tmpl = strings.ReplaceAll(tmpl, "PROJECT", project)
	tmpl = strings.ReplaceAll(tmpl, "REGISTRY", registry)
	return tmpl
}

func init() {
	initCmd.Flags().StringVar(&initProject, "project", "", "project name (default: current directory name)")
	initCmd.Flags().StringVar(&initRegistry, "registry", "", "container registry URL")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config")
	rootCmd.AddCommand(initCmd)
}
