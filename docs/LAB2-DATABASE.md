# Лабораторная 2: База данных и запись данных

## 1. Что сделано в проекте

- **PostgreSQL** подключается по DSN из переменных окружения (или из файла `.env`). Строка подключения собирается в `internal/app/dsn/dsn.go` (host, port, user, password, dbname).
- **GORM** используется как ORM: модели в `internal/app/repository/repository.go`, при старте приложения выполняется `AutoMigrate` — создаются/обновляются таблицы `users`, `services`, `requests`, `request_services`.
- **Пять HTTP-методов** (как в задании):
  - **3 GET** (через ORM): главная с поиском услуг, страница одной услуги, страница одной заявки.
  - **2 POST**: добавление услуги в заявку (ORM), логическое удаление заявки (raw SQL `UPDATE`), завершение заявки с расчётом по формуле (ORM).
- **Расчёт по формуле**: время = Σ (quantity × t_i × (1 + selectivity)); при отображении заявки значения считаются «на лету», при нажатии «Завершить заявку» — записываются в БД (в `requests` и в `request_services.calculated_time_ms`).

---

## 2. Структура БД: четыре таблицы

Имена таблиц в PostgreSQL (GORM даёт им имена в snake_case):

| Таблица            | Назначение |
|--------------------|------------|
| **users**          | Пользователи (создатель заявки, модератор). |
| **services**       | Услуги = типы индексов (B-tree, Hash, GIN и т.д.). |
| **requests**       | Заявки = симуляции запроса (корзина = черновик). |
| **request_services** | Связь «заявка – услуга» (м-м): какие индексы и в каком количестве входят в заявку. |

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
| id            | integer (PK)   | Идентификатор заявки. |
| status        | varchar(32)    | Статус: `draft`, `deleted`, `formed`, `completed`, `rejected`. |
| created_at    | timestamp      | Дата создания. |
| created_by_id | integer (FK→users) | Кто создал. |
| formed_at     | timestamp NULL | Дата формирования (по заданию — 2 действия создателя). |
| finished_at   | timestamp NULL | Дата завершения (2 действия модератора). |
| moderator_id  | integer NULL   | Модератор. |
| selectivity   | numeric        | Селективность (используется в формуле). |
| result_time   | varchar(64)    | Рассчитанное время (например `1.25ms`), заполняется при завершении. |
| result_memory | varchar(64)   | Рассчитанная память (например `96KB`). |

**Где пишется:**

- **INSERT:** при первом «Добавить в заявку», если у пользователя ещё нет черновика — создаётся новая строка в `requests` со статусом `draft`, заполняются `created_at`, `created_by_id`, `selectivity` (код: `AddServiceToDraft` в `repository.go`).
- **UPDATE (ORM):** при «Завершить заявку» — обновляются `status` → `completed`, `result_time`, `result_memory`, `finished_at` (`CompleteRequest`).
- **UPDATE (raw SQL):** при «Логически удалить заявку» — один запрос `UPDATE requests SET status = 'deleted' WHERE id = ? AND created_by_id = ?` (`DeleteRequestLogical`).

---

### 2.4. Таблица `request_services` (м-м)

| Столбец             | Тип            | Описание |
|---------------------|----------------|----------|
| request_id          | integer (PK, FK→requests) | Заявка. |
| service_id          | varchar(64) (PK, FK→services) | Услуга. |
| service_name        | varchar(255)   | Дублирование для отображения. |
| table_size          | varchar(64)    | То же. |
| selectivity         | numeric        | Селективность по позиции. |
| image_key           | varchar(255)   | Картинка. |
| quantity            | integer        | Количество (при повторном добавлении той же услуги увеличивается). |
| position            | integer        | Порядок в заявке. |
| is_main             | boolean        | Первая позиция = главная. |
| calculated_time_ms  | numeric(10,2)  | Рассчитанное время по формуле для этой позиции; заполняется при завершении заявки. |

Составной первичный ключ: `(request_id, service_id)` — одна и та же услуга в одной заявке не дублируется строкой, меняется только `quantity`.

**Где пишется:**

- **INSERT:** при нажатии «Добавить в заявку», если такой пары (request_id, service_id) ещё нет — создаётся новая строка в `request_services` с полями из заявки и услуги (`AddServiceToDraft`).
- **UPDATE:** если пара уже есть — увеличивается `quantity` в той же строке (`AddServiceToDraft`).
- **UPDATE:** при «Завершить заявку» для каждой строки этой заявки заполняется `calculated_time_ms` по формуле: quantity × speed_i × (1 + selectivity) (`CompleteRequest`).

Каскадное удаление отключено (по умолчанию внешние ключи без ON DELETE CASCADE).

---

## 3. Когда и куда что записывается (кратко)

