# Webhook Server

This is a simple webhook server written in Go that receives JSON payloads and sends the data to a Telegram chat. It provides functionality to handle different paths and filter keys to customize the processing of JSON data.

## Features

- Receives webhook requests and processes JSON payloads.
- Sends the processed data to a Telegram chat.
- Supports dynamic paths with customizable filter keys.
- Provides a heartcheck endpoint for health monitoring.

## Requirements

- Go 1.14 or higher
- Telegram Bot API token
- Telegram chat ID

## Configuration

Before running the server, make sure to set the following environment variables:

- `TG_API_TOKEN`: Your Telegram Bot API token.
- `TG_CHAT_ID`: The chat ID of the Telegram chat where you want to send the messages.
- `URL_PATH`: A comma-separated list of paths to be handled by the server. For each path, a corresponding environment variable needs to be set with the filter keys for that path. The environment variable should be in uppercase, prefixed with the path and suffixed with "_FILTER_KEY". For example, if your path is "data" then the filter keys should be set in the environment variable `DATA_FILTER_KEY`.

## Usage

### Running with Golang
```shell
git clone https://github.com/sppo555/webhook.git
cd webhook
go run main.go
```

### Running with Docker Compose
```yaml
version: '3.5'
services:
  webhook-tg:
    image: sppo55/webhook-tg:latest
    hostname: webhook-tg
    container_name: webhook-tg
    restart: always
    ports:
      - "8080:8080"
    environment:
      TG_CHAT_ID: "<Your_TG_CHAT_ID>"
      TG_API_TOKEN: "<Your_TG_API_TOKEN>"
      URL_PATH: "key1,key2,key3"
      KEY1_FILTER_KEY: "value1"
      KEY2_FILTER_KEY: "value2"
      KEY3_FILTER_KEY: "value3"
```
```shell
docker-compose up -d
```

The server listens on port 8080 by default. You can change the port by modifying the `Port` constant in the code.

### Webhook Endpoint

To receive webhook requests, send a POST request to the `/webhook` path. The server expects a JSON payload in the request body. The payload will be processed and sent to the Telegram chat.

Example webhook request:

```shell
curl -X POST -H "Content-Type: application/json" -d '{"key1": "value1", "key2": "value2"}' http://localhost:8080/webhook
```

## Heartcheck Endpoint

The server provides a /heartcheck endpoint for health monitoring. Send a GET request to this path to check if the server is running.


Example heartcheck request:
```shell
curl http://localhost:8080/heartcheck
```

## Dynamic Paths
The server supports dynamic paths specified in the URL_PATH environment variable. For each dynamic path, you need to set a corresponding environment variable with the filter keys for that path. The filter keys should be set in uppercase, prefixed with the path and suffixed with "_FILTER_KEY".

For example, if your URL_PATH is set to "data" and you want to filter keys "key1" and "key2" for that path, you need to set the DATA_FILTER_KEY environment variable with the value "key1,key2".

Then you can send a POST request to the /data path to trigger the dynamic processing.

Example dynamic path request:
```shell
curl -X POST -H "Content-Type: application/


" -d '{"key1": "value1", "key2": "value2", "key3": "value3"}' http://localhost:8080/data
```

## Example

### Docker-compose

```shell
version: '3.5'
services:
  webhook-tg:
    image: sppo55/webhook-tg:latest
    hostname: webhook-tg
    container_name: webhook-tg
    restart: always
    ports:
      - "8080:8080"
    environment:
      TG_CHAT_ID: "<Your_TG_CHAT_ID>"
      TG_API_TOKEN: "<Your_TG_API_TOKEN>"
      URL_PATH: "users,items,orders"
      USERS_FILTER_KEY: "customer_name"
      ITEMS_FILTER_KEY: "items"
      ORDERS_FILTER_KEY: "order_id,total,shipping_address"
```

## Test
```shell
curl -X POST -H "Content-Type: application/json" -d '{
  "order_id": "12345",
  "customer_name": "John Doe",
  "total": 27.97,
  "shipping_address": {
    "street": "123 Main St",
    "country": "Country"
  },
  "items": [
    {
      "product_id": "P001",
      "product_name": "Product 1",
      "unit_price": 10.99
    },
    {
      "product_id": "P002",
      "product_name": "Product 2",
      "unit_price": 5.99
    }
  ]
}' http://localhost:8080/webhook

```

### Telegram 
Here's the content that will be forwarded to Telegram based on the requests to /webhook, /users, /items, and /orders, assuming you have correctly set up the Telegram Bot API token (TG_API_TOKEN) and chat ID (TG_CHAT_ID):


For the /webhook request, the following content will be forwarded to Telegram:
```shell
Message Content:
Order ID: 12345
Customer Name: John Doe
Total: 27.97
Shipping Address:
  Street: 123 Main St
  Country: Country
Items:
  - Product ID: P001
    Product Name: Product 1
    Unit Price: 10.99
  - Product ID: P002
    Product Name: Product 2
    Unit Price: 5.99

```
For the /users request, only the customer_name field will be filtered and forwarded to Telegram:
```shell
Message Content:
Customer Name: John Doe
```
For the /items request, only the items field will be filtered and forwarded to Telegram:
```shell
Message Content:
Items:
  - Product ID: P001
    Product Name: Product 1
    Unit Price: 10.99
  - Product ID: P002
    Product Name: Product 2
    Unit Price: 5.99

```
For the /orders request, only the order_id, total, and shipping_address fields will be filtered and forwarded to Telegram:
```shell
Message Content:
Order ID: 12345
Total: 27.97
Shipping Address:
  Street: 123 Main St
  Country: Country

```
