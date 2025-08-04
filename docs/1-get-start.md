
## 🚀 Quick Start Guide

### **Prerequisites**
- [Go 1.24.5+](https://golang.org/doc/install)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Testzyler/banking-api.git
cd banking-api
```

### 2. Download Mock Data

Download the mock data and store it in the seeds directory:

```bash
# Create seeds directory if it doesn't exist
mkdir -p src/database/migrations/seeds

# Download mock data (replace with actual URL)
curl -L "https://https://drive.google.com/file/d/1hGppnGIvL09eHmZ_EXFwNYL1Jj_axjId/view?usp=drive_link" \
  -o src/database/migrations/seeds/*.sql

# Or if you have the mock data file locally, copy it:
# cp /path/to/your/mock_data.sql src/database/migrations/seeds/
```

### 3. Set up Docker Environment Variables

Create a `.env` file in the root directory:

```bash
# Database Configuration
MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=banking
MYSQL_USER=banking_user
MYSQL_PASSWORD=userpassword

# Application Configuration
PORT=8080
```

### 4. Configure the Application

Copy the example configuration file and update it with your database credentials if needed.

```bash
cp src/config.example.yaml src/config.yaml
```

### 5. Start the Services

Run all services using Docker Compose:

```bash
docker-compose up -d --build
```

This will start:
- MySQL database
- Database migrations  - always check for new migrations is added (first time would be take several times to start because it's take seeds into database)
- Banking API service - start after migration completed

### 6. Verify Installation

Check if the API is running:

```bash
curl http://localhost:8080/healthz
```

### 7. Local Development Setup

For local development without Docker:

```bash
# Navigate to source directory
cd src

# Install dependencies
go mod download

# Run migrations
go run . migrate

# Start the API server
go run . serve_api
```

## Available Commands

```bash
# Start API server
go run . serve_api

# Run database migrations
go run . migrate

# Show help
go run . --help
```

## Environment Configuration

The application supports different environments:

- **Development**: `config.yaml`
- **Docker**: `config.docker.yaml`
- **Production**: Environment variables

### Key Configuration Options

```yaml
Server:
  Port: 8080
  Environment: development
  
Database:
  Host: localhost
  Port: 3306
  Username: banking_user
  Password: userpassword
  Name: banking
  
Logger:
  Level: info
  LogColor: true
  LogJson: false
```


## Deployment

### Docker Deployment

```bash
# Production deployment
docker-compose -f docker-compose.yaml up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f banking-api
```
