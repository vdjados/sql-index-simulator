# Лабораторная 2: База данных и запись данных

## 1. Что сделано в проекте

- **PostgreSQL** подключается по DSN из переменных окружения (или из файла `.env`). Строка подключения собирается в `internal/app/dsn/dsn.go` (host, port, user, password, dbname).
- **GORM** используется как ORM: модели в `internal/app/repository/repository.go`, при старте приложения выполняется `AutoMigrate` — создаются/обновляются таблицы `users`, `services`, `requests`, `request_services`.
- **Пять HTTP-методов** (как в задании):
  - **3 GET** (через ORM): главная с поиском услуг, страница одной услуги, страница одного sql_query.
  - **2 POST**: добавление услуги в sql_query (ORM), удаление sql_query (raw SQL `UPDATE`).
- **Результат в м-м**: время и память запроса считаются по формуле при отображении и сохраняются в БД при завершении (в `requests` и в `request_services.calculated_time_ms`).

---

## 2. Структура БД: четыре таблицы

Имена таблиц в PostgreSQL (GORM даёт им имена в snake_case):

| Таблица            | Назначение |
|--------------------|------------|
| **users**          | Пользователи (создатель sql_query, модератор). |
| **services**       | Услуги = типы индексов (B-tree, Hash, GIN и т.д.). |
| **requests**       | sql_query = симуляции запроса (корзина = черновик). |
| **request_services** | Связь «sql_query – услуга» (м-м): какие индексы и в каком количестве входят в sql_query. |

### 2.1. Таблица `users`

| Столбец   | Тип           | Описание |
|-----------|---------------|----------|
| id        | integer (PK)  | Идентификатор. |
| name      | varchar(255)  | Имя. |
| email     | varchar(255)  | Уникальный индекс. |
| role      | varchar(32)   | По умолчанию `'user'`. |

**Где пишется:** записи создаются вручную или скриптом (например, `scripts/seed.sql`). Приложение всегда работает с пользователем с `id = 1` (зашито в хендлерах).

---

### 2.2. Таблица `services`

| Столбец     | Тип           | Описание |
|-------------|---------------|----------|
| id          | varchar(64) PK | Идентификатор услуги (например `btree-10k`, `gin-500k`). |
| name        | varchar(255)  | Название индекса. |
| table_size  | varchar(64)   | Размер таблицы (например `10k rows`). |
| speed       | varchar(64)   | Время в виде строки (например `0.3ms`). |
| description | text          | Описание. |
| image_key   | varchar(255)  | Имя файла картинки. |
| gif_key     | varchar(255)  | Имя GIF. |
| status      | varchar(32)   | `active` или `deleted` (логическое удаление услуги). |

**Где пишется:** начальное наполнение — скрипт `scripts/seed.sql` (выполняется в Adminer). Новые услуги можно добавлять вручную в Adminer в таблицу `services`.

---

### 2.3. Таблица `requests`

| Столбец       | Тип            | Описание |
|---------------|----------------|----------|
| id                 | integer (PK)   | Идентификатор sql_query. |
| status             | varchar(32)    | Статус: `draft`, `deleted`, `formed`, `completed`, `rejected`. |
| created_at         | timestamp      | Дата создания. |
| created_by_id      | integer (FK→users) | Кто создал. |
| formed_at          | timestamp NULL | Дата формирования. |
| finished_at        | timestamp NULL | Дата завершения. |
| moderator_id       | integer NULL   | Модератор. |
| query_description  | text           | Описание запроса текстом (поле Заявка). |
| selectivity        | numeric        | Селективность в запросе (м-м). |
| result_time        | varchar(64)    | Время запроса. |
| result_memory      | varchar(64)    | Память запроса. |

**Где пишется:**

- **INSERT:** при первом «Добавить в запрос», если у пользователя ещё нет черновика — создаётся новая строка в `requests` со статусом `draft`, заполняются `created_at`, `created_by_id`, `selectivity` (код: `AddServiceToDraft` в `repository.go`).
- **UPDATE (ORM):** при завершении расчёта — обновляются `status` → `completed`, `result_time`, `result_memory`, `finished_at` (`CompleteRequest`).
- **UPDATE (raw SQL):** при «Удалить запрос» — один запрос `UPDATE requests SET status = 'deleted' WHERE id = ? AND created_by_id = ?` (`DeleteRequestLogical`).

---

### 2.4. Таблица `request_services` (м-м)

| Столбец             | Тип            | Описание |
|---------------------|----------------|----------|
| request_id          | integer (PK, FK→requests) | sql_query. |
| service_id          | varchar(64) (PK, FK→services) | Услуга (таблица + индекс). |
| service_name        | varchar(255)   | Дублирование для отображения. |
| table_size          | varchar(64)    | Размер таблицы. |
| selectivity         | numeric        | Селективность в запросе (м-м). |
| image_key           | varchar(255)   | Картинка. |
| quantity            | integer        | Количество. |
| position            | integer        | Порядок в sql_query. |
| is_main             | boolean        | Первая позиция = главная. |
| calculated_time_ms  | numeric(10,2)  | Время по позиции (результат в м-м). |

