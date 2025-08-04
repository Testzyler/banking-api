# 🏦 Banking API - Secure Digital Banking Platform

## Table of Contents

- [**System Overview**](#-project-overview)
- [**Key Features & Capabilities**](#-key-features--capabilities)
- [**System Architecture**](#️-system-architecture)
- [**Technology Stack**](#️-technology-stack)
- [**Quick Start Guide**](#-quick-start-guide)
- [**API Documentation**](#-api-documentation)
- [**Security Implementation**](#-security-implementation)
- [**Performance & Testing**](#-performance--testing)
- [**Deployment & DevOps**](#-deployment--devops)
- [**Project Structure**](#-project-structure)
- [**Technical Highlights**](#-technical-highlights)

---

## System Overview

Banking API เป็นระบบ Backend API สำหรับ Banking Application ที่ใช้ Go (Golang) และ Fiber Framework พัฒนาด้วยสถาปัตยกรรม Clean Architecture และ Repository Pattern

## System Architecture Diagram
![architecture](docs/architechture.png)

## Project Structure

```
banking-api/
├── src/                      # Source code
│   ├── app/                  # Application layer
│   │   ├── entities/         # Domain entities
│   │   ├── features/         # Feature modules
│   │   │   ├── auth/         # Authentication system
│   │   │   ├── home/         # Dashboard features
│   │   ├── models/           # Database models
│   │   └── validators/       # Input validation
│   ├── cmd/                  # CLI commands
│   ├── config/               # Configuration
│   ├── database/             # Database layer
│   │   └── migrations/       # Schema migrations
│   ├── logger/               # Logging setup
│   ├── server/               # HTTP server
│   │   ├── middlewares/      # HTTP middleware
│   │   ├── response/         # Response utilities
│   │   └── routes/           # Route configuration
│   └── main.go               # Application entry
├── stress_test/              # Performance tests
```
## Core Components
### 1. HTTP Server (Fiber Framework)
- **Port**: Configurable (default: 8080)
- **Framework**: Go Fiber v2
- **Features**: High performance

### 2. Middleware Layer
- **Auth Middleware**: JWT token validation and user authentication
- **Logger Middleware**: Request/response logging with request ID
- **Error Handler**: Centralized error handling
- **Token Ban Middleware**: Validate banned tokens

### 3. Handler Layer
- **Auth Handler**: Authentication endpoints (login, refresh, token management)
- **Home Handler**: Home screen data aggregation
- **User Handler**: User profile management

### 4. Service Layer (Business Logic)
- **Auth Service**: Authentication business logic
- **JWT Service**: Token generation, validation, and refresh
- **Home Service**: Home data aggregation logic
- **User Service**: User management logic

### 5. Repository Layer (Data Access)
- **Auth Repository**: User authentication data access
- **Home Repository**: Home screen data access
- **User Repository**: User profile data access

### 6. Database Layer
- **MySQL**: Primary database for persistent data
- **Redis**: Caching layer for sessions and temporary data


## Key Features

### 1. Authentication System
- PIN-based authentication
- JWT access and refresh tokens
- Token versioning and banning
- Failed attempt tracking with lockout mechanism

### 2. Home Screen API
- User profile aggregation
- Account balance summaries
- Recent transactions
- Banking cards information
- Promotional banners

### 3. Security Features
- JWT token validation with ban checking
- Request rate limiting
- Input validation

### 4. Monitoring & Logging
- Structured logging with Zap
- Request ID tracking
- Error tracking

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Framework**: Fiber v2
- **Database**: MySQL 8.0
- **Cache**: Redis
- **JWT**: golang-jwt/jwt/v5
- **ORM**: GORM
- **Logger**: Uber Zap
- **Testing**: Testify

### DevOps
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Load Testing**: K6
- **Database Migration**: Custom migration system

## API Endpoints

### Authentication
- `POST /api/v1/auth/verify-pin` - User authentication
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/tokens` - Ban user tokens
- `GET /api/v1/auth/tokens` - List user tokens

### Home
- `GET /api/v1/home/` - Get home screen data (Protected)

## Configuration

The system uses YAML configuration files:
- `config.yaml` - Main configuration
- `config.docker.yaml` - Docker environment configuration
- `config.example.yaml` - Configuration template

## Deployment

### Docker Deployment
```bash
docker-compose up -d
```

### Manual Deployment
```bash
go build -o banking-api ./src
./banking-api serve
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Load Testing
```bash
k6 run stress_test/stress-test.js
```

## Performance Characteristics

- **Concurrency**: High concurrency support with Fiber
- **Caching**: Redis for session and temporary data
- **Database**: Optimized queries with proper indexing
- **Load Balancing**: Support for horizontal scaling

## Security Considerations

1. **Authentication**: Strong JWT-based authentication
2. **Authorization**: Role-based access control
3. **Data Protection**: Encrypted sensitive data
4. **Input Validation**: Comprehensive input validation
5. **Rate Limiting**: Request rate limiting per client
6. **Token Security**: Token versioning and banning system
