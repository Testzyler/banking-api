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
    - [ ] Implement logging and monitoring
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
    - [ ] `POST /api/v1/auth/login` -  Login with account no
    - [ ] `POST /api/v1/auth/pin/validate` - validate pin

2. **Transaction & UI APIs for Lazy Load**
    - [ ] `GET /api/v1/users/greeting`
    - [ ] `GET /api/v1/transactions`
    - [ ] `GET /api/v1/banners`

3. **Account & Debit card API**
    - [ ] `GET /api/v1/cards`
    - [ ] `GET /api/v1/accounts`
    - [ ] `POST /api/v1/accounts/deposit`
    - [ ] `POST /api/v1/accounts/withdrawal`
    - [ ] `GET /api/v1/accounts/details` - Get all details in 1 API [transactions, banners, greeting msg, account details]

### Phase 3: Security and Middleware
1. **Authentication & Authorization**
   - [ ] Implement JWT token authentication with refresh tokens
   - [ ] Add request validation middleware
   - [ ] Implement rate limiting (especially for PIN attempts)
   - [ ] Add CORS configuration
   - [ ] Implement session management

2. **PIN Security Implementation**
   - [ ] Secure PIN storage using bcrypt/argon2 hashing
   - [ ] PIN attempt tracking and lockout mechanism
   - [ ] Implement PIN requirements (6-digit numeric)
   - [ ] Add PIN aging
   - [ ] PIN brute-force protection with exponential backoff

3. **Data Validation**
   - [ ] Input validation for all endpoints
   - [ ] Sanitization of user inputs
   - [ ] Request/response logging
   - [ ] Error handling middleware
   - [ ] PIN format validation and strength checking

### Phase 4: Testing 
1. **Unit Testing**
    - [ ] Repository layer tests
    - [ ] Service layer tests
    - [ ] Handler
    - [ ] Mock database for testing

2. **Integration Testing**
    - [ ] API endpoint testing
    - [ ] Database integration tests
    - [ ] End-to-end workflow tests

3. **Load Testing**
    - [ ] Performance benchmarks with k6
    - [ ] Concurrent user testing
    - [ ] Database query optimization
   
### Phase 5: Documentation and Deployment
1. **API Documentation**
    - [ ] README with setup instructions
    - [ ] API usage examples

2. **Deployment Preparation**
    - [ ] Production Docker configuration
    - [ ] Environment-specific configurations
    - [ ] Health check endpoints