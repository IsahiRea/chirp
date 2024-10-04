# Chirpy API

Chirpy is a RESTful API for managing user authentication, creating and retrieving "chirps" (messages), and other related actions. The API also includes admin features and webhook handlers.

## Features
- User authentication with JWT.
- Support for refresh tokens.
- CRUD operations for "chirps."
- Metrics tracking for file server hits.
- Admin endpoints for user management.
- Integration with external webhooks (e.g., Polka).

## Requirements
- Go 1.18+
- PostgreSQL
- Environment variables:
  - `DB_URL`: PostgreSQL database connection string.
  - `PLATFORM`: The environment in which the app is running (e.g., `dev`, `prod`).
  - `TOKEN_STRING`: Secret key used for JWT signing.
  - `POLKA_KEY`: API key for handling external webhooks.

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/IsahiRea/chirp.git
    cd chirp
    ```

2. Set up the `.env` file:
    ```bash
    cp .env.example .env
    ```

3. Update the `.env` file with your environment variables:
    ```
    DB_URL=your_postgres_db_url
    PLATFORM=dev
    TOKEN_STRING=your_secret_token_string
    POLKA_KEY=your_polka_key
    ```

4. Install dependencies:
    ```bash
    go mod tidy
    ```

5. Run the server:
    ```bash
    go run main.go
    ```

The API will be accessible at `http://localhost:8080`.

## API Endpoints

### Auth Endpoints

- **Login**
  - `POST /api/login`
  - Request body:
    ```json
    {
      "email": "user@example.com",
      "password": "password"
    }
    ```

- **Refresh Token**
  - `POST /api/refresh`
  - Requires Bearer Token in the header.
  
- **Revoke Token**
  - `POST /api/revoke`
  - Requires Bearer Token in the header.

### User Endpoints

- **Create User**
  - `POST /api/users`
  - Request body:
    ```json
    {
      "email": "newuser@example.com",
      "password": "password"
    }
    ```

- **Update User**
  - `PUT /api/users`
  - Requires Bearer Token in the header.
  - Request body:
    ```json
    {
      "email": "updateduser@example.com",
      "password": "newpassword"
    }
    ```

### Chirps Endpoints

- **Get Chirps**
  - `GET /api/chirps`
  - Query params:
    - `author_id`: UUID of the author.
    - `sort`: Sorting order (`asc` or `desc`).

- **Create Chirp**
  - `POST /api/chirps`
  - Requires Bearer Token in the header.
  - Request body:
    ```json
    {
      "body": "This is a chirp!",
      "user_id": "user_uuid"
    }
    ```

- **Delete Chirp**
  - `DELETE /api/chirps/{chirpID}`
  - Requires Bearer Token in the header.

### Admin Endpoints

- **File Server Hits**
  - `GET /admin/metrics`
  
- **Reset Users**
  - `POST /admin/reset`
  - Only available in `dev` mode.

### Polka Webhooks

- **User Upgraded**
  - `POST /api/polka/webhooks`
  - Requires API key in the header.

### Health Check

- `GET /api/healthz` - Check API health status.

## Database

This project uses PostgreSQL for storing user data and chirps. Make sure to set up the appropriate schema in the database.

## Middleware

- **Metrics Middleware**: Tracks the number of file server hits.

## Error Handling

Standard HTTP status codes are used to indicate errors, with appropriate log messages in case of failures.

