# Production-Ready TOTP Backend

A secure, RFC 6238 compliant TOTP authentication system in Go.

## Features
- **Security**: AES-256-GCM encryption, Envelope Encryption, Constant-time verification.
- **Persistence**: SQLite database for storing user secrets and recovery codes.
- **Routing**: `gorilla/mux` for robust HTTP routing.
- **Compatibility**: Works with Google Authenticator, Authy, etc.
- **Protection**: Rate limiting to prevent brute-force attacks.
- **Recovery**: One-time use backup codes.

## Requirements
- Go 1.20+
- CGO enabled (for SQLite support)

## Getting Started

### 1. Run the API Server
Start the backend service. It will create `totp.db` automatically.
```bash
go run cmd/api/main.go
```
The server listens on `localhost:8080`.

### 2. Run the Interactive Demo
Open a new terminal to run the client demo.
```bash
go run cmd/demo/main.go
```
Follow the on-screen instructions to:
1. **Enroll**: Enter a username. The demo will print a QR code in your terminal.
2. **Scan**: Use Google Authenticator app to scan the QR code.
3. **Verify**: Enter the 6-digit code from the app to enable 2FA.
4. **Test**: Validate subsequent codes or test recovery codes.

## API Endpoints

- **POST /enroll**: `{ "user_id": "string" }` -> Returns Secret & QR URL.
- **POST /verify**: `{ "user_id": "string", "code": "string" }` -> Enables TOTP.
- **POST /validate**: `{ "user_id": "string", "code": "string" }` -> Checks code.
- **POST /recover**: `{ "user_id": "string", "code": "string" }` -> Uses recovery code.

## Architecture
- `cmd/`: Entrypoints (API, Demo).
- `internal/auth/`: Core logic (TOTP, Enrollment, Recovery, RateLimit).
- `internal/crypto/`: Encryption services.
- `internal/storage/`: Database persistence (SQLite).
- `internal/http/`: API Handlers & Routing.
