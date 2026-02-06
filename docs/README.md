# Peroot - WhatsApp Order Notification Connector

A lightweight Go application that acts as a messaging bridge between your online shop and WhatsApp Business. When an order is placed, your shop calls this app's API, which then sends a WhatsApp notification to your business phone number using the **WhatsApp Cloud API (Meta)**.

## Architecture

```
┌──────────────┐     POST /api/v1/orders/notify     ┌──────────┐     WhatsApp Cloud API     ┌──────────────┐
│  Online Shop │ ──────────────────────────────────> │  Peroot  │ ──────────────────────────> │  WhatsApp    │
│  (Your App)  │ <────────────────────────────────── │  Server  │ <──────────────────────────  │  Business    │
└──────────────┘          JSON Response              └──────────┘        Message Sent         └──────────────┘
```

**Tech Stack:**
- **Go** with **Gin** web framework
- **WhatsApp Cloud API** (Meta) - free tier: 1,000 service conversations/month
- **No paid services** (no Twilio, no MessageBird, etc.)

## Project Structure

```
peroot/
├── main.go                 # Entry point, router setup
├── config/
│   └── config.go           # Configuration loader & validation
├── handler/
│   └── order.go            # HTTP request handlers
├── service/
│   └── whatsapp.go         # WhatsApp Cloud API client with retry
├── model/
│   ├── order.go            # Order request/response models
│   └── whatsapp.go         # WhatsApp API models
├── middleware/
│   └── auth.go             # API key authentication
├── .env.example            # Environment variable template
└── docs/
    └── README.md           # This file
```

## Prerequisites

- Go 1.21 or later
- A Meta (Facebook) Developer account
- A WhatsApp Business Account

## WhatsApp Cloud API Setup

### Step 1: Create a Meta Developer App

1. Go to [developers.facebook.com](https://developers.facebook.com) and log in.
2. Click **Create App** > select **Business** type > give it a name (e.g., "Peroot Notifications").
3. In the app dashboard, click **Add Product** > find **WhatsApp** > click **Set Up**.

### Step 2: Get Your API Credentials

In the WhatsApp **API Setup** page, you'll find:

- **Phone Number ID** - a numeric ID (not the phone number itself). This goes into `WHATSAPP_PHONE_NUMBER_ID`.
- **Temporary Access Token** - valid for 24 hours (fine for testing). This goes into `WHATSAPP_ACCESS_TOKEN`.

### Step 3: Add a Test Recipient

1. In the API Setup page, under **To**, click **Manage phone number list**.
2. Add your business phone number and verify it with the SMS code.

### Step 4: Create a Message Template

1. Go to [business.facebook.com](https://business.facebook.com) > **WhatsApp Manager** > **Message Templates**.
2. Click **Create Template**.
3. Configure:
   - **Category:** Utility
   - **Name:** `order_notification`
   - **Language:** English (or your preferred language)
   - **Body:** `New order *{{1}}* from *{{2}}*. Total: *{{3}}*. Items: {{4}}`

   The placeholders map to:
   | Placeholder | Value |
   |---|---|
   | `{{1}}` | Order ID |
   | `{{2}}` | Customer Name |
   | `{{3}}` | Total Amount + Currency |
   | `{{4}}` | Item Summary |

4. Submit for approval. Utility templates are typically approved within minutes.

### Step 5: Generate a Permanent Access Token (Production)

The temporary token expires in 24 hours. For production:

1. Go to [business.facebook.com](https://business.facebook.com) > **Business Settings** > **System Users**.
2. Create a new system user (Admin role).
3. Click **Generate New Token**, select your app, and grant the `whatsapp_business_messaging` permission.
4. Copy the generated token — this does not expire.

### Step 6: Register a Real Business Number (Production)

1. In the WhatsApp API Setup page, click **Add Phone Number**.
2. Follow the verification process for your real business number.
3. Update `WHATSAPP_PHONE_NUMBER_ID` with the new number's ID.

## Installation & Running

### 1. Clone and configure

```bash
git clone https://github.com/achmadnp/peroot.git
cd peroot
cp .env.example .env
```

Edit `.env` with your actual values:

```env
APP_PORT=:8080
APP_ENV=development
API_KEY=your-secret-api-key-here

WHATSAPP_API_URL=https://graph.facebook.com/v21.0
WHATSAPP_PHONE_NUMBER_ID=123456789012345
WHATSAPP_ACCESS_TOKEN=EAAxxxxxxxxxxxxxx
WHATSAPP_TEMPLATE_NAME=order_notification
WHATSAPP_TEMPLATE_LANGUAGE=en
WHATSAPP_RECIPIENT_NUMBER=6281234567890
```

### 2. Install dependencies and run

```bash
go mod tidy
go run .
```

The server starts on the configured port (default `:8080`).

## API Reference

### Health Check

```
GET /health
```

No authentication required.

**Response:**
```json
{
  "status": "ok",
  "message": "peroot is running",
  "time": "2026-02-06T17:30:00+07:00"
}
```

### Send Order Notification

```
POST /api/v1/orders/notify
```

**Headers:**
| Header | Required | Description |
|---|---|---|
| `Content-Type` | Yes | Must be `application/json` |
| `X-API-Key` | Yes* | Your API key from `.env`. *Skipped if `API_KEY` is not set. |

**Request Body:**
```json
{
  "order_id": "ORD-20240115-001",
  "customer_name": "Budi Santoso",
  "total_amount": 350000,
  "currency": "IDR",
  "items": [
    {
      "name": "Kaos Polos Hitam",
      "quantity": 2,
      "price": 75000
    },
    {
      "name": "Celana Jeans",
      "quantity": 1,
      "price": 200000
    }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `order_id` | string | Yes | Unique order identifier |
| `customer_name` | string | Yes | Customer's name |
| `total_amount` | number | Yes | Order total (must be > 0) |
| `currency` | string | Yes | Currency code (e.g., "IDR", "USD") |
| `items` | array | Yes | At least one item |
| `items[].name` | string | Yes | Item name |
| `items[].quantity` | integer | Yes | Quantity (must be > 0) |
| `items[].price` | number | Yes | Price per unit (must be > 0) |

**Success Response (200):**
```json
{
  "success": true,
  "message": "Order notification sent successfully via WhatsApp",
  "order_id": "ORD-20240115-001",
  "message_id": "wamid.HBgLMTIzNDU2Nzg5MBUCAA==",
  "timestamp": "2026-02-06T17:30:00Z"
}
```

**Validation Error (400):**
```json
{
  "success": false,
  "error": "Invalid request body: Key: 'OrderRequest.OrderID' Error:Field validation for 'OrderID' failed on the 'required' tag",
  "code": "VALIDATION_ERROR"
}
```

**Authentication Error (401):**
```json
{
  "success": false,
  "error": "Missing X-API-Key header",
  "code": "UNAUTHORIZED"
}
```

**Forbidden (403):**
```json
{
  "success": false,
  "error": "Invalid API key",
  "code": "FORBIDDEN"
}
```

**WhatsApp Send Failure (500):**
```json
{
  "success": false,
  "error": "Failed to send WhatsApp notification",
  "code": "WHATSAPP_SEND_FAILED"
}
```

## Integration Example

### From your online shop (cURL)

```bash
curl -X POST http://localhost:8080/api/v1/orders/notify \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key" \
  -d '{
    "order_id": "ORD-001",
    "customer_name": "Budi Santoso",
    "total_amount": 350000,
    "currency": "IDR",
    "items": [
      {"name": "Kaos Polos Hitam", "quantity": 2, "price": 75000},
      {"name": "Celana Jeans", "quantity": 1, "price": 200000}
    ]
  }'
```

### From PHP (Laravel/WordPress)

```php
$response = Http::withHeaders([
    'X-API-Key' => 'your-secret-api-key',
])->post('http://localhost:8080/api/v1/orders/notify', [
    'order_id' => $order->id,
    'customer_name' => $order->customer_name,
    'total_amount' => $order->total,
    'currency' => 'IDR',
    'items' => $order->items->map(fn($item) => [
        'name' => $item->name,
        'quantity' => $item->quantity,
        'price' => $item->price,
    ])->toArray(),
]);
```

### From JavaScript (Node.js/Next.js)

```javascript
const response = await fetch('http://localhost:8080/api/v1/orders/notify', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-secret-api-key',
  },
  body: JSON.stringify({
    order_id: order.id,
    customer_name: order.customerName,
    total_amount: order.total,
    currency: 'IDR',
    items: order.items.map(item => ({
      name: item.name,
      quantity: item.quantity,
      price: item.price,
    })),
  }),
});
```

### From Python

```python
import requests

