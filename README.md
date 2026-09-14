<p align="center">
  <img src=".github/logo.svg" alt="AniLiberty Indexer Logo" width="192" height="192" />
</p>

# AniLiberty Indexer

Легковесный Torznab-совместимый прокси-сервер для интеграции [AniLiberty](https://anilibria.top/) (AniLibria) API с Prowlarr, Sonarr и Radarr.

## Содержание

- [Особенности](#особенности)
- [Принцип работы](#принцип-работы)
- [Переменные окружения](#переменные-окружения)
- [Настройка шаблона названия](#настройка-шаблона-названия)
- [Запуск в Docker](#запуск-в-docker)
- [Запуск Go-бинарника](#запуск-go-бинарника)
- [Добавление и настройка в Prowlarr](#добавление-и-настройка-в-prowlarr)
- [Спецификация API](#спецификация-api)
  - [Авторизация](#авторизация)
  - [Обзор эндпоинтов](#обзор-эндпоинтов)
  - [Подробное описание эндпоинтов](#подробное-описание-эндпоинтов)

## Особенности

- **Продвинутый парсинг**: Точное автоматическое извлечение номеров сезонов и серий из названий и меток торрентов.
- **Гибкие шаблоны**: Настраиваемый формат заголовков раздач с автоматической очисткой от пустых скобок и артефактов.
- **Извлечение метаданных**: Распознавание года, разрешения (1080p, 720p), типа источника (BDRip, WEBRip) и кодека (x264, x265).
- **Стандарт Torznab**: Полноценная поддержка Torznab API для работы с Prowlarr, Sonarr и Radarr (поиск `search`/`tvsearch`/`movie`, RSS-лента и capabilities).
- **Высокая производительность**: Написан на Go, скомпилирован в единый легкий бинарник без внешних зависимостей.
- **Поддержка прокси**: Встроенное проксирование скачивания `.torrent` файлов и поддержка исходящих HTTP/SOCKS5 прокси.

## Принцип работы

1. **Запрос**: Принимает Torznab-запрос от Prowlarr/Sonarr/Radarr.
2. **Поиск**: Запрашивает релизы и торренты через API AniLiberty.
3. **Парсинг и адаптация**: Извлекает сезон, серии, качество, кодек и формирует название по шаблону.
4. **Проксирование**: Перенаправляет скачивание `.torrent` файлов через эндпоинт `/download/{hash}`.

---

## Переменные окружения

| Переменная                     | По умолчанию                      | Описание                                                                                                                     |
| ------------------------------ | --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `HOST`                         | `0.0.0.0`                         | Адрес интерфейса, к которому привязывается сервер.                                                                           |
| `PORT`                         | `3649`                            | Внутренний порт, на котором работает сервер.                                                                                 |
| `TORRENT_TITLE_TEMPLATE`       | [Примеры](#примеры-использования) | Кастомный шаблон для формирования названия торрента для сериалов. [Настройка шаблона названия](#настройка-шаблона-названия). |
| `TORRENT_MOVIE_TITLE_TEMPLATE` | [Примеры](#примеры-использования) | Кастомный шаблон для формирования названия торрента для фильмов. [Настройка шаблона названия](#настройка-шаблона-названия).  |
| `ANILIBERTY_API`               | `https://anilibria.top/api/v1`    | Адрес API AniLiberty.                                                                                                        |
| `ANILIBERTY_SITE`              | `https://anilibria.top`           | Основной адрес сайта для генерации внешних ссылок на релизы.                                                                 |
| `API_KEY`                      | _(пусто)_                         | Опциональный секретный ключ для авторизации клиентов.                                                                        |
| `UPSTREAM_PROXY`               | _(пусто)_                         | Опциональный URL прокси-сервера (HTTP, HTTPS, SOCKS5) для запросов к AniLiberty API.                                         |

## Настройка шаблона названия

| Плейсхолдер           | Пример значения                     | Описание                                                                                          |
| --------------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------- |
| `{title_latin_clean}` | `Monogatari Series`                 | Латинское (ромадзи/оригинальное) название релиза без информации о сезонах и годах.                |
| `{title_latin}`       | `Monogatari Series: Second Season`  | Оригинальное латинское название релиза в сыром виде из API.                                       |
| `{title_ru}`          | `Цикл Историй: Второй Сезон`        | Русское название релиза в сыром виде из API.                                                      |
| `{season}`            | `2`                                 | Номер сезона в виде числа.                                                                        |
| `{episodes}`          | `E01-E23` или `E01`                 | Серии, извлеченные из описания торрента. Если серии не определены - пусто.                        |
| `{season_episodes}`   | `S02E01-S02E23`, `S02E01` или `S02` | Сезон и серии, извлеченные из описания торрента. Если серии не определены - выводит только сезон. |
| `{year}`              | `2013`                              | Год выхода релиза (если не определен - пусто).                                                    |
| `{type}`              | `WEBRip`                            | Тип источника.                                                                                    |
| `{quality}`           | `1080p`                             | Разрешение.                                                                                       |
| `{codec}`             | `x264`                              | Нормализованный кодек.                                                                            |

- _Индексатор автоматически очищает результирующее название от артефактов, возникающих из-за пустых плейсхолдеров. Конструкции вида `()`, `[]`, `{}` автоматически удаляются. Например, если `{year}` отсутствует, шаблон `({year})` превратится в пустые скобки `()` и удалится автоматически._

### Примеры использования

1. **По умолчанию**:

   ```env
   TORRENT_TITLE_TEMPLATE="[AniLiberty] {title_latin_clean} - S{season} [RUS][{type} {quality} {codec}] ({year})"
   ```

   Результат для сериала:
   `[AniLiberty] Monogatari Series - S2 [RUS][BDRip 1080p x264] (2013)`

   ```env
   TORRENT_MOVIE_TITLE_TEMPLATE="[AniLiberty] {title_latin_clean} ({year}) [RUS][{type} {quality} {codec}]"
   ```

   Результат для фильма:
   `[AniLiberty] Gekijouban Steins;Gate: Fuka Ryouiki no Deja vu (2013) [RUS][BDRip 1080p x264]`

2. **С сериями**:
   ```env
   TORRENT_TITLE_TEMPLATE="[AniLiberty] {title_latin_clean} - {season_episodes} [RUS][{type} {quality} {codec}] ({year})"
   ```
   Результат для сериала:
   `[AniLiberty] Monogatari Series - S02E01-S02E23 [RUS][BDRip 1080p x264] (2013)`

---

## Запуск в Docker

1. Скачайте файлы `docker-compose.yml` и `.env.example` из репозитория.
2. Создайте файл конфигурации `.env` из примера:
   ```bash
   cp .env.example .env
   ```
   _Отредактируйте параметры в файле `.env` при необходимости._
3. Запустите контейнер:
   ```bash
   docker compose up -d
   ```

## Запуск Go-бинарника

Вы можете скачать уже готовый скомпилированный бинарник для вашей ОС (Linux, Windows, macOS) из раздела [Releases](https://github.com/retrilzzy/aniliberty-indexer/releases/latest) или собрать его самостоятельно:

1. Соберите бинарник:
   ```bash
   go build -o aniliberty-indexer ./cmd/indexer/
   ```
2. Создайте файл конфигурации `.env` из примера:
   ```bash
   cp .env.example .env
   ```
   _Отредактируйте параметры в файле `.env` при необходимости._
3. Запустите его:
   ```bash
   ./aniliberty-indexer
   ```
4. **Опциональная настройка службы systemd (для Linux)**:
   Чтобы сервер работал в фоновом режиме и автоматически запускался при старте системы, создайте файл службы `/etc/systemd/system/aniliberty-indexer.service`:

   ```ini
   [Unit]
   Description=AniLiberty Torznab Indexer
   After=network.target

   [Service]
   Type=simple
   User=nobody
   WorkingDirectory=/opt/aniliberty-indexer
   ExecStart=/opt/aniliberty-indexer/aniliberty-indexer
   Restart=always
   EnvironmentFile=/opt/aniliberty-indexer/.env

   [Install]
   WantedBy=multi-user.target
   ```

   После этого выполните команды для регистрации и запуска службы:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now aniliberty-indexer
   ```

---

## Добавление и настройка в Prowlarr

1. Перейдите в веб-панель Prowlarr.
2. Откройте вкладку **Settings** -> **Indexers**.
3. Нажмите **Add Indexer** и выберите **Generic Torznab** (for torrents).
4. Введите параметры:
   - **Name**: `AniLiberty`
   - **URL**: `http://<ip_or_hostname>:3649`.
   - **API Path**: `/api`
   - **API Key**: _(оставьте пустым, если `API_KEY` не задан)_.
5. Нажмите кнопку **Test**, после сохраните настройки (**Save**).

---

# Спецификация API

## Обзор эндпоинтов

| Метод | Эндпоинт                             | Авторизация | Описание                                   |
| ----- | ------------------------------------ | ----------- | ------------------------------------------ |
| `GET` | `/` / `/health`                      | Нет         | Health check статус сервера                |
| `GET` | `/api` / `/torznab` / `/torznab/api` | Требуется   | Torznab API (Capabilities, Search, Browse) |
| `GET` | `/download/{hash}`                   | Требуется   | Проксирование скачивания `.torrent` файла  |

## Авторизация

Если в переменной окружения `API_KEY` задан секретный ключ, авторизация становится обязательной для всех эндпоинтов, кроме `GET /` и `GET /health`.

Поддерживается 3 стандартных способа передачи ключа:

| Способ                   | Метод передачи                          | Пример                                  |
| ------------------------ | --------------------------------------- | --------------------------------------- |
| **Query Parameter**      | URL-параметр `apikey`                   | `GET /api?apikey=your_secret_key`       |
| **HTTP Header**          | Заголовок `X-Api-Key`                   | `X-Api-Key: your_secret_key`            |
| **Authorization Header** | `Authorization: Bearer` (или raw token) | `Authorization: Bearer your_secret_key` |

> [!WARNING]
> Если авторизация включена, а ключ не передан или недействителен, сервер вернет ответ `401 Unauthorized` в формате Torznab XML: `<error code="100" description="Incorrect user credentials"/>`.

---

## Подробное описание эндпоинтов

### 1. Health Check

Проверка работоспособности и доступности сервиса.

```http
GET /health
GET /
```

- **Авторизация**: Не требуется
- **Ответ (`200 OK`)**:

  ```http
  HTTP/1.1 200 OK
  Content-Type: text/plain; charset=utf-8

  OK
  ```

### 2. Torznab Capabilities

Возвращает поддерживаемые категории, типы поиска и ограничения индексатора.

```http
GET /api?t=caps
```

_(Поддерживаются также алиасы: `/torznab?t=caps`, `/torznab/api?t=caps`)_

#### Query-параметры

| Параметр | Тип      | Обязательный | Описание                                     |
| -------- | -------- | ------------ | -------------------------------------------- |
| `t`      | `string` | **Да**       | Тип операции. Фиксированное значение: `caps` |

#### Пример ответа (`200 OK`)

```http
HTTP/1.1 200 OK
Content-Type: application/xml; charset=utf-8
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<caps>
  <server version="1.0" title="AniLiberty"/>
  <limits max="50" default="25"/>
  <searching>
    <search available="yes" supportedParams="q"/>
    <tv-search available="yes" supportedParams="q,season,ep"/>
    <movie available="yes" supportedParams="q"/>
  </searching>
  <categories>
    <category id="5000" name="TV"/>
    <category id="5070" name="TV/Anime"/>
  </categories>
</caps>
```

### 3. Torznab Search & Browse

Поиск аниме-релизов по названию или получение ленты последних добавленных раздач (Browse mode).

```http
GET /api?t={t}&q={q}&limit={limit}&offset={offset}
```

_(Поддерживаются также алиасы: `/torznab`, `/torznab/api`)_

#### Query-параметры

| Параметр | Тип       | По умолчанию | Описание                                                                                                                                                     |
| -------- | --------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `t`      | `string`  | Не задан     | Тип поиска: `search`, `tvsearch`, `movie` (если не задан — режим Browse)                                                                                     |
| `q`      | `string`  | Не задан     | Поисковый запрос.<br>• **Задан**: поиск по релизу в API AniLiberty с параллельной загрузкой торрентов.<br>• **Отсутствует**: возвращает ленту свежих раздач. |
| `limit`  | `integer` | `25`         | Количество результатов на страницу (максимум `50`)                                                                                                           |
| `offset` | `integer` | `1`          | Страница (смещение) для пагинации                                                                                                                            |

#### Пример ответа (`200 OK`)

```http
HTTP/1.1 200 OK
Content-Type: application/xml; charset=utf-8
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:torznab="http://torznab.com/schemas/2015/feed">
    <channel>
        <title>AniLiberty</title>
        <link>https://anilibria.top</link>
        <description>AniLiberty Torznab Feed</description>
        <item>
            <title>[AniLiberty] Vinland Saga - S01 [RUS][WEBRip 1080p x265] (2019)</title>
            <guid>04241e64625a3b3fb8e2a5037d0f6bff0dd0a547</guid>
            <link>https://anilibria.top/anime/releases/release/vinland-saga</link>
            <comments>https://anilibria.top/anime/releases/release/vinland-saga</comments>
            <pubDate>Sat, 04 Jan 2020 21:00:00 UTC</pubDate>
            <size>18020309025</size>
            <description>Сага о Винланде / Vinland Saga...</description>
            <enclosure url="https://anilibria.top/api/v1/anime/torrents/04241e64625a3b3fb8e2a5037d0f6bff0dd0a547/file" length="18020309025" type="application/x-bittorrent"/>
            <torznab:attr name="magneturl" value="magnet:?xt=urn:btih:04241e64625a3b3fb8e2a5037d0f6bff0dd0a547&amp;dn=Vinland+Saga+-+AniLibria.TV+%5BWEBRip+1080p%5D&amp;xl=18020309025&amp;tr=http://tr.libria.fun:2710/announce&amp;tr=http://retracker.local/announce"/>
            <torznab:attr name="infohash" value="04241e64625a3b3fb8e2a5037d0f6bff0dd0a547"/>
            <torznab:attr name="category" value="5070"/>
            <torznab:attr name="seeders" value="20"/>
            <torznab:attr name="peers" value="24"/>
            <torznab:attr name="grabs" value="28901"/>
            <torznab:attr name="downloadvolumefactor" value="0"/>
            <torznab:attr name="uploadvolumefactor" value="1"/>
            <torznab:attr name="resolution" value="1080"/>
            <torznab:attr name="poster" value="/storage/releases/posters/8424/Pt0yaatoRpp2adNWT8CZgYkWmzyLo1kQ.webp"/>
        </item>
    </channel>
</rss>
```

### 4. Скачивание .torrent файла

Проксирует скачивание `.torrent` файла с серверов AniLiberty API. Используется, когда клиент или торрент-качалка (например, qBittorrent) не имеет прямого доступа к API AniLiberty.

```http
GET /download/{hash}?pk={pk}
```

#### Path-параметры

| Параметр | Тип      | Описание                                         |
| -------- | -------- | ------------------------------------------------ |
| `hash`   | `string` | Infohash раздачи или идентификатор торрент-файла |

#### Query-параметры

| Параметр | Тип      | Обязательный | Описание                                                 |
| -------- | -------- | ------------ | -------------------------------------------------------- |
| `pk`     | `string` | Нет          | Pass-through ключ приватного доступа API _(опционально)_ |

#### Пример ответа (`200 OK`)

```http
HTTP/1.1 200 OK
Content-Type: application/x-bittorrent
Content-Disposition: attachment; filename="04241e64625a3b3fb8e2a5037d0f6bff0dd0a547.torrent"

[Binary .torrent data stream]
```
