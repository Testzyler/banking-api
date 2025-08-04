# Project Banking-api

## Analysis of Provided Resources

### 1. Database Schema

**Core models**
- `users` - ข้อมูลทั่วไปของ user
- `accounts` - บัญชีธนาคารของ user
- `account_balances` - ยอดเงินคงเหลือ assume เป็นแบบ real-time
- `account_flags` - ฟีเจอร์ของบัญชีที่ระบุได้

**Card Management**
- `debit_cards` - ข้อมูลพื้นฐานบัตร
- `debit_card_details` - หมายเลขบัตร / ข้อมูลผู้ออกบัตร issuer
- `debit_card_status`
- `debit_card_design`

**Transaction & UI**
- `transactions`
- `banners`
- `user_greetings`

**Authentication & Security** 
- `user_pins` - เก็บข้อมูล PIN ของ users ด้วย hashing
- `pin_attemps` - track การใส่ PIN และจัดการการล็อกป้องการ Brute-Force (exponential backoff retry)
- `user_session` - เก็บข้อมูล session สำหรับจัดาการ expire

## Tasks breakdown and Implementation

### Phase 1: Project setup and infrastructure
<!-- **Priority: Critical | Est. time: 1-2 days** -->

1. **Project Structure Setup**
    - [x] Setup project structure with clean architecture
    - [x] Configure environment-specific settings
    - [x] Implement logging (ZAP Logger)
    - [x] Error handling
2. **Database Integration**
    - [x] Set up database connection pool
    - [x] Create data access layer (Repository pattern)
3. **Docker Configuration**
    - [x] Create Dockerfile for the application
    - [x] Set up docker-compose
    - [x] Set up environment variable management

### Phase 2: Core API Development
<!-- **Priority: Critical | Est. time: 3-4 days** -->
1. **Authentication & PIN Management APIs**
    - [x] `POST /api/v1/auth/pin/validate` - validate pin

2. **Dashboard API**
    - [x] `GET /api/v1/dashboard/accounts`

### Phase 3: Security and Middleware
| Failed Attempts | Lock Duration |
| --------------- | ------------- |
| 1–2             | 0             |
| 3               | 10s           |
| 4               | 20s           |
| 5               | 40s           |
| 6               | 80s           |
| 7               | 160s          |
| 8               | 300s (maxed)  |

1. **Authentication & Authorization**
   - [x] Implement JWT token authentication with refresh tokens
   - [x] Add request validators
   - [x] Implement rate limiting (especially for PIN attempts)
   - [x] Add CORS configuration

2. **PIN Security Implementation**
   - [x] Secure PIN storage using bcrypt hashing
   - [x] PIN attempt tracking and lockout mechanism
   - [x] Implement PIN requirements (6-digit numeric)
   - [x] Add PIN aging
   - [x] PIN brute-force protection with exponential backoff

3. **Data Validation**
   - [x] Input validation for all endpoints
   - [x] Sanitization of user inputs
   - [x] Middleware logging
   - [x] Error handling middleware
   - [x] PIN format validation and strength checking

### Phase 4: Testing 
1. **Unit Testing**
    - [x] Repository layer tests
    - [x] Service layer tests
    - [x] Handler
    - [x] Mock database for testing

2. **Integration Testing**
    - [x] API endpoint testing
    - [x] Database integration tests
    - [x] End-to-end workflow tests (k6)

3. **Load Testing**
    - [x] Performance benchmarks with k6
    - [x] Concurrent user testing
    - [x] Database query optimization
   
### Phase 5: Documentation and Deployment
1. **API Documentation**
    - [ ] README with setup instructions
    - [ ] API usage examples

2. **Deployment Preparation**
    - [x] Production Docker configuration
    - [x] Environment-specific configurations
    - [x] Health check endpoints