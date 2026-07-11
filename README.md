<p align="center">
  <img src=".github/logo.svg" alt="AniLiberty Indexer Logo" width="256" height="256" />
</p>

# AniLiberty Indexer

Легковесный Torznab-совместимый прокси-сервер для интеграции [AniLiberty](https://anilibria.top/) (AniLibria) API с Prowlarr и Sonarr. Прокси объединяет поиск релиза по ключевым словам и получение списка торрентов для него в один стандартный запрос Torznab XML.

## Содержание
- [Особенности](#особенности)
- [Переменные окружения](#переменные-окружения)
- [Настройка шаблона названия](#настройка-шаблона-названия)
- [Запуск в Docker](#запуск-в-docker)
- [Запуск Go-бинарника](#запуск-go-бинарника)
- [Добавление и настройка в Prowlarr](#добавление-и-настройка-в-prowlarr)
- [Спецификация API](#спецификация-api)
  - [Авторизация](#авторизация)
  - [Эндпоинты](#эндпоинты)
    - [Проверка работоспособности (Health check)](#проверка-работоспособности-health-check)
    - [Запрос возможностей индексатора (Torznab Capabilities)](#запрос-возможностей-индексатора-torznab-capabilities)
    - [Поиск и просмотр раздач (Torznab Search/Browse)](#поиск-и-просмотр-раздач-torznab-searchbrowse)
    - [Проксирование скачивания .torrent файлов](#проксирование-скачивания-torrent-файлов)

## Особенности
* **Продвинутый парсинг**: точное распознавание номеров сезонов и серий.
* **Гибкие шаблоны**: настраиваемый формат названий торрентов с очисткой от артефактов.
* **Извлечение метаданных**: автоматическое определение года, качества и кодека.
* **Torznab API**: полноценный поиск, RSS-лента и проксирование скачивания `.torrent` файлов.
* **Поддержка прокси**: встроенная поддержка HTTP/SOCKS5 прокси.
* **Минималистичность**: легкий бинарник без внешних зависимостей.


## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `HOST` | `0.0.0.0` | Адрес интерфейса, к которому привязывается сервер. |
| `PORT` | `3649` | Внутренний порт, на котором работает сервер. |
| `TORRENT_TITLE_TEMPLATE` | [Примеры](#примеры-использования) | Кастомный шаблон для формирования названия торрента для сериалов. [Настройка шаблона названия](#настройка-шаблона-названия). |
| `TORRENT_MOVIE_TITLE_TEMPLATE` | [Примеры](#примеры-использования) | Кастомный шаблон для формирования названия торрента для фильмов. [Настройка шаблона названия](#настройка-шаблона-названия). |
| `ANILIBERTY_API` | `https://anilibria.top/api/v1` | Адрес API AniLiberty. |
| `ANILIBERTY_SITE` | `https://anilibria.top` | Основной адрес сайта для генерации внешних ссылок на релизы. |
| `API_KEY` | *(пусто)* | Опциональный секретный ключ для авторизации клиентов. |
| `UPSTREAM_PROXY` | *(пусто)* | Опциональный URL прокси-сервера (HTTP, HTTPS, SOCKS5) для запросов к AniLiberty API. |


## Настройка шаблона названия

| Плейсхолдер | Пример значения | Описание |
|---|---|---|
| `{title_latin_clean}` | `Monogatari Series` | Латинское (ромадзи/оригинальное) название релиза без информации о сезонах и годах. |
| `{title_latin}` | `Monogatari Series: Second Season` | Оригинальное латинское название релиза в сыром виде из API. |
| `{title_ru}` | `Цикл Историй: Второй Сезон` | Русское название релиза в сыром виде из API. |
| `{season}` | `2` | Номер сезона в виде числа. |
| `{episodes}` | `E01-E23` или `E01` | Серии, извлеченные из описания торрента. Если серии не определены - пусто. |
| `{season_episodes}` | `S02E01-E23`, `S02E01` или `SO2` | Сезон и серии, извлеченные из описания торрента. Если серии не определены - выводит только сезон. |
| `{year}` | `2013` | Год выхода релиза (если не определен - пусто). |
| `{type}` | `WEBRip` | Тип источника. |
| `{quality}` | `1080p` | Разрешение. |
| `{codec}` | `x264` | Нормализованный кодек. |

* *Индексатор автоматически очищает результирующее название от артефактов, возникающих из-за пустых плейсхолдеров. Конструкции вида `()`, `[]`, `{}` автоматически удаляются. Например, если `{year}` отсутствует, шаблон `({year})` превратится в пустые скобки `()` и удалится автоматически.*

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
   `[AniLiberty] Monogatari Series - S02E01-E23 [RUS][BDRip 1080p x264] (2013)`


## Запуск в Docker

1. Склонируйте или скопируйте файлы репозитория.
2. Создайте файл конфигурации `.env` из примера:
   ```bash
   cp .env.example .env
   ```
   *Отредактируйте параметры в файле `.env` при необходимости.*
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
   *Отредактируйте параметры в файле `.env` при необходимости.*
3. Запустите его:
   ```bash
   ./aniliberty-indexer
   ```


## Добавление и настройка в Prowlarr

1. Перейдите в веб-панель Prowlarr.
2. Откройте вкладку **Settings** -> **Indexers**.
3. Нажмите **Add Indexer** и выберите **Generic Torznab** (for torrents).
4. Введите параметры:
   * **Name**: `AniLiberty`
   * **URL**: `http://<ip_or_hostname>:3649`.
   * **API Path**: `/api`
   * **API Key**: *(оставьте пустым, если `API_KEY` не задан)*.
5. Нажмите кнопку **Test**, после сохраните настройки (**Save**).


# Спецификация API

## Авторизация

Если в переменных окружения настроен параметр `API_KEY`, то для всех эндпоинтов (кроме `/health` и `/`) требуется передача ключа. Поддерживаются следующие способы передачи ключа:
1. **Query-параметр**: `?apikey=ВАШ_КЛЮЧ`
2. **HTTP-заголовок**: `X-Api-Key: ВАШ_КЛЮЧ`
3. **HTTP-заголовок**: `Authorization: Bearer ВАШ_КЛЮЧ` (или `Authorization: ВАШ_КЛЮЧ`)

## Эндпоинты

### Проверка работоспособности (Health check)
* **Путь**: `GET /` или `GET /health`
* **Авторизация**: Не требуется.

### Запрос возможностей индексатора (Torznab Capabilities)
Позволяет клиентам узнать поддерживаемые категории и параметры поиска.

* **Путь**: `GET /api?t=caps` *(или `/torznab?t=caps` / `/torznab/api?t=caps`)*
* **Параметры**:
  * `t=caps` (обязательный)

### Поиск и просмотр раздач (Torznab Search/Browse)
Поиск релизов по ключевому слову или получение последних добавленных раздач.

* **Путь**: `GET /api` *(или `/torznab` / `/torznab/api`)*
* **Параметры**:
  * `t` - тип операции. Поддерживаются `search`, `tvsearch`, `movie` (или пусто для просмотра ленты).
  * `q` - поисковый запрос (опционально).
    * Если параметр **задан**: выполняется поиск релизов по ключевому слову через API AniLiberty, затем параллельно запрашиваются торренты для каждого релиза.
    * Если параметр **отсутствует**: возвращается лента последних добавленных торрентов (Browse mode).
  * `limit` - максимальное количество элементов (по умолчанию `25`, максимум `50`).
  * `offset` - страница для пагинации (по умолчанию `1`).

**Пример ответа** (XML, HTTP 200):
```xml
<?xml
version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:torznab="http://torznab.com/schemas/2015/feed">
    <channel>
        <title>AniLiberty</title>
        <link>https://anilibria.top</link>
        <item>
            <title>[AniLiberty] Vinland Saga - S01 [RUS][WEBRip 1080p x265] (2019)</title>
            <guid>04241e64625a3b3fb8e2a5037d0f6bff0dd0a547</guid>
            <link>https://anilibria.top/anime/releases/release/vinland-saga</link>
            <comments>https://anilibria.top/anime/releases/release/vinland-saga</comments>
            <pubDate>Sat, 04 Jan 2020 21:00:00 UTC</pubDate>
            <size>18020309025</size>
            <description>Сага о Винланде / Vinland Saga / Время господства викингов. Людей, которые известны своими варварскими обычаями и жестокими убийствами. Торфинн - сын одного из величайших викингов. Вот только вырос мальчик без отца, ведь тот погиб на поле боя. Желая отомстить, Торфинн поклялся убить виновного. Однако юноше ещё только предстоит овладеть искусством битвы.</description>
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
        <item>
            <title>[AniLiberty] Vinland Saga - S02 [RUS][WEBRip 1080p x264] (2023)</title>
            <guid>477f7dc431698311420ec470fcad6ca145933f26</guid>
            <link>https://anilibria.top/anime/releases/release/vinland-saga-2</link>
            <comments>https://anilibria.top/anime/releases/release/vinland-saga-2</comments>
            <pubDate>Mon, 19 Jun 2023 21:50:47 UTC</pubDate>
            ...
        </item>
        ...
    </channel>
</rss>
```

### Проксирование скачивания .torrent файлов
Позволяет скачивать торрент-файл через прокси, если клиент не имеет прямого доступа к API AniLiberty.
* **Путь**: `GET /download/{hash}`
* **Параметры**:
  * `{hash}` в пути - инфохеш (или ID) раздачи.
  * `pk` (опционально) - pass-through параметр приватного ключа API.
* **Поведение**: Прокси отправляет запрос на `ANILIBERTY_API/anime/torrents/{hash}/file` и пересылает файл клиенту.
* **Пример ответа**: Бинарный поток торрент-файла с заголовком `Content-Type: application/x-bittorrent`.

