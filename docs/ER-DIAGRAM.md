# ER-диаграмма (структура для StarUML)

Предметная область: **SQL Index Simulator** — заявки на симуляцию запросов с выбором индексов (услуг).

---

## Зачем нужна каждая таблица и что в ней хранится

**users (пользователи)**  
Нужна, чтобы знать, **кто** создал заявку и кто её завершил (модератор). Хранит: id, имя, email, роль. По заданию у каждого пользователя не больше одной заявки в статусе «черновик» — это проверяется по связи заявки с пользователем (created_by_id). Без таблицы пользователей нельзя привязать заявки к разным людям и различать «свою» корзину.

**services (услуги)**  
Справочник **типов индексов** (B-tree, Hash, GIN и т.д.), которые можно добавить в заявку. Хранит: id, название, описание, статус (действует/удалён), ссылки на картинку и GIF, размер таблицы, время (speed). Это то, что показывается на главной странице и в карточке «Добавить в заявку». Наполняется через Adminer (или seed), приложение только читает и использует при расчёте формулы (speed).

**requests (заявки)**  
Одна строка = **одна заявка** пользователя (симуляция запроса): черновик, удалённая, сформированная, завершённая или отклонённая. Хранит: id, статус, даты создания/формирования/завершения, создателя и модератора, селективность, рассчитанные время и память. Черновик по сути и есть «корзина» — в неё через м-м добавляются выбранные услуги. После «Завершить заявку» сюда записываются результат расчёта (result_time, result_memory) и дата завершения.

**request_services (заявка–услуга, м-м)**  
Связывает заявки и услуги: **какие индексы и в каком количестве** входят в заявку. Хранит: пару (request_id, service_id), дублирующие поля для отображения (название, размер, селективность, картинка), количество, порядок, признак «главный», рассчитанное время по позиции (calculated_time_ms). Одна и та же услуга в одной заявке — одна строка с увеличенным quantity. Без этой таблицы нельзя было бы хранить несколько услуг в одной заявке и считать по формуле вклад каждой позиции.

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

### 3. **requests** (Заявки)

| Столбец       | Тип      | Длина | Not Null | PK | FK      | Описание           |
|---------------|-----------|-------|----------|----|---------|--------------------|
| id            | INTEGER   | —     | ✓        | ✓  |         | Идентификатор      |
| status        | VARCHAR   | 32    | ✓        |    |         | draft/deleted/formed/completed/rejected |
| created_at    | TIMESTAMP | —     | ✓        |    |         | Дата создания      |
| created_by_id | INTEGER   | —     | ✓        |    | →users  | Создатель          |
| formed_at     | TIMESTAMP | —     | —        |    |         | Дата формирования  |
| finished_at   | TIMESTAMP | —     | —        |    |         | Дата завершения    |
| moderator_id  | INTEGER   | —     | —        |    | →users  | Модератор          |
| selectivity   | NUMERIC   | —     | ✓        |    |         | Селективность      |
| result_time   | VARCHAR   | 64    | —        |    |         | Рассчитанное время |
| result_memory | VARCHAR   | 64    | —        |    |         | Рассчитанная память|

**Внешние ключи:**  
- created_by_id → users(id)  
- moderator_id → users(id)  
Каскадное удаление **не** используется.

---

### 4. **request_services** (Связь заявка–услуга, м-м)

| Столбец            | Тип     | Длина | Not Null | PK | FK        | Описание                |
|--------------------|---------|-------|----------|----|-----------|-------------------------|
| request_id         | INTEGER | —     | ✓        | ✓  | →requests | Заявка                  |
| service_id         | VARCHAR | 64    | ✓        | ✓  | →services | Услуга                  |
| service_name       | VARCHAR | 255   | ✓        |    |           | Дублирование названия   |
| table_size         | VARCHAR | 64    | ✓        |    |           | Дублирование размера    |
| selectivity        | NUMERIC | —     | ✓        |    |           | Селективность по позиции|
| image_key          | VARCHAR | 255   | —        |    |           | Картинка                |
| quantity           | INTEGER | —     | ✓        |    |           | Количество              |
| position           | INTEGER | —     | ✓        |    |           | Порядок в заявке       |
| is_main            | BOOLEAN | —     | ✓        |    |           | Главная позиция        |
| calculated_time_ms | NUMERIC | (10,2)| —        |    |           | Рассчитанное время (при завершении) |

**Составной первичный ключ:** (request_id, service_id).  
**Внешние ключи:**  
- request_id → requests(id)  
- service_id → services(id)  
Каскадное удаление **не** используется.

---

## Связи (для ER-диаграммы)

| Связь                    | Тип        | Участники              | Описание                          |
|--------------------------|------------|------------------------|-----------------------------------|
| users — requests         | 1 : N      | users (1) — requests (N) | Один пользователь — много заявок. Со стороны requests: created_by_id, moderator_id. |
| requests — request_services | 1 : N   | requests (1) — request_services (N) | Одна заявка — много позиций. FK: request_id. |
| services — request_services | 1 : N  | services (1) — request_services (N) | Одна услуга может быть в многих заявках. FK: service_id. |
| request_services        | —          | —                      | Связующая таблица м-м между requests и services. |

Итог: **users** и **services** — независимые сущности; **requests** связана с **users** (создатель, модератор); **request_services** реализует связь **requests** ↔ **services** (м-м) с атрибутами quantity, position, is_main, calculated_time_ms.

---

## Краткая схема для рисования в StarUML

1. **Users**  
   id (PK), name, email, role.

2. **Services**  
   id (PK), name, description, status, image_key, gif_key, table_size, speed.

3. **Requests**  
   id (PK), status, created_at, created_by_id (FK→Users), formed_at, finished_at, moderator_id (FK→Users), selectivity, result_time, result_memory.

4. **Request_services**  
   (request_id (PK,FK→Requests), service_id (PK,FK→Services)), service_name, table_size, selectivity, image_key, quantity, position, is_main, calculated_time_ms.

5. Связи:  
   - Users 1 ——< Requests (по created_by_id);  
   - Users 1 ——< Requests (по moderator_id, опционально);  
   - Requests 1 ——< Request_services;  
   - Services 1 ——< Request_services.

У связей со стороны Requests и Services выберите кратность «много» (N), каскадное удаление не использовать.
