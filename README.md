# Go Stripe

A Go web application for processing payments with Stripe.

## Architecture

```
├── cmd/
│   ├── api/          # Backend API server (port 4001)
│   └── web/          # Frontend web server (port 4000)
├── internal/
│   └── cards/        # Stripe payment processing logic
```

- **Web Server** — Serves HTML templates with a virtual terminal for card payments
- **API Server** — RESTful API handling payment intents and Stripe integration
- **Cards Package** — Internal library wrapping Stripe SDK for payment processing

## Requirements

- Go 1.21+
- Stripe account (test or live keys)

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/dunky-star/go-stripe.git
   cd go-stripe
   ```

2. Create a `.env` file in the project root:
   ```env
   STRIPE_KEY=pk_test_your_publishable_key
   STRIPE_SECRET=sk_test_your_secret_key
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

## Running

Start both servers in separate terminals:

```bash
# Terminal 1 - Web server (port 4000)
go run ./cmd/web

# Terminal 2 - API server (port 4001)
go run ./cmd/api
```

Then open http://localhost:4000/v1/virtual-card in your browser.

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | 4000 (web), 4001 (api) | Server port |
| `-env` | dev | Environment (dev/qa/prod) |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payment-intent` | Create a Stripe PaymentIntent |

## License

MIT
