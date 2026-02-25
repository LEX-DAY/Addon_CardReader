# Addon_CardReader

## Установка (Windows)

1. Положите `cardreader-host.exe` в `C:\CardReader\cardreader-host.exe`.
2. Установите расширение из папки `extension/`:
   - Chrome: `chrome://extensions`
   - Edge: `edge://extensions`
   - Yandex Browser: `browser://extensions`
   - включите `Developer mode` -> `Load unpacked` -> выберите `extension/`.
3. Скопируйте `ID` расширения и выполните в PowerShell:

```powershell
C:\CardReader\cardreader-host.exe --install --extension-id kkffgncccmjpphlemnbllmbmjkchoodf --browser <chrome|edge|yandex>
```

## Запуск

1. Перезапустите браузер.
2. Откройте нужный сайт с полем ввода.
3. Поставьте курсор в поле и приложите карту.
