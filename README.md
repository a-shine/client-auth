# Simple user authentication/authorization service

Simple user authentication service providing functionality to register, login (generate JWT authentication tokens),
refresh (old tokens), return data, logout, delete and suspend users.

This user management service integrates well with the [a-shine/api-gateway](https://github.com/a-shine/api-gateway) and
can be used as a pre-built user management service.

## Getting started

This user management service has two main dependencies:

- MongoDB for storing user data
- Redis as a blacklist of suspended user IDs (enables realtime user account suspension) and user deletion cascade by
  publishing user delete message to user-delete pubsub channel

### Standalone

Easiest way to use locally is with Docker Compose to manage orchestration of dependent services (e.g. MongoDB and Redis)

A pre-built Docker image is available on Docker Hub:
[ashinebourne/user-management](https://hub.docker.com/r/ashinebourne/user-auth)

A sample docker-compose.yml would look something like this:

```yaml
services:
  user-management:
  image: ashinebourne/user-auth:latest
  ports:
    - "8000:8000"
  environment:
    - JWT_SECRET_KEY=secret
    - JWT_TOKEN_EXP_MIN=60
    - DB_HOST=user-db
    - DB_PORT=27017
    - DB_USER=root
    - DB_PASSWORD=secret
    - DB_NAME=user_management
    - REDIS_HOST=user-cache
    - REDIS_PORT=6379
    - REDIS_PASSWORD=password123

  depends_on:
    - user-db
    - user-cache
  user-db:
    image: mongo
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=root
      - MONGO_INITDB_ROOT_PASSWORD=secret
      - MONGO_INITDB_DATABASE=user_management
  user-cache:
    image: redis
    ports:
      - 6379:6379
    command: /bin/sh -c "redis-server --requirepass $$REDIS_PASSWORD"
    environment:
      - REDIS_PASSWORD=password123
```

### Integrating with the [a-shine/api-gateway](https://github.com/a-shine/api-gateway)

Check out the [a-shine/microservice-template](https://github.com/a-shine/microservice-template) for a template project
on how to configure this user management service and the [a-shine/api-gateway](https://github.com/a-shine/api-gateway).

## Using and testing the API

Registering a new user:

```bash
curl -v -H "Content-type: application/json" -d '{"password": "secret", "email":"bob@myemail.com", "first_name":"Bob",
"last_name":"Smith"}' localhost:8000/register
```

Login to get a JWT:

```bash
curl -v -H "Content-type: application/json" -d '{"password": "secret", "email":"bob@myemail.com"}' localhost:8000/login
```

Refresh a JWT:

```bash
curl -v --cookie "token=[TOKEN]" localhost:8000/refresh
```

Return user data:

```bash
curl -v --cookie "token=[TOKEN]" localhost:8000/me
```

## Testing

The test image is built within the docker-compose build process (just to make things a bit easier)

To run the tests:

```bash
docker-compose run user-management-test go test
```
