# Yandex Browser Blocker: Active protection
# Блокиратор (Запрещатор) Яндекс Браузера: Активная защита

## Install libs
## Установка библиотек

```cmd
go mod tidy
```

## Compile
## Компиляция

For Common (Without console) Program:

Для Обычной (Без консольной) Программы:

```cmd
go build -ldflags="-H windowsgui"
```
Flag `-ldflags="-H windowsgui"` allows hide the console

Флаг `-ldflags="-H windowsgui"` позволяет скрыть консоль

For Debug (With console) Program:

Для Дебаг (С консолью) Программы:
```cmd
go build
```

Please give a star)

Пожалуйста дайте звездочку)


#### Создать это меня вдохновил автор "Марк Аддерли" с его "MakuYan", по факту я просто украл его идею и сделал свое, но главное что СВОЕ