response = requests.post(
    "http://localhost:8080/api/v1/orders/notify",
    headers={"X-API-Key": "your-secret-api-key"},
    json={
        "order_id": order.id,
        "customer_name": order.customer_name,
        "total_amount": order.total,
        "currency": "IDR",
        "items": [
            {"name": item.name, "quantity": item.qty, "price": item.price}
            for item in order.items
        ],
    },
)
```

## Features

- **WhatsApp Cloud API (Meta)** - free tier with 1,000 service conversations/month
- **API key authentication** - secure your endpoint with a simple API key
- **Request validation** - validates all required fields with proper error messages
- **Retry with exponential backoff** - automatically retries failed WhatsApp API calls (3 retries: 2s, 4s, 8s)
- **Structured JSON logging** - production-ready logging via `log/slog`
- **Context-aware** - respects request cancellation and timeouts
- **Versioned API** - routes under `/api/v1/` for future compatibility

## Configuration Reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_PORT` | No | `:8080` | Server listen address |
| `APP_ENV` | No | `development` | Environment (`development` or `production`) |
| `API_KEY` | No | _(empty)_ | API key for authentication. If empty, auth is skipped. |
| `WHATSAPP_API_URL` | No | `https://graph.facebook.com/v21.0` | WhatsApp Cloud API base URL |
| `WHATSAPP_PHONE_NUMBER_ID` | **Yes** | - | Your WhatsApp Business phone number ID from Meta |
| `WHATSAPP_ACCESS_TOKEN` | **Yes** | - | Your WhatsApp Cloud API access token |
| `WHATSAPP_TEMPLATE_NAME` | No | `order_notification` | Name of your approved message template |
| `WHATSAPP_TEMPLATE_LANGUAGE` | No | `en` | Template language code |
| `WHATSAPP_RECIPIENT_NUMBER` | **Yes** | - | Business number to receive notifications (international format, no +) |

## Troubleshooting

### "Configuration validation failed"
Ensure `WHATSAPP_PHONE_NUMBER_ID`, `WHATSAPP_ACCESS_TOKEN`, and `WHATSAPP_RECIPIENT_NUMBER` are set in your `.env` file.

### WhatsApp API returns error code 132000
Your template parameters don't match the template. Ensure your `order_notification` template has exactly 4 body parameters (`{{1}}` through `{{4}}`).

### WhatsApp API returns error code 190
Your access token has expired. Generate a new permanent token from Meta Business Suite (see Step 5 above).

### Template not approved
Use the pre-approved `hello_world` template for testing by setting `WHATSAPP_TEMPLATE_NAME=hello_world` in your `.env`. Note: `hello_world` has no parameters, so the notification content will be generic.

## License

MIT
