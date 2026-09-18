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


2. **Abertura de PR:** O desenvolvedor abre um Pull Request apontando para a branch `develop`.
3. **Pipeline de Validação (`pr-feature-validation.yml`):**
* **Disparo:** Eventos `pull_request` (abertura e sincronização) contra a branch `develop`.
* **Etapas executadas pelo runner:**
* Validação de padronização de código (**Linter**).
* Execução da suíte de **testes unitários** e cobertura de código.


* **Auto-Merge:** Configurado via `gh pr merge --auto --squash` (ou action equivalente). Assim que os status checks forem aprovados e as regras de proteção forem atendidas, o GitHub Actions realiza o merge automático na `develop`.



---

### Fase 2: Reteste em `develop` e Abertura Automática de PR para `main`

1. **Pipeline de Integração (`develop-integration.yml`):**
* **Disparo:** Evento `push` na branch `develop` (acionado logo após o merge da feature).
* **Etapas executadas pelo runner:**
* Execução completa dos testes no contexto integrado da `develop` (testes unitários, de integração e validação de build).




2. **Criação Automática do PR:**
* Caso todos os testes passem com sucesso, a Action utiliza a GitHub CLI (`gh pr create`) ou a action `peter-evans/create-pull-request` para verificar se já existe um PR aberto de `develop` para `main`.
* Se não existir, a esteira abre automaticamente o Pull Request de `develop` com destino à `main`.



---

### Fase 3: Validação Final, Tag de Release e Merge em `main`

1. **Pipeline de Release (`release-pipeline.yml`):**
* **Disparo:** Eventos `pull_request` contra a branch `main`.
* **Etapas executadas pelo runner:**
* Execução da suíte de testes finais (regressão/smoke tests).




2. **Geração da Release Tag:**
* Com os testes 100% aprovados e antes da conclusão do merge, o workflow calcula a próxima versão semântica (SemVer) com base no histórico de commits.
* A Action gera e publica a tag no repositório (ex: `v1.3.0`) apontando para o commit validado.


3. **Merge para Produção:**
* O PR é mesclado na `main` (usando estratégia de *Merge Commit* para preservar histórico de releases).
* O deploy de produção é disparado a partir da criação da tag ou do merge na `main`.



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
* Habilitar a opção **Allow auto-merge**.
* Habilitar **Automatically delete head branches** (para limpar as branches de feature após o merge).


3. **Regras de Proteção de Branch (`Settings` > `Branches`):**
* **`develop`**: Exigir aprovação de status checks (Linter e Testes Unitários) antes de permitir merges.
* **`main`**: Exigir status checks da pipeline de validação final e bloquear commits diretos.