# Комп’ютерний практикум 4. Розроблення базової клієнтської частини вебзастосунку

REST API для керування притулками та котами (Go, Gin, GORM, PostgreSQL, Redis) і клієнтська частина на HTML, CSS та JavaScript.

## Клієнтська частина (КП4)

Сторінка відкривається за адресою http://localhost:8080/ і віддається тим самим сервером, що й API.

- `web/index.html` — розмітка: головний екран, коти, притулки, форми керування, шаблони `<template>` для карток.
- `web/assets/css/style.css` — стилі.
- `web/assets/js/api.js` — клієнт REST API: по одній функції на кожен ендпоінт, усі запити через Fetch API.
- `web/assets/js/app.js` — стан сторінки, відображення даних через DOM, обробники подій і форм.

Усі операції API доступні з інтерфейсу: перегляд котів і притулків, котів одного притулку, додавання, редагування, зміна статусу усиновлення, видалення одного запису й усіх записів. Після кожної зміни дані перезавантажуються без перезавантаження сторінки. Ключ API вводиться в розділі Manage і передається в заголовку `X-API-Key`. Форми перевіряються атрибутами HTML (`required`, `min`, `max`, `maxlength`), а помилки сервера показуються в рядку статусу.

## Сервер (КП1–КП3)

## Завдання

1. Замінити тимчасове зберігання даних із першої роботи на постійне зберігання в базі даних.
2. Додати валідацію вхідних даних.
3. Використовувати ORM і шаблон «Репозиторій» для роботи з базою даних.
4. Отримувати репозиторії через ін’єкцію залежностей.
5. Використовувати окремі DTO замість передачі сутностей бази даних через REST API.
6. Валідувати вхідні DTO за допомогою аналога FluentValidation для Go.

## Реалізація

- HTTP-фреймворк: Gin.
- ORM: GORM.
- База даних: PostgreSQL.
- `CatRepoPsql` і `ShelterRepoPsql` реалізують постійне зберігання.
- PostgreSQL foreign key зв’язує кота з притулком і забороняє видаляти притулок, у якому є коти (`ON DELETE RESTRICT`).
- Транзакції з `SELECT FOR UPDATE` не дозволяють одночасним запитам переповнити притулок.
- Репозиторії передаються в handlers через конструктори та інтерфейси.
- In-memory репозиторії збережені для ізольованих тестів.
- Вхідні та вихідні DTO знаходяться в `internal/dto`.
- Перетворення між DTO і моделями знаходяться в `internal/mapper`.
- Валідація виконується Gin через `go-playground/validator`.
- Невідомі JSON-поля відхиляються.
- Технічні помилки бази даних записуються в серверний лог і не повертаються клієнту.
- Відомі помилки мають однакові відповіді: `400` для неправильного зв’язку, `404` для відсутнього ресурсу та `409` для ресурсу із залежними записами або заповненого притулку.

## Правила валідації

### Cat

- `name` — обов’язкове, від 1 до 100 символів;
- `ageMonths` — обов’язкове, від 0 до 360 місяців;
- `gender` — обов’язкове, значення `male` або `female`;
- `shelterID` — обов’язкове, більше нуля;
- кота не можна додати або перемістити до притулку без вільних місць;
- `id` призначається базою даних і не приймається у вхідному DTO.

### Shelter

- `name` — обов’язкове, від 1 до 100 символів;
- `city` — обов’язкове, від 1 до 100 символів;
- `address` — обов’язкове, від 1 до 200 символів;
- `capacity` — обов’язкова, більше нуля та не менша за поточну кількість котів;
- `id` призначається базою даних і не приймається у вхідному DTO.

## Конфігурація, кешування та логування (КП3)

Конфігурація читається з `config.json` (шлях можна змінити змінною `CONFIG_PATH`). У Docker файл монтується в контейнер, тому після зміни достатньо `docker compose restart api`. Секрети підключення до PostgreSQL і Redis залишаються у `.env`.

