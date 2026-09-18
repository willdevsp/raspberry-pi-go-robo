# Diretrizes de Git Flow e Automação CI/CD com GitHub Actions

Este documento define o padrão de branches e a automação de integração e entrega contínua executada via **GitHub Actions**.

---

## 1. Estrutura de Branches

* **`main`**: Código em produção. Protegida contra commits diretos e merges manuais não auditados.
* **`develop`**: Branch de integração contínua e homologação. Recebe as features finalizadas.
* **`feature/<nome-da-feature>`**: Branches temporárias criadas exclusivamente a partir da `develop` para desenvolvimento de novas funcionalidades.

---

## 2. Ciclo de Desenvolvimento e Workflows (GitHub Actions)

```text
feature/*  ──(PR)──> [CI: Lint & Testes] ──(Auto-merge)──> develop
                                                              │
develop    <──────────────────────────────────────────────────┘
   │
   └── [CI: Reteste em develop] ──(Auto-create PR)──> PR: develop -> main
                                                           │
main       <──(Merge)── [Gera Tag de Release] ◄── [CI: Testes Finais]

```

---

### Fase 1: Feature e Merge Automático em `develop`

1. **Criação da Branch:** O desenvolvedor parte da `develop` atualizada:
```bash
git checkout develop
git pull origin develop
git checkout -b feature/nome-da-feature

```


2. **Push e Pipeline de Validação (`pr-feature-validation.yml`):**
* Ao concluir as edições, o desenvolvedor faz um `push` direto para sua branch `feature/*`.
* **Disparo:** Evento `push` contra branches `feature/*`.
* **Etapas executadas pelo runner:**
* Validação de padronização de código (**Linter**).
* Execução da suíte de **testes unitários** e cobertura de código.
3. **Criação de PR Automática:**
* Se a esteira passar com sucesso, uma action (`peter-evans/create-pull-request`) abre um PR apontando automaticamente para a branch `develop`.
* **Aprovação Manual (Code Review):** O PR permanece aberto aguardando a avaliação e aprovação manual do desenvolvedor responsável antes de ser mesclado na `develop`.



---

### Fase 2: Reteste, Geração da Tag e Abertura Automática de PR para `main`

1. **Pipeline de Integração (`develop-integration.yml`):**
* **Disparo:** Evento `push` na branch `develop` (acionado logo após o merge da feature).
* **Etapas executadas pelo runner:**
* Execução completa dos testes no contexto integrado da `develop` (testes unitários, de integração e validação de build).
2. **Geração da Release Tag:**
* Com os testes aprovados na `develop`, o workflow calcula a próxima versão semântica (SemVer) e cria a tag (ex: `v1.3.0`) ainda na branch `develop`.
3. **Criação Automática do PR:**
* A Action utiliza a action `peter-evans/create-pull-request` para criar um Pull Request de `develop` com destino à `main`, incluindo a nova tag gerada no título, aguardando aprovação manual.

---

### Fase 3: Validação Final e Merge em `main`

1. **Pipeline de Release (`release-pipeline.yml`):**
* **Disparo:** Eventos `pull_request` contra a branch `main`.
* **Etapas executadas pelo runner:**
* Execução da suíte de testes finais (regressão/smoke tests).
2. **Merge para Produção:**
* O PR é avaliado manualmente e mesclado na `main`.
* O deploy de produção é disparado a partir do merge na `main`.



---

## 3. Requisitos e Configurações no GitHub

Para que as automações funcionem corretamente, o repositório deve conter as seguintes configurações:

1. **Permissões do `GITHUB_TOKEN`:**
* No workflow, garantir permissões de escrita:
```yaml
permissions:
  contents: write
  pull-requests: write

```




2. **Configurações de Repositório (`Settings` > `General`):**
* Habilitar **Automatically delete head branches** (para limpar as branches de feature após o merge).


3. **Regras de Proteção de Branch (`Settings` > `Branches`):**
* **`develop`**: Exigir aprovação de status checks (Linter e Testes Unitários) antes de permitir merges.
* **`main`**: Exigir status checks da pipeline de validação final e bloquear commits diretos.