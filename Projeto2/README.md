<div align="center">
  <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original-wordmark.svg" width="80" />
  <h1>forge</h1>
  <p>CLI de deploy em Go para gerenciar serviços em múltiplos ambientes.</p>

  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/Cobra-1.10-7C3AED?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="https://github.com/spf13/viper"><img src="https://img.shields.io/badge/Viper-1.21-0EA5E9?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-22c55e?style=flat-square" /></a>
</div>

<br/>

Declare um `.forge.yaml`, o resto é com o forge — rolling, blue-green ou recreate com réplicas, health checks, rollback automático via histórico persistente e variáveis por ambiente.

```yaml
# .forge.yaml
project: myapp
registry: registry.example.com/myapp

deploy:
  strategy: rolling   # rolling | blue-green | recreate
  timeout: 300

environments:
  production:
    host: prod.example.com
    services:
      - name: api
        port: 8080
        health_url: http://prod.example.com:8080/health
        replicas: 3
        image: api
```

<br/>

---

## Instalação

> **Pré-requisitos** — `go 1.22+`

```bash
git clone https://github.com/user/forge
cd forge
go build -o forge .

# ou instalar direto no PATH
go install github.com/user/forge@latest
```

Para inicializar um projeto existente:

```bash
forge init                         # gera .forge.yaml com dev/staging/production
forge init --project myapp         # define o nome do projeto
forge init --registry ghcr.io/org  # define o registry
```

---

## Como fica na prática

Após configurar o `.forge.yaml`, o forge cuida do ciclo completo de deploy.

**Rolling deploy com 3 réplicas**

```
$ forge deploy api --env production --version v2.1.0

  ┌──────────────────────────────────────────────┐
  │  Deploying api → production  [rolling]       │
  └──────────────────────────────────────────────┘

  → Validating configuration ✓ (116ms)
  → Pre-flight checks (prod.example.com:22) ✓ (459ms)
  → Pulling image registry.example.com/myapp/api:v2.1.0 ✓ (1.4s)

  [1/3] Updating replica api-production-1
    → Draining traffic from replica 1
    → Stopping container (previous version)
    → Starting registry.example.com/myapp/api:v2.1.0
    → Health check
  ✓ Replica 1 healthy — routing traffic back

  [2/3] Updating replica api-production-2
    ...
  ✓ Replica 2 healthy — routing traffic back

  [3/3] Updating replica api-production-3
    ...
  ✓ Replica 3 healthy — routing traffic back

  → Health check http://prod.example.com:8080/health ✓ (515ms)

  ╔══════════════════════════════════════╗
  ║  ✓ api@v2.1.0 deployed to production ║
  ╚══════════════════════════════════════╝

  Replicas:           3 updated
  Duration:           14.7s
  Deploy ID:          dep-20260520-204419
  Deployed by:        alice
```

**Status de todos os ambientes**

```
$ forge status

  ── Environment: production ──

  Host:               prod.example.com

  SERVICE  REPLICAS  PORT  STATUS      HEALTH URL
  api      3         8080  ● healthy   http://prod.example.com:8080/health
  worker   2         —     ○ unknown   —


  ── Environment: staging ──

  Host:               staging.example.com

  SERVICE  REPLICAS  PORT  STATUS       HEALTH URL
  api      2         8080  ● healthy    http://staging.example.com:8080/health
  worker   1         —     ○ unknown    —
```

**Histórico de deploys**

```
$ forge history --env production

  ── Deployment History — production ──

  ID                SERVICE  ENV         VERSION  STRATEGY    STATUS      DURATION
  dep-20260520-210  api      production  v2.1.0   rolling     ● success   14.7s
  dep-20260519-183  api      production  v2.0.0   blue-green  ● success   9.2s
  dep-20260518-094  api      production  v1.9.1   rolling     ● success   11.3s
  dep-20260517-162  api      production  v1.9.0   rolling     ● rolled-back 8.1s

  Showing 4 of 4 total deployments
```

---

## Estratégias de deploy

### Rolling (padrão)

Substitui réplicas uma a uma, sem downtime. Cada réplica passa por health check antes da próxima ser atualizada.

```bash
forge deploy api --env production --version v2.1.0
forge deploy api --env production --version v2.1.0 --strategy rolling
```

### Blue-Green

Sobe um ambiente paralelo ("green"), valida com smoke tests e então migra o tráfego. O ambiente antigo ("blue") só é destruído após a confirmação.

```bash
forge deploy api --env production --version v2.1.0 --strategy blue-green
```

```
  → Spinning up green environment          ✓ (1.2s)
  → Deploying api:v2.1.0 to green          ✓ (2.1s)
  → Running smoke tests on green           ✓ (0.9s)
  → Switching load balancer to green       ✓ (0.3s)
  → Draining blue environment (30s grace)  ✓ (0.5s)
```

### Recreate

Para todas as réplicas e sobe a nova versão. Causa downtime breve, indicado para ambientes de desenvolvimento ou migrações de banco que exigem exclusividade.

```bash
forge deploy api --env staging --version v2.1.0 --strategy recreate
```

### Dry-run

Mostra exatamente o que seria feito, sem executar nada.

