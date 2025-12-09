# Настройка SSH для удобного подключения

## Зачем нужно?

Настройка SSH ключа позволит:
- Подключаться к серверу без ввода пароля
- Безопаснее защитить доступ к серверу
- Автоматизировать деплой с помощью скриптов

## Шаг 1: Проверка существующего SSH ключа

На вашем локальном компьютере:

```bash
ls -la ~/.ssh/
```

Если видите файлы `id_rsa` и `id_rsa.pub` (или `id_ed25519` и `id_ed25519.pub`), ключ уже есть.

## Шаг 2: Генерация SSH ключа (если нужно)

Если ключа нет, создайте новый:

```bash
# Рекомендуемый современный алгоритм
ssh-keygen -t ed25519 -C "ваш_email@example.com"

# Или старый, но совместимый
ssh-keygen -t rsa -b 4096 -C "ваш_email@example.com"
```

При появлении запросов:
- "Enter file in which to save the key" - нажмите Enter (использовать путь по умолчанию)
- "Enter passphrase" - можете указать пароль или оставить пустым (Enter)

## Шаг 3: Копирование ключа на сервер

### Вариант A: Автоматически (рекомендуется)

```bash
ssh-copy-id cashback@82.26.150.98
```

Введите пароль пользователя `cashback` когда попросят.

### Вариант B: Вручную

Скопируйте содержимое публичного ключа:

```bash
cat ~/.ssh/id_ed25519.pub
# или
cat ~/.ssh/id_rsa.pub
```

Подключитесь к серверу:

```bash
ssh cashback@82.26.150.98
```

На сервере выполните:

```bash
# Создание директории для SSH ключей (если её нет)
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Добавление ключа
nano ~/.ssh/authorized_keys
# Вставьте скопированный ключ
# Сохраните (Ctrl+O, Enter, Ctrl+X)

# Установка правильных прав
chmod 600 ~/.ssh/authorized_keys
```

## Шаг 4: Проверка подключения

На локальном компьютере:

```bash
ssh cashback@82.26.150.98
```

Должны подключиться **без запроса пароля**!

## Шаг 5: Настройка SSH конфига (опционально)

Для ещё большего удобства создайте конфиг:

```bash
nano ~/.ssh/config
```

Добавьте:

```
Host cashback-server
    HostName 82.26.150.98
    User cashback
    IdentityFile ~/.ssh/id_ed25519
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

Теперь можно подключаться просто:

```bash
ssh cashback-server
```

## Шаг 6: Отключение парольной аутентификации (для безопасности)

⚠️ **ВАЖНО**: Делайте это ТОЛЬКО после того, как убедитесь, что SSH ключ работает!

На сервере:

```bash
sudo nano /etc/ssh/sshd_config
```

Найдите и измените/добавьте строки:

```
PasswordAuthentication no
PubkeyAuthentication yes
ChallengeResponseAuthentication no
```

Перезапустите SSH:

```bash
sudo systemctl restart sshd
# или
sudo service ssh restart
```

## Проверка работы автоматического деплоя

После настройки SSH ключа можно использовать автоматический деплой:

```bash
export TELEGRAM_BOT_TOKEN="ваш_токен"
./remote-deploy.sh
```

Скрипт подключится к серверу автоматически без запроса пароля!

## Troubleshooting

### SSH ключ не работает

1. Проверьте права на файлы:
```bash
# На локальной машине
chmod 600 ~/.ssh/id_ed25519
chmod 644 ~/.ssh/id_ed25519.pub

# На сервере
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

2. Проверьте логи подключения:
```bash
ssh -v cashback@82.26.150.98
```

### Доступ заблокирован после отключения пароля

Если вы отключили парольную аутентификацию, но SSH ключ не работает:

1. Подключитесь через консоль сервера (VNC, KVM)
2. Включите обратно парольную аутентификацию:
```bash
sudo nano /etc/ssh/sshd_config
# Установите: PasswordAuthentication yes
sudo systemctl restart sshd
```
3. Исправьте настройку SSH ключей
4. Повторите попытку

### Permission denied (publickey)

Проблема: сервер не принимает ваш ключ.

Решение:
```bash
# Проверьте, что ключ добавлен в ssh-agent
ssh-add -l

# Если пусто, добавьте ключ
ssh-add ~/.ssh/id_ed25519
```

## Дополнительная безопасность

### Изменение порта SSH

В файле `/etc/ssh/sshd_config` на сервере:

```
Port 2222  # Вместо стандартного 22
```

Подключение:
```bash
ssh -p 2222 cashback@82.26.150.98
```

### Настройка fail2ban

Защита от брутфорс атак:

```bash
sudo apt-get install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### Использование только конкретных ключей

В `/etc/ssh/sshd_config`:

```
AuthorizedKeysFile .ssh/authorized_keys
PermitRootLogin no
MaxAuthTries 3
```

## Полезные команды

```bash
# Просмотр активных SSH сессий
who

# Просмотр истории входов
last

# Проверка статуса SSH
sudo systemctl status sshd

# Тест конфигурации SSH
sudo sshd -t
```

## Резервная копия ключей

⚠️ **ВАЖНО**: Сохраните резервную копию приватного ключа в безопасном месте!

```bash
# Экспорт ключей
cp ~/.ssh/id_ed25519 /путь/к/безопасному/месту/
cp ~/.ssh/id_ed25519.pub /путь/к/безопасному/месту/

# Или создайте архив с паролем
tar -czf ssh-backup.tar.gz ~/.ssh/
gpg -c ssh-backup.tar.gz  # Зашифровать
```

Если потеряете приватный ключ и отключили парольную аутентификацию - доступ к серверу будет заблокирован!

