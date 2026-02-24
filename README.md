# Addon_CardReader

Расширение браузера + Native Messaging host на Go для работы со считывателем карт.

## Что реализовано

- **Browser extension (JS, Manifest V3)**:
  - получает данные от native host;
  - автоматически вставляет считанное значение в активное поле `<input>` / `<textarea>` на странице.
- **Native host (Go)**:
  - принимает данные от внешнего считывателя по TCP (`localhost:9099`), формат строки: `FORMAT:DATA`;
  - поддерживает форматы `W34B` и `W26`;
  - отправляет разобранный payload в расширение;
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

## ACR1252: можно ли без дополнительных приложений на ПК?

Короткий ответ: **в большинстве случаев нет**.

- ACR1252 обычно работает через CCID/PCSC стек драйверов ОС, а браузерное расширение не имеет прямого доступа к такому интерфейсу.
- Поэтому нужен локальный bridge-процесс (в этом проекте это `native-host/cardreader-host`/`cardreader-host.exe`) и Native Messaging.
- При этом вручную запускать host каждый раз не нужно: Chrome поднимает его автоматически при `connectNative`.

Если считыватель переключен в keyboard-wedge режим (эмуляция клавиатуры), можно работать без host, но тогда теряется логика декодирования в Go и контроль протокола.

## Готовый exe для пользователя (без установки Go)

Вы можете собрать exe один раз и отдать пользователю архив:

```bash
cd native-host
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o cardreader-host.exe .
```

Пользователю нужен только:
1. `cardreader-host.exe`
2. папка `extension/`

### Установка на ПК пользователя

1. Скопировать `cardreader-host.exe`, например в `C:\CardReader\cardreader-host.exe`
2. Загрузить расширение в Chrome через `chrome://extensions` -> **Load unpacked** (`extension/`)
3. Узнать `EXTENSION_ID` расширения
4. Один раз выполнить установку manifest:

```powershell
C:\CardReader\cardreader-host.exe --install --extension-id cjhjdbhjocikbieijgiamhekhplaefge --browser yandex
```

После этого host будет запускаться Chrome автоматически.

## Локальный запуск для разработки

### 1) Сборка native host

```bash
cd native-host
go test ./...
go build -o cardreader-host .
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
