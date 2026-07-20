# CI/CD

## Триггеры
- Push в `main` → CI Main
- Pull Request в `develop` → CI Pull Request
- Push в `release/**` → CI Release

## Этапы пайплайна (reusable.yaml)

### 1. Lint
- golangci-lint v2.6, только новые проблемы

### 2. Security
- govulncheck (проверка уязвимостей)

### 3. Build & Test
- go mod download
- gofmt проверка форматирования
- go test -v -race -coverprofile
- загрузка покрытия в Codecov
- go build

### 4. Check Conflicts
- проверка merge-конфликтов

### 5. Build Docker
- Docker Buildx
- пуш в ghcr.io (GitHub Container Registry)
- теги: SHA, image_tag, latest, dev

### 6. Comment Result
- автоматический комментарий к PR со статусом всех этапов

## PR Workflow
- При успешном CI — авто-approve + метка ready-to-merge
- Удаление метки WIP при успехе
