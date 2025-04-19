# Сервис аутентификации, часть с JWT токеном


## Для запуска приложения

```
docker-compose down -v && docker-compose up --build

```

## Доступные адреса

```
http://localhost:8081/auth/tokens 

http://localhost:8081/auth/refresh

http://localhost:8081/auth/logout 

http://localhost:8081/api/user 
```

## Примеры запросов

```
curl -X GET "http://localhost:8081/auth/tokens?user_id=test_user"
```
```
curl -X POST "http://localhost:8081/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user", "refresh_token": "<токен>"}'
```
```
curl -X GET "http://localhost:8081/api/user" \
  -H "Authorization: Bearer <токен>"
```
```
curl -X POST "http://localhost:8081/auth/logout" \
  -H "Authorization: Bearer <токен>"
```

### Также для тестирования изменения ip, можно использовать

 ```
curl -X GET "http://localhost:8081/auth/tokens?user_id=test123"


curl -X POST "http://localhost:8081/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test123", "refresh_token": "токен"}'

curl -X POST "http://localhost:8081/auth/refresh" \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 1.2.3.4" \
  -d '{"user_id": "test123", "refresh_token": "токен"}'
 ```