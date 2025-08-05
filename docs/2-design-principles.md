# Banking API - Design Principles and System Architecture Rationale

## 1. Architecture and Design Patterns

### 1.1 Clean Architecture
**Principle:** Implement Clean Architecture to clearly separate responsibilities across different layers

**Rationale:**
- **Separation of Concerns**: Each layer has specific responsibilities, resulting in high code maintainability
- **Testability**: Easy to perform unit testing through dependency mocking
- **Flexibility**: Ability to change database or framework without affecting business logic
- **Scalability**: Easy to add new features without impacting existing code

### 1.2 Repository Pattern
**Principle:** Separate data access logic from business logic

**Rationale:**
- **Database Independence**: Change database without modifying business logic
- **Testing**: Easy repository mocking for unit testing
- **Consistency**: Uniform interface for data access
- **Caching**: Ability to add caching layer within repository

### 1.3 Dependency Injection
**Principle:** Inject dependencies through constructor

**Rationale:**
- **Loose Coupling**: Components don't depend on each other
- **Easy Testing**: Ability to inject mock objects for testing
- **Configuration Flexibility**: Easy to change implementations

## 2. Authentication and Security

### 2.1 JWT Token Strategy
**Principle:** Use JWT tokens separated into Access Token and Refresh Token

**Rationale:**
- **Scalability**: Excellent support for horizontal scaling
- **Security**: Short-lived access tokens (15 minutes) reduce risk from token theft
- **User Experience**: Refresh tokens prevent frequent user logins

### 2.2 Token Versioning System
**Principle:** Use timestamp as token version to invalidate old tokens

**Rationale:**
- **Mass Token Revocation**: Ability to revoke all user tokens at once
- **Security Incident Response**: Immediate token invalidation during security breaches
- **Account Security**: Token invalidation when user changes PIN/password or suspects account hijacking

### 2.3 Token Banning System
**Principle:** Store banned token IDs in Redis cache

**Rationale:**
- **Immediate Effect**: Tokens are banned instantly without waiting for expiry
- **Performance**: Redis provides high-speed ban status checking

### 2.4 Failed Attempt Protection
**Principle:** Lock PIN after 3 failed attempts

**Rationale:**
- **Brute Force Protection**: Prevent brute force PIN attacks
- **Account Security**: Protect user accounts from unauthorized access

## 3. Database

### 3.1 GORM ORM Framework
**Principle:** Use GORM as Object-Relational Mapping (ORM) framework

**Rationale:**
- **Type Safety**: Supports Go's type system, reducing runtime errors
- **Developer Productivity**: Reduces time spent writing manual SQL queries
- **Database Migration**: Built-in migration system that's easy to use
- **Association Handling**: Efficient management of relationships between models
- **Connection Pooling**: Built-in connection pooling
- **Database Agnostic**: Easy database switching (MySQL, PostgreSQL, SQLite)
- **Active Community**: Strong community support and excellent documentation

### 3.3 Redis for Caching
**Principle:** Use Redis as cache layer

**Rationale:**
- **Performance**: Extremely high speed for in-memory operations
- **Session Storage**: Ideal for storing temporary data like login sessions
- **Token Management**: Perfect for storing banned tokens and PIN lockout data
- **Expiration**: Supports automatic data expiration
- **Data Structures**: Supports diverse data structures

### 3.4 Database Migration System
**Principle:** Use custom migration system with GORM

**Rationale:**
- **Version Control**: Track database schema changes
- **Deployment Safety**: Rollback capability when issues occur
- **Team Collaboration**: Multiple developers can work together effectively
- **Environment Consistency**: Identical database schema across all environments
- **Data Seeding**: Systematic initial data seeding capability

## 4. Performance and Scalability

### 4.1 Fiber Framework
**Principle:** Use Fiber as web framework

**Rationale:**
- **High Performance**: Superior performance due to fasthttp usage
- **Memory Efficient**: Lower memory usage compared to other frameworks
- **Built-in Features**: Complete set of necessary middleware and utilities

## 5. Security Considerations

### 5.1 Input Validation
**Principle:** Validate input data at every point

**Rationale:**
- **SQL Injection Prevention**: Protect against SQL injection attacks
- **Data Integrity**: Ensure data entering the system has correct format
- **Business Logic Protection**: Protect business logic from invalid data
- **User Experience**: Provide clear error messages for invalid data

### 5.2 Error Handling and Logging
**Principle:** Handle errors securely and log comprehensively

**Rationale:**
- **Information Disclosure Prevention**: Avoid exposing sensitive information in error messages
- **Debugging**: Enable efficient problem debugging
- **Monitoring**: Track system health
- **Audit Trail**: Maintain logs for auditing purposes
