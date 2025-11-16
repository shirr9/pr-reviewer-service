# PR Reviewer Service

Микросервис для автоматического назначения ревьюеров на Pull Request.

## Быстрый старт

```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`

## API

### Команды

**Создать команду**
```bash
POST /team/add
```

**Получить команду**
```bash
GET /team/get?team_name=backend
```

**Деактивировать команду**
```bash
POST /team/deactivate
```

### Пользователи

**Изменить статус**
```bash
POST /users/setIsActive
```

**Получить PR пользователя**
```bash
GET /users/getReview?user_id=u1
```

### Pull Requests

**Создать PR**
```bash
POST /pullRequest/create
```

**Merge PR**
```bash
POST /pullRequest/merge
```

**Переназначить ревьюера**
```bash
POST /pullRequest/reassign
```

### Статистика

**Получить статистику**
```bash
GET /statistics
```

## Тестирование

**Unit-тесты**
```bash
make test
```

**E2E тесты**
```bash
E2E_TEST=true make e2e-test
```

**Нагрузочное тестирование**
```bash
make load-test
```

**Линтер**
```bash
make lint
```

## Makefile

- `make build` - сборка
- `make test` - тесты
- `make lint` - линтер
- `make docker-up` - запуск
- `make docker-down` - остановка
- `make help` - справка
