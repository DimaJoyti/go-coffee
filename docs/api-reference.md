# API Reference

This document provides a reference for the API endpoints exposed by the Coffee Order System.

## Base URL

By default, the API is available at:

```
http://localhost:3000
```

The port can be configured using the `SERVER_PORT` environment variable or the `server.port` configuration option.

## Endpoints

### Place an Order

Places a new coffee order.

**URL**: `/order`

**Method**: `POST`

**Content Type**: `application/json`

**Request Body**:

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `customer_name` | string | The name of the customer | Yes |
| `coffee_type` | string | The type of coffee | Yes |

**Example Request**:

```http
POST /order HTTP/1.1
Host: localhost:3000
Content-Type: application/json

{
  "customer_name": "John Doe",
  "coffee_type": "Latte"
}
```

**Success Response**:

**Code**: `200 OK`

**Content**:

```json
{
  "success": true,
  "msg": "Order for John Doe placed successfully!",
  "order": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "customer_name": "John Doe",
    "coffee_type": "Latte",
    "status": "pending",
    "created_at": "2023-06-01T12:00:00Z",
    "updated_at": "2023-06-01T12:00:00Z"
  }
}
```

**Error Responses**:

**Code**: `400 Bad Request`

**Content**:

```json
{
  "error": "Invalid request body"
}
```

**Code**: `405 Method Not Allowed`

**Content**:

```json
{
  "error": "Invalid request method"
}
```

**Code**: `500 Internal Server Error`

**Content**:

```json
{
  "error": "Error placing order"
}
```

### Get Order

Retrieves a specific order by ID.

**URL**: `/order/{id}`

**Method**: `GET`

**Example Request**:

```http
GET /order/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: localhost:3000
```

**Success Response**:

**Code**: `200 OK`

**Content**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_name": "John Doe",
  "coffee_type": "Latte",
  "status": "pending",
  "created_at": "2023-06-01T12:00:00Z",
  "updated_at": "2023-06-01T12:00:00Z"
}
```

**Error Responses**:

**Code**: `404 Not Found`

**Content**:

```json
{
  "error": "Order not found"
}
```

### Cancel Order

Cancels a specific order by ID.

**URL**: `/order/{id}/cancel`

**Method**: `POST`

**Example Request**:

```http
POST /order/550e8400-e29b-41d4-a716-446655440000/cancel HTTP/1.1
Host: localhost:3000
```

**Success Response**:

**Code**: `200 OK`

**Content**:

```json
{
  "success": true,
  "msg": "Order cancelled successfully!",
  "order": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "customer_name": "John Doe",
    "coffee_type": "Latte",
    "status": "cancelled",
    "created_at": "2023-06-01T12:00:00Z",
    "updated_at": "2023-06-01T12:05:00Z"
  }
}
```

**Error Responses**:

**Code**: `404 Not Found`

**Content**:

```json
{
  "error": "Order not found"
}
```

**Code**: `400 Bad Request`

**Content**:

```json
{
  "error": "Order cannot be cancelled"
}
```

### List Orders

Lists all orders, with optional filtering.

**URL**: `/orders`

**Method**: `GET`

**Query Parameters**:

| Parameter | Description |
|-----------|-------------|
| `status` | Filter orders by status (pending, processing, completed, cancelled) |
| `customer` | Filter orders by customer name |

**Example Request**:

```http
GET /orders?status=pending HTTP/1.1
Host: localhost:3000
```

**Success Response**:

**Code**: `200 OK`

**Content**:

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "customer_name": "John Doe",
    "coffee_type": "Latte",
    "status": "pending",
    "created_at": "2023-06-01T12:00:00Z",
    "updated_at": "2023-06-01T12:00:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "customer_name": "Jane Smith",
    "coffee_type": "Espresso",
    "status": "pending",
    "created_at": "2023-06-01T12:05:00Z",
    "updated_at": "2023-06-01T12:05:00Z"
  }
]
```

### Health Check

Checks the health of the API.

**URL**: `/health`

**Method**: `GET`

**Example Request**:

```http
GET /health HTTP/1.1
Host: localhost:3000
```

**Success Response**:

**Code**: `200 OK`

**Content**:

```json
{
  "status": "ok"
}
```

## Headers

The API includes the following headers in responses:

| Header | Description |
|--------|-------------|
| `Content-Type` | The content type of the response, always `application/json` |
| `X-Request-ID` | A unique ID for the request, useful for tracing |
| `Access-Control-Allow-Origin` | CORS header, set to `*` |
| `Access-Control-Allow-Methods` | CORS header, set to `GET, POST, PUT, DELETE, OPTIONS` |
| `Access-Control-Allow-Headers` | CORS header, set to `Content-Type, Authorization` |

## Error Handling

The API returns appropriate HTTP status codes and error messages in the response body. Error responses have the following format:

```json
{
  "error": "Error message"
}
```

## Rate Limiting

The API does not currently implement rate limiting.

## Authentication

The API does not currently implement authentication.

## Next Steps

- [Kafka Integration](kafka-integration.md): Learn about Kafka integration.
- [Development Guide](development-guide.md): Learn how to develop the API.
- [Testing](testing.md): Learn how to test the API.
