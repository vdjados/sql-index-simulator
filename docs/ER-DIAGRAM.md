# ER-диаграмма (структура для StarUML)

Предметная область: **SQL Index Simulator** — sql_query на симуляцию запросов с выбором индексов (услуг: таблица + индекс).

---

## Зачем нужна каждая таблица и что в ней хранится

**users (пользователи)**  
Нужна, чтобы знать, **кто** создал sql_query. Хранит: id, имя, email, роль. У каждого пользователя не больше одного sql_query в статусе «черновик» (корзина). Без таблицы пользователей нельзя привязать sql_query к разным людям.

**services (услуги)**  
Справочник **типов индексов** (таблица + индекс, размер таблицы). Хранит: id, название, описание, статус, ссылки на картинку и GIF, размер таблицы, время (speed). Показывается на главной и в «Добавить в запрос». Наполняется через Adminer (или seed).

**requests (sql_query)**  
Одна строка = **один sql_query** (симуляция запроса): черновик (корзина), удалённый, завершённый и т.д. Хранит: id, статус, даты, создателя, описание запроса текстом (query_description), селективность (м-м), результат — время и память запроса (result_time, result_memory). В корзину через м-м добавляются выбранные услуги.

**request_services (sql_query–услуга, м-м)**  
Связывает sql_query и услуги: **какие индексы и в каком количестве** входят в запрос. Хранит: пару (request_id, service_id), название, размер таблицы, селективность (м-м), количество, порядок, рассчитанное время по позиции (результат в м-м). Одна и та же услуга в одном sql_query — одна строка с увеличенным quantity.

---

## Сущности (таблицы)

### 1. **users** (Пользователи)

| Столбец | Тип           | Длина | Not Null | PK | FK | Описание        |
|---------|---------------|-------|----------|----|----|-----------------|
| id      | INTEGER       | —     | ✓        | ✓  |    | Идентификатор   |
| name    | VARCHAR       | 255   | ✓        |    |    | Имя             |
| email   | VARCHAR       | 255   | —        |    |    | Уникальный      |
| role    | VARCHAR       | 32    | ✓        |    |    | Роль (user/…)   |

---

### 2. **services** (Услуги / типы индексов)

| Столбец     | Тип    | Длина | Not Null | PK | FK | Описание              |
|-------------|--------|-------|----------|----|----|-----------------------|
| id          | VARCHAR| 64    | ✓        | ✓  |    | Идентификатор (напр. btree-10k) |
| name        | VARCHAR| 255   | ✓        |    |    | Наименование индекса  |
| description | TEXT   | —     | ✓        |    |    | Описание              |
| status      | VARCHAR| 32   | ✓        |    |    | active / deleted      |
| image_key   | VARCHAR| 255  | —        |    |    | URL/ключ изображения (nullable) |
| gif_key     | VARCHAR| 255  | —        |    |    | URL/ключ GIF (nullable) |
| table_size  | VARCHAR| 64   | ✓        |    |    | Размер таблицы        |
| speed       | VARCHAR| 64   | ✓        |    |    | Время (напр. 0.3ms)   |

---

### 3. **requests** (sql_query)

| Столбец            | Тип      | Длина | Not Null | PK | FK      | Описание                |
|--------------------|-----------|-------|----------|----|---------|--------------------------|
| id                 | INTEGER   | —     | ✓        | ✓  |         | Идентификатор            |
| status             | VARCHAR   | 32    | ✓        |    |         | draft/deleted/formed/completed/rejected |
| created_at         | TIMESTAMP | —     | ✓        |    |         | Дата создания            |
| created_by_id      | INTEGER   | —     | ✓        |    | →users  | Создатель                |
| formed_at          | TIMESTAMP | —     | —        |    |         | Дата формирования       |
| finished_at        | TIMESTAMP | —     | —        |    |         | Дата завершения          |
| moderator_id       | INTEGER   | —     | —        |    | →users  | Модератор                |
| query_description  | TEXT      | —     | —        |    |         | Описание запроса (Заявка)|
| selectivity        | NUMERIC   | —     | ✓        |    |         | Селективность (м-м)      |
| result_time        | VARCHAR   | 64    | —        |    |         | Время запроса            |
| result_memory      | VARCHAR   | 64    | —        |    |         | Память запроса           |

**Внешние ключи:**  
- created_by_id → users(id)  
- moderator_id → users(id)  
Каскадное удаление **не** используется.

---

### 4. **request_services** (Связь sql_query–услуга, м-м)

| Столбец            | Тип     | Длина | Not Null | PK | FK        | Описание                |
|--------------------|---------|-------|----------|----|-----------|-------------------------|
| request_id         | INTEGER | —     | ✓        | ✓  | →requests | sql_query               |
| service_id         | VARCHAR | 64    | ✓        | ✓  | →services | Услуга (таблица+индекс) |
| service_name       | VARCHAR | 255   | ✓        |    |           | Дублирование названия   |
| table_size         | VARCHAR | 64    | ✓        |    |           | Размер таблицы          |
| selectivity        | NUMERIC | —     | ✓        |    |           | Селективность (м-м)    |
| image_key          | VARCHAR | 255   | —        |    |           | Картинка                |
| quantity           | INTEGER | —     | ✓        |    |           | Количество              |
| position           | INTEGER | —     | ✓        |    |           | Порядок в sql_query     |
| is_main            | BOOLEAN | —     | ✓        |    |           | Главная позиция         |
| calculated_time_ms | NUMERIC | (10,2)| —        |    |           | Время (результат в м-м)  |

**Составной первичный ключ:** (request_id, service_id).  
**Внешние ключи:**  
- request_id → requests(id)  
- service_id → services(id)  
Каскадное удаление **не** используется.

---

## Связи (для ER-диаграммы)

| Связь                    | Тип        | Участники              | Описание                          |
|--------------------------|------------|------------------------|-----------------------------------|
| users — requests         | 1 : N      | users (1) — requests (N) | Один пользователь — много sql_query. Со стороны requests: created_by_id, moderator_id. |
| requests — request_services | 1 : N   | requests (1) — request_services (N) | Один sql_query — много позиций. FK: request_id. |
| services — request_services | 1 : N  | services (1) — request_services (N) | Одна услуга может быть в многих sql_query. FK: service_id. |
| request_services        | —          | —                      | Связующая таблица м-м между requests и services. |

Итог: **users** и **services** — независимые сущности; **requests** связана с **users** (создатель, модератор); **request_services** реализует связь **requests** ↔ **services** (м-м) с атрибутами quantity, position, is_main, calculated_time_ms.

---

## Краткая схема для рисования в StarUML

1. **Users**  
   id (PK), name, email, role.

2. **Services**  
   id (PK), name, description, status, image_key, gif_key, table_size, speed.

3. **Requests** (sql_query)  
   id (PK), status, created_at, created_by_id (FK→Users), formed_at, finished_at, moderator_id (FK→Users), query_description, selectivity, result_time, result_memory.

4. **Request_services**  
   (request_id (PK,FK→Requests), service_id (PK,FK→Services)), service_name, table_size, selectivity, image_key, quantity, position, is_main, calculated_time_ms.

5. Связи:  
   - Users 1 ——< Requests (по created_by_id);  
   - Users 1 ——< Requests (по moderator_id, опционально);  
   - Requests 1 ——< Request_services;  
   - Services 1 ——< Request_services.

У связей со стороны Requests и Services выберите кратность «много» (N), каскадное удаление не использовать.