Составной первичный ключ: `(request_id, service_id)` — одна и та же услуга в одном sql_query не дублируется строкой, меняется только `quantity`.

**Где пишется:**

- **INSERT:** при нажатии «Добавить в запрос», если такой пары (request_id, service_id) ещё нет — создаётся новая строка в `request_services` с полями из sql_query и услуги (`AddServiceToDraft`).
- **UPDATE:** если пара уже есть — увеличивается `quantity` в той же строке (`AddServiceToDraft`).
- **UPDATE:** при завершении расчёта для каждой строки заполняется `calculated_time_ms` (`CompleteRequest`).

Каскадное удаление отключено (по умолчанию внешние ключи без ON DELETE CASCADE).

---

## 3. Когда и куда что записывается (кратко)

| Действие пользователя      | Таблица(-ы)        | Операция |
|----------------------------|--------------------|----------|
| Запуск приложения          | все 4 таблицы      | AutoMigrate (создание/обновление схемы). |
| Наполнение через Adminer   | users, services    | INSERT из `scripts/seed.sql`. |
| «Добавить в запрос» (первый раз для пользователя) | requests, request_services | INSERT в обе. |
| «Добавить в запрос» (черновик уже есть, услуга новая) | request_services | INSERT. |
| «Добавить в запрос» (та же услуга уже в черновике) | request_services | UPDATE quantity. |
| «Удалить запрос» | requests | UPDATE status = 'deleted' (один raw SQL). |

При открытии страницы sql_query или главной с корзиной данные только **читаются** (SELECT через ORM); при пустом `result_time` время и память считаются в коде и показываются.

---

## 4. Где смотреть записи в БД

### 4.1. Adminer

- URL: **http://localhost:8081** (если контейнер adminer из docker-compose запущен).
- Подключение: System **PostgreSQL**, Server **127.0.0.1** (или **host.docker.internal** из контейнера), User **postgres**, Password из `.env` (например `postgres`), Database **mydb**.

В левом меню выбираете базу **mydb**, затем таблицу: **users**, **services**, **requests**, **request_services**. Вкладка «Select» — просмотр данных, «SQL command» — произвольные запросы.

### 4.2. Полезные SQL-запросы (выполнять в Adminer → SQL command)

- Все sql_query и их статусы:
```sql
SELECT id, status, created_at, created_by_id, selectivity, result_time, result_memory, finished_at
FROM requests
ORDER BY id;
```

- Удалённые sql_query (для показа на защите):
```sql
SELECT id, status, created_at, created_by_id
FROM requests
WHERE status = 'deleted';
```

- Состав sql_query и время по позициям (м-м):
```sql
SELECT rs.request_id, rs.service_id, rs.service_name, rs.quantity, rs.selectivity, rs.calculated_time_ms
FROM request_services rs
ORDER BY rs.request_id, rs.position;
```

- Один sql_query со всеми услугами:
```sql
SELECT r.id, r.status, r.result_time, r.result_memory,
       rs.service_id, rs.service_name, rs.quantity, rs.calculated_time_ms
FROM requests r
LEFT JOIN request_services rs ON rs.request_id = r.id
WHERE r.id = 1
ORDER BY rs.position;
```
(подставьте нужный `r.id`).

---

## 5. Маршруты приложения и их работа с БД

| Метод и путь              | Действие с БД | Где в коде |
|---------------------------|----------------|------------|
| GET `/`                   | SELECT из `services` (с фильтром), поиск черновика в `requests` + `request_services` (Preload). Для черновика при пустом result_time — расчёт в памяти (в БД не пишется). | `GetServices` → `GetServices`, `GetCurrentRequest` |
| GET `/service/:id`        | SELECT из `services` по id. | `GetService` → `GetService` |
| GET `/sql_query/:id`     | SELECT sql_query с Preload `Services.Service`; если не найден или удалён — редирект на `/`. При пустом result_time — расчёт в памяти. | `GetSqlQuery` → `GetRequest` |
| POST `/sql_query/add`    | SELECT услуги и черновика; при отсутствии черновика — INSERT в `requests`, затем INSERT/UPDATE в `request_services`. | `AddToSqlQuery` → `AddServiceToDraft` |
| POST `/sql_query/:id/delete` | Один raw SQL: `UPDATE requests SET status = 'deleted' WHERE id = ? AND created_by_id = ?`. | `DeleteSqlQuery` → `DeleteRequestLogical` |

Роуты объявлены в `internal/api/server.go`. Обработчики — в `internal/app/handler/handler.go`, работа с БД — в `internal/app/repository/repository.go`.

---

## 6. Результат в м-м: время и память запроса

Время и память считаются в коде (`calculateResultTimeAndMemory` в `repository.go`) и отображаются на странице sql_query и в корзине. При завершении расчёта результат записывается в `requests.result_time`, `requests.result_memory` и в `request_services.calculated_time_ms`.

**Где смотреть** — в Adminer в базе `mydb`, таблицы `users`, `services`, `requests`, `request_services`, плюс примеры SQL выше.
