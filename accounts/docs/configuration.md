# Configuration

The Accounts Service can be configured using environment variables or a configuration file.

## Configuration File

The service looks for a configuration file at `config.json` in the current directory. You can specify a different location using the `CONFIG_FILE` environment variable.

Example configuration file:

```json
{
  "server": {
    "port": 4000
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "dbname": "coffee_accounts",
    "sslmode": "disable"
  },
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "account_events",
    "retry_max": 5,
    "required_acks": "all"
  }
}
```

## Environment Variables

You can also configure the service using environment variables. Environment variables take precedence over the configuration file.

### Server Configuration

| Variable | Description | Default |
| --- | --- | --- |
| `SERVER_PORT` | The port to listen on | `4000` |

### Database Configuration

| Variable | Description | Default |
| --- | --- | --- |
| `DB_HOST` | The database host | `localhost` |
| `DB_PORT` | The database port | `5432` |
| `DB_USER` | The database user | `postgres` |
| `DB_PASSWORD` | The database password | `postgres` |
| `DB_NAME` | The database name | `coffee_accounts` |
| `DB_SSLMODE` | The database SSL mode | `disable` |

### Kafka Configuration

| Variable | Description | Default |
| --- | --- | --- |
| `KAFKA_BROKERS` | The Kafka brokers (comma-separated or JSON array) | `localhost:9092` |
| `KAFKA_TOPIC` | The Kafka topic | `account_events` |
| `KAFKA_RETRY_MAX` | The maximum number of retries | `5` |
| `KAFKA_REQUIRED_ACKS` | The required acknowledgements (`no`, `local`, or `all`) | `all` |

## Docker Environment

When running in Docker, you can configure the service using environment variables in the `docker-compose.yml` file:

```yaml
accounts-service:
  build:
    context: ./accounts-service
    dockerfile: Dockerfile
  container_name: accounts-service
  depends_on:
    postgres:
      condition: service_healthy
    kafka-setup:
      condition: service_completed_successfully
  ports:
    - "4000:4000"
  environment:
    SERVER_PORT: 4000
    DB_HOST: postgres
    DB_PORT: 5432
    DB_USER: postgres
    DB_PASSWORD: postgres
    DB_NAME: coffee_accounts
    DB_SSLMODE: disable
    KAFKA_BROKERS: '["kafka:9092"]'
    KAFKA_TOPIC: account_events
    KAFKA_RETRY_MAX: 5
    KAFKA_REQUIRED_ACKS: all
```

## Kubernetes Environment

When running in Kubernetes, you can configure the service using environment variables in the deployment manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: accounts-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: accounts-service
  template:
    metadata:
      labels:
        app: accounts-service
    spec:
      containers:
      - name: accounts-service
        image: accounts-service:latest
        ports:
        - containerPort: 4000
        env:
        - name: SERVER_PORT
          value: "4000"
        - name: DB_HOST
          value: postgres
        - name: DB_PORT
          value: "5432"
        - name: DB_USER
          value: postgres
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-credentials
              key: password
        - name: DB_NAME
          value: coffee_accounts
        - name: DB_SSLMODE
          value: disable
        - name: KAFKA_BROKERS
          value: '["kafka:9092"]'
        - name: KAFKA_TOPIC
          value: account_events
        - name: KAFKA_RETRY_MAX
          value: "5"
        - name: KAFKA_REQUIRED_ACKS
          value: all
```
