# event-horizon

A template for web apps in Go, based on the Let's Go and Let's Go Further books.

## Requirements

1. [mise](https://mise.jdx.dev) for task running and ENV variables.
2. golang-migrate
3. PostgreSQL

## Configuration

The following environment variables should be set:

```bash
HOST="https://example.com" # Default: :4000
DATABASE_URL="postgres://..."
COOKIE_SECRET_KEY="a-secure-token" # New one generated every run if not set
```
