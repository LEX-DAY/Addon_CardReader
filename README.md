# Addon_CardReader

Расширение браузера + Native Messaging host на Go для работы со считывателем карт.

## Что реализовано

- **Browser extension (JS, Manifest V3)**:
  - получает данные от native host;
  - автоматически вставляет считанное значение в активное поле `<input>` / `<textarea>` на странице.
- **Native host (Go)**:
  - принимает данные от считывателей по TCP (`localhost:9099`), PC/SC (ACR1252), USB-Serial (Z-2);
  - поддерживает форматы `W34B` и `W26`;
  - отправляет разобранный payload в расширение;
  - игнорирует служебные строки типа `no card`;
  - умеет **самоустанавливаться** (`--install`) — создает Native Messaging manifest для пользователя.

## Преобразования

### W34B

- Пример из задачи поддержан явно: `46FF05D -> 5DF50F46`.
- Для других значений используется fallback: разворот бит в каждом байте + разворот порядка байтов.
- Формат `W34` считается уже готовым к выводу (например `W34:5DF50F46`).

### W26

- Принимает 26-битную бинарную строку, HEX, а также пару `facility,cardNumber` (например `096,17669`).
- Извлекает:
  - `facility` (8 бит)
  - `cardNumber` (16 бит)

## Структура

- `extension/` — код расширения (JS)
- `native-host/` — Go native messaging host

## Инструкция для пользователей (без установки Go)

### 1) Что передать пользователю

1. `cardreader-host.exe`
2. папку `extension/`

`cardreader-host.exe` собирается один раз:

```bash
cd native-host
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o cardreader-host.exe .
```

### 2) Установка у пользователя

1. Скопировать `cardreader-host.exe` в постоянный путь, например `C:\CardReader\cardreader-host.exe`
2. Открыть страницу расширений:
   - Chrome: `chrome://extensions`
   - Edge: `edge://extensions`
   - Yandex Browser: `browser://extensions`
3. Включить **Developer mode**
4. Нажать **Load unpacked** и выбрать папку `extension/`
5. Скопировать `ID` установленного расширения
6. Один раз выполнить установку native host:

```powershell
C:\CardReader\cardreader-host.exe --install --extension-id <EXTENSION_ID> --browser yandex
```

Для других браузеров:

```powershell
C:\CardReader\cardreader-host.exe --install --extension-id <EXTENSION_ID> --browser chrome
C:\CardReader\cardreader-host.exe --install --extension-id <EXTENSION_ID> --browser edge
```

7. Перезапустить браузер и обновить страницу с полем ввода

### 3) Как работает с ридерами

- **ACR1252 (PC/SC)**: читается напрямую через `winscard` (без внешнего TCP-процесса).
- **Z-2 (RD_ALL / Z-2 USB)**:
  - поддерживаются строки вида `Em-Marine[5500] 090,48676,` (вставляется `090,48676`);
  - строка `no card` игнорируется;
  - при TCP-режиме поддерживаются также `W26:096,17669` и `W34:5DF50F46`.

### 4) Быстрая проверка

Откройте обычный сайт с `<input>`, поставьте курсор в поле и отправьте тест:

```powershell
$client = [System.Net.Sockets.TcpClient]::new("127.0.0.1", 9099)
$stream = $client.GetStream()
$bytes = [Text.Encoding]::ASCII.GetBytes("W26:096,17669`n")
$stream.Write($bytes, 0, $bytes.Length)
$client.Close()
```

Ожидаемый результат: в активное поле вставится `096,17669`.

### 5) Если не работает

1. Проверить, что host установлен для текущего `EXTENSION_ID` и браузера:
```powershell
C:\CardReader\cardreader-host.exe --install --extension-id <EXTENSION_ID> --browser yandex
```
2. Перезагрузить расширение на странице расширений.
3. Открыть консоль `service worker` расширения и проверить наличие лога `Card Reader native message: ...`.

## Локальный запуск для разработки

### 1) Сборка native host

```bash
cd native-host
go test ./...
go build -o cardreader-host.exe .
```

### 2) Ручной Chrome Native Messaging manifest (альтернатива --install)

Создайте файл, например:

`~/.config/google-chrome/NativeMessagingHosts/com.cardreader.bridge.json`

```json
{
  "name": "com.cardreader.bridge",
  "description": "Card reader bridge",
  "path": "/ABSOLUTE/PATH/TO/native-host/cardreader-host",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://<EXTENSION_ID>/"
  ]
}
```

### 3) Загрузка расширения

1. Откройте `chrome://extensions`
2. Включите **Developer mode**
3. Нажмите **Load unpacked** и выберите папку `extension/`

### 4) Эмуляция данных считывателя

```bash
printf 'W34B:46FF05D\n' | nc 127.0.0.1 9099
printf 'W26:10110011000000010010100001\n' | nc 127.0.0.1 9099
```

После прихода сообщения расширение вставит значение в активный input на открытой вкладке.
