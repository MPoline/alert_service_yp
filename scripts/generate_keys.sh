#!/bin/bash

echo "Генерация RSA ключей"

mkdir -p keys

openssl genrsa -out keys/private_key.pem 2048
echo "Приватный ключ создан: keys/private_key.pem"

openssl rsa -in keys/private_key.pem -pubout -out keys/public_key.pem
echo "Публичный ключ создан: keys/public_key.pem"

echo "Проверка ключей:"
ls -la keys/
echo ""

echo "Готово! Ключи успешно сгенерированы."