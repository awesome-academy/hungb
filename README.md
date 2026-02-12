# SUN Booking Tours

A tour booking web application with separate public and admin interfaces, built with Go and server-side rendering.

## Tech Stack

- **Backend**: Go 1.22+, Gin Web Framework
- **Database**: PostgreSQL 16, GORM
- **Auth**: bcrypt, OAuth2 (goth)
- **Templating**: Go html/template (SSR)
- **UI**: Bootstrap 5
- **Container**: Docker, Docker Compose

## Features

### Public Site
- Browse tours and categories
- User registration and authentication (email + OAuth2)
- Tour booking with schedule selection
- User profile and bank account management
- Tour ratings and reviews with comments

### Admin Site
- User management
- Tour and category management
- Booking and payment tracking
- Review moderation

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.22+ (for local development)

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd sun-booking-tours
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Start services:
```bash
make up
```

4. Run migrations:
```bash
make migrate-up
```

5. Access the application:
- Public site: http://localhost:8080
- Admin site: http://localhost:8080/admin

## Development

```bash
# Start containers
make up

# View logs
make logs

# Run migrations
make migrate-up

# Stop containers
make down

# Rebuild and restart
make restart
```

## Project Structure

```
├── cmd/app/                # Application entry point
├── internal/
│   ├── config/            # Configuration
│   ├── models/            # GORM models
│   ├── repository/        # Data access layer
│   ├── services/          # Business logic
│   ├── handlers/          # HTTP handlers (admin & public)
│   └── middleware/        # Auth, CSRF, etc.
├── templates/             # Go html templates
├── static/                # CSS, JS, images
├── migrations/            # Database migrations
└── specs/                 # Task specifications
```

## Architecture

The application follows a layered architecture:

```
Handler → Service → Repository → GORM Model
```

- **Handler**: Parse requests, render templates
- **Service**: Business logic and validation
- **Repository**: Database queries
- **Model**: GORM entity definitions

## License

MIT