```bash
forge deploy --env production --dry-run

  ⚠ DRY RUN — no changes will be made

  Service:            api
  Environment:        production
  Image:              registry.example.com/myapp/api:latest
  Strategy:           rolling
  Replicas:           3
  Health check:       http://prod.example.com:8080/health
```

---

## Rollback

O forge mantém histórico persistente em `~/.forge/history.json`. O rollback consulta esse histórico automaticamente.

```bash
# reverte para o deploy de sucesso anterior
forge rollback api --env production

# reverte para dois deploys atrás
forge rollback api --env production --steps 2

# reverte para uma versão específica
forge rollback api --env production --version v1.9.1

# preview sem executar
forge rollback api --env production --dry-run
```

```
  ┌──────────────────────────────────────────────┐
  │  Rolling back api → production  [to v1.9.1]  │
  └──────────────────────────────────────────────┘

  ⚠ Rolling back to version v1.9.1

  → Validating rollback target ✓ (200ms)
  → Pulling image registry.example.com/myapp/api:v1.9.1 ✓ (1.5s)

  [1/3] Reverting replica 1  ✓ Replica 1 reverted to v1.9.1
  [2/3] Reverting replica 2  ✓ Replica 2 reverted to v1.9.1
  [3/3] Reverting replica 3  ✓ Replica 3 reverted to v1.9.1

  ╔══════════════════════════════════════╗
  ║  ✓ api rolled back to v1.9.1         ║
  ╚══════════════════════════════════════╝
```

---

## Variáveis de ambiente

Variáveis ficam declaradas por ambiente no `.forge.yaml` e são gerenciadas via `forge env`.

```bash
forge env list --env staging

  KEY           VALUE
  database_url  postgres://staging-db/myapp_staging
  redis_url     redis://staging-redis:6379
  log_level     info

forge env get DATABASE_URL --env staging
# → postgres://staging-db/myapp_staging

forge env set FEATURE_FLAG true --env staging
forge env unset OLD_KEY --env production

# exportar como variáveis de shell
eval $(forge env export --env dev)
```

> Chaves que contêm `PASSWORD`, `SECRET`, `TOKEN` ou `KEY` têm o valor mascarado no `list`.

---

## Logs

```bash
# últimas 50 linhas (padrão)
forge logs api --env dev

# últimas 100 linhas filtrando por padrão
forge logs api --env staging --tail 100 --filter "ERROR"

# seguir em tempo real
forge logs api --env dev --follow

# a partir de uma janela de tempo
forge logs api --env production --since 1h
```

```
  2026-05-20T20:44:10.000-03:00  INFO   GET /api/users status=200 latency=12ms
  2026-05-20T20:44:13.000-03:00  WARN   Slow query detected query="SELECT *" duration=234ms
  2026-05-20T20:44:16.000-03:00  ERROR  Failed to send email to=user@example.com error=timeout
  2026-05-20T20:44:19.000-03:00  INFO   Background job completed job=email-digest duration=1.2s
```

---

## Referência da spec

```yaml
project: myapp
version: "1.0"
registry: registry.example.com/myapp   # prefixo do container registry

deploy:
  strategy: rolling    # rolling | blue-green | recreate
  timeout: 300         # segundos antes de abort
  retries: 3

environments:
  production:
    host: prod.example.com
    ssh_user: deploy             # usuário SSH
    ssh_port: 22
    deploy_path: /opt/apps/myapp

    services:
      - name: api
        port: 8080
        health_url: http://prod.example.com:8080/health   # opcional
        container: myapp-api
        replicas: 3              # 1 = single instance · 2+ = múltiplas réplicas
        image: api               # concatenado com registry → registry/image:version

    env_vars:
      DATABASE_URL: postgres://prod-db/myapp_production
      LOG_LEVEL: warn
```

---

## Comandos úteis

```bash
# deploy de todos os serviços de um ambiente
forge deploy --env staging --version v2.1.0

# watch de status com refresh automático a cada 5s
forge status --watch

# filtrar status por ambiente e serviço
forge status --env production --service api

# histórico em JSON para integrar com outras ferramentas
forge history --env production --format json

# limpar histórico de um ambiente
forge history --env staging --clear

# versão instalada
forge version
```

---

## Estrutura

```
main.go

cmd/
  root.go          flags globais (--config, --verbose, --no-color) + Viper
  init.go          forge init — gera .forge.yaml
  deploy.go        forge deploy — rolling, blue-green, recreate
  rollback.go      forge rollback — histórico automático ou versão explícita
  status.go        forge status — health checks por ambiente
  logs.go          forge logs — tail, follow e filter
  env.go           forge env list | get | set | unset | export
  history.go       forge history — tabela ou JSON, com --clear
  version.go       forge version

internal/
  config/
    config.go      struct Config + Load() com Viper unmarshal
  deployer/
    deployer.go    Deploy() e Rollback() com as três estratégias
  history/
    history.go     persistência em ~/.forge/history.json
  ui/
    ui.go          Banner, RunStep, StatusBadge, tabelas coloridas
```

---

<div align="center">
  <sub>Built with <a href="https://github.com/spf13/cobra">Cobra</a> · <a href="https://github.com/spf13/viper">Viper</a> · <a href="https://github.com/fatih/color">color</a> · MIT</sub>
</div>