| Параметр | Вплив на API |
|---|---|
| `apiKey` | ключ, який потрібно передавати в `X-API-Key` для `POST`, `PUT`, `DELETE` |
| `readOnly` | `true` вимикає змінюючі операції: вони повертають `403 API is in read-only mode` |
| `maxListSize` | максимальна кількість елементів у відповідях-списках; загальна кількість повертається в `X-Total-Count` |
| `cacheTTLSeconds` | час життя запису в кеші; `0` вимикає кешування |
| `logLevel` | рівень логування (`debug`, `info`, `warn`, `error`) |

Кешування: `GET /cats/{id}` і `GET /shelters/{id}` кешуються в Redis. Заголовок `X-Cache` показує `HIT`, `MISS` або `DISABLED`. Зміна чи видалення запису інвалідовує кеш.

Логування: декоратори `LoggedCatRepo` і `LoggedShelterRepo` записують кожен виклик репозиторію (метод, аргументи, тривалість, помилку) через `log/slog` у консоль. HTTP-запити логуються middleware Gin.

Усі залежності (конфігурація, логер, кеш, репозиторії) створюються в `internal/app` і передаються через конструктори.
## Запуск у Docker

Потрібні Docker і Docker Compose v2 або новіший.

```bash
cp .env.example .env
# Перед запуском задайте власні POSTGRES_PASSWORD і REDIS_USER_PASSWORD у .env.
docker compose up --build -d
```

У PowerShell:

```powershell
Copy-Item .env.example .env
docker compose up --build -d
```

Сервіси:

- API: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/swagger/index.html
- OpenAPI YAML: http://localhost:8080/openapi.yaml
- PostgreSQL з хоста: `localhost:5432`
- PostgreSQL у мережі Compose: `postgres:5432`
- Redis з хоста: `localhost:6380`
- Redis у мережі Compose: `redis:6379`

Для захищених операцій `POST`, `PUT` і `DELETE` потрібен заголовок:

```text
X-API-Key: demo
```

Перегляд логів:

```bash
docker compose logs -f
```

Зупинка:

```bash
docker compose down
```

Команда `docker compose down` зберігає дані в томі `postgres_data`. Не додавайте `-v`, якщо дані потрібно зберегти: цей параметр видаляє том разом із базою даних.

## Приклади запитів

Створення притулку:

```bash
curl -X POST http://localhost:8080/api/v1/shelters \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo" \
  -d '{"name":"Happy Paws","city":"Kyiv","address":"Main Street 1","capacity":50}'
```

Створення кота:

```bash
curl -X POST http://localhost:8080/api/v1/cats \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo" \
  -d '{"name":"Barsik","ageMonths":12,"gender":"male","adopted":false,"shelterID":1}'
```

Приклад помилки валідації:

```bash
curl -X POST http://localhost:8080/api/v1/cats \
  -H "Content-Type: application/json" \
  -H "X-API-Key: demo" \
  -d '{"id":10,"name":"Barsik","ageMonths":361,"gender":"unknown","shelterID":1}'
```

Відповідь:

```json
{"error":"invalid request body"}
```

## Тести

```bash
go test ./...
go vet ./...
```

## Генерація OpenAPI

Встановлення `swag`:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
```

Linux/macOS:

```bash
./scripts/generate-openapi.sh
```

PowerShell:

```powershell
./scripts/generate-openapi.ps1
```

Генеруються файли:

- `docs/docs.go`;
- `docs/swagger.json`;
- `docs/swagger.yaml`;
- `docs/openapi.yaml`.

## Структура

```text
cmd/api                  точка входу
internal/app             складання залежностей
internal/config          конфігурація
internal/dto             DTO REST API
internal/handlers        HTTP handlers
internal/mapper          перетворення DTO і моделей
internal/models          моделі бази даних
internal/repositories    PostgreSQL та in-memory репозиторії
internal/router          маршрути й middleware
web                      клієнтська частина (HTML, CSS, JS)
docs                     Swagger та OpenAPI
```
