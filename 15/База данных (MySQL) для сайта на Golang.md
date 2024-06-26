# База данных (MySQL) для сайта на Golang

---
Чтобы наше веб-приложение приносило пользу, введенные пользователями данные требуется где-то хранить, а также иметь возможность динамически запрашивать данные хранилища при запросе этих самых данных.

Существует много разных хранилищ данных, которые можно использовать для нашего приложения — каждое со своими достоинствами и недостатками. Мы используем популярную реляционную базу данных **MySQL**.

## Установка MySQL для Golang

Для дальнейшей работы нам требуется **установить MySQL на компьютер**. Официальная документация MySQL содержит [подробную инструкцию](https://dev.mysql.com/doc/refman/5.7/en/installing.html) по установке для всех типов операционных систем. Если вы используете Mac OS, то **установить MySQL** можно следующим образом:

При установке MySQL вас могут попросить указать пароль для пользователя `root`. Не забудьте об этом, так как пароль будет важен для следующего шага.

## Создание базы данных в MySQL

После **успешной установки MySQL** к ней можно подключиться через [терминал](https://golangs.org/managing-configuration-settings "запуск приложения в терминале") используя данные от `root` пользователя. Используемая для этого команда зависит от установленной версии MySQL. С **MySQL 5.7** можно попробовать подключиться к MySQK через следующую команду:

Дальше требуется **создать новую базу данных в MySQL** для хранения всех данных от нашего сайта. Скопируйте и вставьте следующие команды в терминале от mysql, чтобы создать новую базу данных `snippetbox` с использованием [UTF-8 кодировки](https://golangs.org/kodirovka-zamena "меняем кодировку в golang").

```sql
-- Создание новой UTF-8 базы данных `snippetbox`
CREATE DATABASE snippetbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Переключение на использование базы данных `snippetbox`
USE snippetbox;
```
Затем скопируйте ниже предоставленный SQL-запрос для создания новой таблицы `snippets`, в которой будут храниться текстовые заметки нашего приложения.

```sql
-- Создание таблицы `snippets`
CREATE TABLE snippets (
    id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created DATETIME NOT NULL,
    expires DATETIME NOT NULL
);

-- Добавление индекса для созданного столбца
CREATE INDEX idx_snippets_created ON snippets(created);
```
У каждой записи в таблице будет столбик `id` поле с данными типа [integer](https://golangs.org/integer "integer golang"), которые будут действовать как **уникальный идентификатор** для каждой заметки. В ней также будет короткое текстовое поле `title` (заголовок) и поле `content` для хранения текста самой заметки. Метаданные о времени создания заметки будут храниться в поле `created`, а срок жизни заметки будет храниться в поле `expires`.

Давайте добавим несколько записей в таблицу `snippets` (которую мы будем использовать в следующих нескольких уроках). Мы будем использовать несколько **коротких пословиц** в качестве содержания текстовых заметок, но на самом деле здесь контент не имеет значения.

```sql
-- Добавляем несколько тестовых записей
INSERT INTO snippets (title, content, created, expires) VALUES (
    'Не имей сто рублей',
    'Не имей сто рублей,\nа имей сто друзей.',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'Лучше один раз увидеть',
    'Лучше один раз увидеть,\nчем сто раз услышать.',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 365 DAY)
);

INSERT INTO snippets (title, content, created, expires) VALUES (
    'Не откладывай на завтра',
    'Не откладывай на завтра,\nчто можешь сделать сегодня.',
    UTC_TIMESTAMP(),
    DATE_ADD(UTC_TIMESTAMP(), INTERVAL 7 DAY)
);
```
## Создание нового пользователя в MySQL

С точки зрения безопасности не рекомендуется подключаться к MySQL через `root` пользователя из веб-приложения. Будет лучше **создать пользователя** для базы данных с **ограниченными правами доступа**.

Пока вы все еще подключены к командной строки от MySQL, выполните следующие команды для создания нового `web` пользователя с привилегиями выполнения `SELECT` и `INSERT` запросов только для его базы данных.

```sql
CREATE USER 'web'@'localhost';
GRANT SELECT, INSERT, UPDATE ON snippetbox.* TO 'web'@'localhost';

-- Важно: Не забудьте заменить 'pass' на свой пароль, иначе это и будет паролем.
ALTER USER 'web'@'localhost' IDENTIFIED BY 'pass';
```
После этого выполните в командной строке `exit` для выхода из MySQL.


### Скачать исходный код + MySQL дамп

В архивами с исходным кодом появилась новая папка `SQL` которая содержит дамп (**резервная копия**) нашей базы данных `snippetbox.sql`. Если вдруг, у вас появились какие либо проблемы с выполнением SQL запросов, или вы начали курс с этой статьи, то можете загрузить дамп `snippetbox.sql` в вашу базу данных и всё будет работать.

**Скачать**: [snippetbox-15.zip](https://golangs.org/wp-content/uploads/2021/01/snippetbox-15.zip)
