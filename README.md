# GophKeeper

GophKeeper представляет собой клиент-серверную систему, позволяющую пользователю безопасно хранить
логины, пароли, данные банковских карт, произвольные текстовые и бинарные данные.

Сервер поддерживает следующий функционал:
 * регистрация, аутентификация и авторизация пользователей;
 * хранение приватных данных пользователей;
 * синхронизация данных между несколькими авторизованными клиентами одного владельца;
 * передача приватных данных владельцу по запросу.

Клиент реализует следующую бизнес-логику:
 * регистрация, аутентификация и авторизация пользователей на удалённом сервере;
 * доступ к приватным данным по запросу.

## Подготовка

### Хранилище GophKeeper

Перед запуском системы необходимо создать БД для хранения данных GophKeeper.

#### Установка СУБД

В качестве СУБД можно использовать PostreSQL.

 * [Установка на Linux](https://ruvds.com/ru/helpcenter/postgresql-pgadmin-ubuntu/)
 * [Установка на Windows](https://winitpro.ru/index.php/2019/10/25/ustanovka-nastrojka-postgresql-v-windows/)
 * [Установка на macOS](https://wiki.postgresql.org/wiki/Russian/PostgreSQL-One-click-Installer-Guide)

#### Создание БД

После установки СУБД необходимо создать БД. Пример команд создания БД для Linux с использованием PostreSQL:

```
sudo -i -u postgres psql -c "create database gophkeeper;"
sudo -i -u postgres psql -c "create user gophkeeper with encrypted password '1234';"
sudo -i -u postgres psql -c "grant all privileges on database gophkeeper to gophkeeper;"
```

### Конфигурация сервера

Перед запуском пользователь может задать настройки сервера в виде конфигурационного файла или набора переменных окружения.
Расположение файла конфигурации: ./internal/config/server_config.yaml

Пример конфигурационного файла:

```
# server_config.yaml

server:
 address: "localhost:8080"
storage:
 connection_string: "postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable"
security:
 jwt_key: "yoursecretkey"
 expiration_time: "3h"
```

Пример настройки сервера через переменные окружения:

```
# .env

GOPHKEEPER_SERVER_ADDRESS=localhost:8080
GOPHKEEPER_STORAGE_CONNECTION_STRING=postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable
GOPHKEEPER_SECURITY_JWT_KEY=yoursecretkey
GOPHKEEPER_SECURITY_EXPIRATION_TIME=3h
```

В случае отсутствия конфигурационного файла и переменных окружения будут использованы значения по умолчанию: localhost:8080 и postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable

### Конфигурация клиента

Перед запуском пользователь может задать настройки клиента в виде конфигурационного файла или набора переменных окружения.
Расположение файла конфигурации: ./internal/config/.gophkeeper.yaml либо $HOME/.gophkeeper.yaml

Пример конфигурационного файла:

```
# .gophkeeper.yaml

grpc_address: "localhost:8080"
encryption_key: "gophkeeperclient"
```

Пример настройки сервера через переменные окружения:

```
# .env

GRPC_ADDRESS=localhost:8080
GOPHKEEPER_ENCRYPTION_KEY=yourencryptionkey
```

## Запуск сервера

1. Сборка бинарного файла сервера c информацией о версии и дате сборки:

```
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o ./cmd/server/gophkeeper_server ./cmd/server/main.go
```

2. Запуск сервера:

./cmd/server/gophkeeper_server

## Сборка клиентского приложения

Сборка клиента для всех платформ c информацией о версии и дате сборки бинарного файла:

```
chmod +x ./scripts/build_client.sh

./scripts/build_client.sh 1.0.0
```

## Функции GophKeeper

### Регистрация и авторизации

Для регистрации пользователя необходимо указать логин и пароль.
Пример команды регистрации:

```
./dist/gophkeeper-[os]-[arch] register -u user@mail.сom -p 1234
```

При успешной регистрации пользователя, сервер вернет токен доступа. Токен будет сохранен в текстовый файл token.txt.

Токен доступа можно также запросить с помощью команды авторизации:

```
./dist/gophkeeper-[os]-[arch] login -u user@mail.сom -p 1234
```

После успешной регистрации/авторизации, пользователь может управлять своими приватными данными.

### Сохранение приватных данных пользователя в GophKeeper

1. Пример команды добавления пары логин/пароль:

```
./dist/gophkeeper-[os]-[arch] secret create credentials \
  --login user@mail.com \
  --password 12345678 \
  --note mail.com
```

2. Пример команды добавления данных банковской карты:

```
./dist/gophkeeper-[os]-[arch] secret create card \
  --number 1234567812345678 \
  --date 02/25 \
  --holder Name \
  --code 000 \
  --note bank
```

3. Пример команды добавления текстовой информации:

```
./dist/gophkeeper-[os]-[arch] secret create text \
  --text "secret phrase" \
  --note secret
```

4. Пример команды добавления бинарных данных:

```
./dist/gophkeeper-[os]-[arch] secret create bin \
  -f gophkeeper \
  --note code
```

### Получение данных

Команда запроса списка всех приватных данных:

```
./dist/gophkeeper-[os]-[arch] secret all
```

### Обновление данных

Чтобы обновить приватные данные, используйте команду secret update с флагами, аналогичными команде secret create и дополнительным флагом --id.

Пример команды обновления данных учетной записи:

```
./dist/gophkeeper-[os]-[arch] secret update credentials \
  --id 1 \
  --login user@yandex.ru \
  --password 12345678 \
  --note yandex
```
