# pr-reviewer-service
Сервис назначения ревьюеров для Pull Request’ов

## Запуск
```bash
docker-compose up --build
```

## Пояснения
- Добавлено ограничение: только активный пользователь может создавать Pull Request. 
Я добавила его для корректной работы системы и предотвращения действий со стороны заблокированных,
удалённых или неактуальных пользователей.
- В эндпоинте `/pullRequest/reassign` поле для старого ревьюера называется `old_reviewer_id` (а не `old_user_id`),
так как это более логичное название. В required указано `[pull_request_id, old_user_id]`, 
но в примере используется `old_reviewer_id`.

## Покрытие бизнес логики unit тестами
```
   ~/GolandProjects/pr-reviewer-service  go test ./internal/app/service/... -cover  
ok      github.com/shirr9/pr-reviewer-service/internal/app/service      0.010s  coverage: 75.0% of statements
```