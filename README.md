# Simple client authentication service

Simple client authentication service providing functionality to register, login
(generate JWT authentication tokens), refresh (old tokens), return data, logout,
delete and suspend clients.

Clients are generalised to be any external entity interacting with an
application that requires authentication (e.g. users, IoT devices, other
services etc.). A client at a minimum has an email address and password, but the
schema can be extended to include other fields such as first name, last name,
phone number of a user.

This user management service integrates well with the
[a-shine/api-gateway](https://github.com/a-shine/api-gateway) and can be used
as a pre-built client authentication service.

## Getting started

This user management service has two main dependencies:

- [MongoDB](https://www.mongodb.com/) for storing client data
- [Redis](https://redis.io/)
  - As a blacklist of suspended user IDs to enable real-time client
    suspension
  - To enable a client data deletion cascade (to other services storing client
    data) by publishing a `client-data-deletion-request` to a pubsub channel
    (this feature is dependent on you designing your services to listen to this
    channel)

The easiest way to use the service locally is with Docker Compose to manage
orchestration of dependent services (MongoDB and Redis).

A pre-built Docker image is available on Docker Hub:
[ashinebourne/user-management](https://hub.docker.com/r/ashinebourne/user-auth)

A sample docker-compose.yml would look something like this:

```yaml
services:
  client-auth:
  image: ashinebourne/user-auth:latest
  ports:
    - "8000:8000"
  environment:
    - JWT_SECRET_KEY=secret
    - JWT_TOKEN_EXP_MIN=60
    - DB_HOST=client-db
    - DB_PORT=27017
    - DB_USER=root
    - DB_PASSWORD=secret
    - DB_NAME=user_management
    - REDIS_HOST=client-auth-cache
    - REDIS_PORT=6379
    - REDIS_PASSWORD=password123
  depends_on:
    - user-db
    - user-cache
  client-db:
    image: mongo
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=root
      - MONGO_INITDB_ROOT_PASSWORD=secret
      - MONGO_INITDB_DATABASE=user_management
  client-auth-cache:
    image: redis
    ports:
      - 6379:6379
    command: /bin/sh -c "redis-server --requirepass $$REDIS_PASSWORD"
    environment:
      - REDIS_PASSWORD=password123
```

### Integrating with the [a-shine/api-gateway](https://github.com/a-shine/api-gateway)

Check out the
[a-shine/microservice-template](https://github.com/a-shine/microservice-template)
for an example on how to configure this client service with the
[a-shine/api-gateway](https://github.com/a-shine/api-gateway).

## Using and testing the API

Registering a new user:

```bash
curl -v \
    -H "Content-type: application/json" \
    -d '{"password": "secret", "email":"bob@myemail.com", "firstName":"Bob", "lastName":"Smith"}' \
    localhost:8000/register
```

Login to get a JWT (returned with a set-cookie header):

```bash
curl -v \
     -H "Content-type: application/json" \
     -d '{"password": "secret", "email":"bob@myemail.com"}'\
     localhost:8000/login
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

The test image is built within the `docker-compose.yaml` build process.

To run the tests locally:

```bash
docker-compose run user-management-test go test
```
