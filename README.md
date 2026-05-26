# applycation

Self-host платформа для поиска вакансий и автооткликов на hh.ru.

## Quick start

1. Создайте локальный `.env` с переменными окружения.

2. Установите toolchain через mise:

```bash
mise install
```

3. Поставьте зависимости:

```bash
mise run setup
```

4. Запуск локально:

```bash
mise run backend:dev
mise run frontend:dev
```

5. Запуск в Docker:

```bash
mise run docker:up
```

## Основные команды

- `mise run lint`
- `mise run test`
- `mise run codegen`
- `mise run docker:build`
- `mise run docker:up`
- `mise run docker:down`
- `mise run db:migrate-up`
- `mise run db:migrate-down`
