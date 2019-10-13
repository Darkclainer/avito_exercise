Сие есть реализация тестого задания для стажировки в авито: https://github.com/avito-tech/backend-trainee-assignment

Маленькие коментарии:
1. ID в JSON передаётся в виде числа, а не строки как в оригинале. Данное поведение меняется одной строчкой.
2. Для хранения данных был использован sqlite3

Ниже оригинальный текст задания. 

# Тестовое задание на позицию стажера-бекендера

Цель задания – разработать чат-сервер, предоставляющий HTTP API для работы с чатами и сообщениями пользователя.

Детали реализации:

* Писать код можно на любом языке программирования
* В качестве хранилища данных можно использовать любую технологию
* При перезапуске сервера добавленные данные должны сохраняться
* Сервер должен быть доступен на порту 9000
* Визуализация данных в виде пользовательского интерфейса (веб-приложение, мобильное приложение) не требуется – достаточно только обозначенного ниже API, доступного из командной строки. Однако простор фантазии не ограничиваем, покуда соблюдаются основные требования
* Предоставить инструкцию по запуску приложения. В идеале (но не обязательно) – использовать контейнеризацию с возможностью запустить проект командой `docker-compose up`
* Финальную версию нужно выложить на github.com

## Основные сущности

Ниже перечислены основные сущности, которыми должен оперировать сервер.

### User

Пользователь приложения. Имеет следующие свойства:

* **id** - уникальный идентификатор пользователя
* **username** - уникальное имя пользователя
* **created_at** - время создания пользователя

### Chat

Отдельный чат. Имеет следующие свойства:

* **id** - уникальный идентификатор чата
* **name** - уникальное имя чата
* **users** - список пользователей в чате, отношение многие-ко-многим
* **created_at** - время создания

### Message

Сообщение в чате. Имеет следующие свойства:

* **id** - уникальный идентификатор сообщения
* **chat** - ссылка на идентификатор чата, в который было отправлено сообщение
* **author** - ссылка на идентификатор отправителя сообщения, отношение многие-к-одному
* **text** - текст отправленного сообщения
* **created_at** - время создания

## Основные API методы

Методы обрабатывают HTTP POST запросы c телом, содержащим все необходимые параметры в JSON.

### Добавить нового пользователя

Запрос:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username": "user_1"}' \
  http://localhost:9000/users/add
```

Ответ: `id` созданного пользователя или HTTP-код ошибки.

### Создать новый чат между пользователями

Запрос:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"name": "chat_1", "users": ["<USER_ID_1>", "<USER_ID_2>"]}' \
  http://localhost:9000/chats/add
```

Ответ: `id` созданного чата или HTTP-код ошибки.

Количество пользователей не ограничено.

### Отправить сообщение в чат от лица пользователя

Запрос:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"chat": "<CHAT_ID>", "author": "<USER_ID>", "text": "hi"}' \
  http://localhost:9000/messages/add
```

Ответ: `id` созданного сообщения или HTTP-код ошибки.

### Получить список чатов конкретного пользователя

Запрос:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"user": "<USER_ID>"}' \
  http://localhost:9000/chats/get
```

Ответ: cписок всех чатов со всеми полями, отсортированный по времени создания последнего сообщения в чате (от позднего к раннему). Или HTTP-код ошибки.

### Получить список сообщений в конкретном чате

Запрос:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"chat": "<CHAT_ID>"}' \
  http://localhost:9000/messages/get
```

Ответ: список всех сообщений чата со всеми полями, отсортированный по времени создания сообщения (от раннего к позднему). Или HTTP-код ошибки.