| Действие пользователя      | Таблица(-ы)        | Операция |
|----------------------------|--------------------|----------|
| Запуск приложения          | все 4 таблицы      | AutoMigrate (создание/обновление схемы). |
| Наполнение через Adminer   | users, services    | INSERT из `scripts/seed.sql`. |
| «Добавить в заявку» (первый раз для пользователя) | requests, request_services | INSERT в обе. |
| «Добавить в заявку» (черновик уже есть, услуга новая) | request_services | INSERT. |
| «Добавить в заявку» (та же услуга уже в черновике) | request_services | UPDATE quantity. |
| «Завершить заявку»         | requests, request_services | UPDATE: status, result_time, result_memory, finished_at; в каждой строке request_services — calculated_time_ms. |
| «Логически удалить заявку» | requests           | UPDATE status = 'deleted' (один raw SQL). |

При открытии страницы заявки или главной с корзиной данные только **читаются** (SELECT через ORM); при пустом `result_time` время и память считаются в коде и показываются, но в БД не пишутся до нажатия «Завершить заявку».

---

## 4. Где смотреть записи в БД

### 4.1. Adminer

- URL: **http://localhost:8081** (если контейнер adminer из docker-compose запущен).
- Подключение: System **PostgreSQL**, Server **127.0.0.1** (или **host.docker.internal** из контейнера), User **postgres**, Password из `.env` (например `postgres`), Database **mydb**.

В левом меню выбираете базу **mydb**, затем таблицу: **users**, **services**, **requests**, **request_services**. Вкладка «Select» — просмотр данных, «SQL command» — произвольные запросы.

### 4.2. Полезные SQL-запросы (выполнять в Adminer → SQL command)

- Все заявки и их статусы:
```sql
SELECT id, status, created_at, created_by_id, selectivity, result_time, result_memory, finished_at
FROM requests
ORDER BY id;
```

- Логически удалённые заявки (для показа на защите):
```sql
SELECT id, status, created_at, created_by_id
FROM requests
WHERE status = 'deleted';
```

- Состав заявки и рассчитанное время по позициям (м-м):
```sql
SELECT rs.request_id, rs.service_id, rs.service_name, rs.quantity, rs.selectivity, rs.calculated_time_ms
FROM request_services rs
ORDER BY rs.request_id, rs.position;
```

- Одна заявка со всеми услугами:
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
| GET `/request/:id`        | SELECT заявки с Preload `Services.Service`; если удалённая — 404; при наличии услуг и пустом result_time — расчёт в памяти. | `GetRequest` → `GetRequest` |
| POST `/request/add`       | SELECT услуги и черновика; при отсутствии черновика — INSERT в `requests`, затем INSERT/UPDATE в `request_services`. | `AddToRequest` → `AddServiceToDraft` |
| POST `/request/:id/delete` | Один raw SQL: `UPDATE requests SET status = 'deleted' WHERE id = ? AND created_by_id = ?`. | `DeleteRequest` → `DeleteRequestLogical` |
| POST `/request/:id/complete` | SELECT заявки с услугами; расчёт времени/памяти; UPDATE `request_services` (calculated_time_ms); UPDATE `requests` (status, result_time, result_memory, finished_at). | `CompleteRequest` → `CompleteRequest` |

Роуты объявлены в `internal/api/server.go`. Обработчики — в `internal/app/handler/handler.go`, работа с БД — в `internal/app/repository/repository.go`.

---

## 6. Формула расчёта и где она используется

- **Время по заявке:**  
  `Время = Σ (quantity × t_i × (1 + selectivity))`  
  где `t_i` — число из поля `speed` услуги (например из `"0.3ms"` берётся 0.3). Сумма по всем строкам `request_services` этой заявки.

- **Память (упрощённо):**  
  `64 + 16 × количество_позиций` (KB).

В коде:

- `parseSpeedMs(s string)` в `repository.go` — парсит строку вида `"0.3ms"` в число.
- `calculateResultTimeAndMemory(req *Request)` — считает сумму по заявке и память.
- При отображении заявки (GET `/request/:id`) и при отображении корзины (GET `/`) для черновика с пустым `result_time` вызывается этот расчёт и результат показывается в интерфейсе (в БД не сохраняется).
- При нажатии «Завершить заявку» тот же расчёт вызывается в `CompleteRequest`; результат записывается в `requests.result_time` и `requests.result_memory`, а по каждой строке м-м — в `request_services.calculated_time_ms` (вклад позиции: quantity × t_i × (1 + selectivity)).

Итог: **куда записываются записи** — в таблицы `users` (вручную/seed), `services` (seed/Adminer), `requests` и `request_services` (приложение при добавлении в заявку, завершении и логическом удалении). **Где смотреть** — в Adminer в базе `mydb`, таблицы `users`, `services`, `requests`, `request_services`, плюс примеры SQL выше.
