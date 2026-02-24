# Addon_CardReader

Расширение браузера + Native Messaging host на Go для работы со считывателем карт.

## Что реализовано

- **Browser extension (JS, Manifest V3)**:
  - получает данные от native host;
  - автоматически вставляет считанное значение в активное поле `<input>` / `<textarea>` на странице.
- **Native host (Go)**:
  - принимает данные от внешнего считывателя по TCP (`localhost:9099`), формат строки: `FORMAT:DATA`;
  - поддерживает форматы `W34B` и `W26`;
  - отправляет разобранный payload в расширение.

## Преобразования

### W34B

- Пример из задачи поддержан явно: `46FF05D -> 5DF50F46`.
- Для других значений используется fallback: разворот бит в каждом байте + разворот порядка байтов.

### W26

- Принимает 26-битную бинарную строку или HEX.
- Извлекает:
  - `facility` (8 бит)
  - `cardNumber` (16 бит)

## Структура

- `extension/` — код расширения (JS)
- `native-host/` — Go native messaging host

## Локальный запуск

### 1) Сборка native host

```bash
cd native-host
go test ./...
go build -o cardreader-host .
```

### 2) Chrome Native Messaging manifest

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

> Укажите реальный `EXTENSION_ID` после загрузки unpacked extension.

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
