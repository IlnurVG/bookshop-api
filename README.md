# Bookshop API

RESTful API service for a bookstore, implemented in Go 1.24.

## Project Description

Bookshop API provides a backend for a bookstore with the following functionality:

- User management (registration, authentication, profile management)
- Book catalog with categories and search
- Shopping cart system with item reservations
- Order processing and management
- Payment integration
- Admin panel for content management

## Technologies

- Go 1.24
- Echo framework
- PostgreSQL
- Redis
- Docker & Docker Compose
- JWT authentication
- OpenAPI/Swagger documentation

## Project Structure

```
bookshop-api/
├── cmd/
│   └── api/                - Application entry point
├── config/                 - Configuration files
├── internal/
│   ├── domain/             - Domain models and repository interfaces
│   ├── handlers/           - HTTP handlers for API endpoints
│   ├── middleware/         - HTTP middleware components
│   ├── repository/         - Repository implementations
│   ├── server/             - Server configuration and setup
│   └── service/            - Service layer (business logic)
├── migrations/             - Database migrations
├── pkg/                    - Reusable packages
│   ├── errors/             - Error handling utilities
│   ├── logger/             - Logging package
│   └── validator/          - Request validation
└── docs/                   - Documentation
```

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- Make

### Running Locally

1. Clone the repository:
   ```
   git clone https://github.com/username/bookshop-api.git
   cd bookshop-api
   ```

2. Start dependencies (PostgreSQL and Redis):
   ```
   make docker-up
   ```

3. Run migrations:
   ```
   make migrate-up
   ```

4. Start the server:
   ```
   make run
   ```

The API will be available at `http://localhost:8080`.

## API Documentation

When the server is running, Swagger documentation is available at:
`http://localhost:8080/swagger`

## Testing

Run tests:
```
make test
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
