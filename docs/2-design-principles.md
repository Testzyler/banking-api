# Banking API - หลักการและเหตุผลในการออกแบบระบบ

## 1. Architecture และ Design Patterns

### 1.1 Clean Architecture
**หลักการ:** ใช้ Clean Architecture เพื่อแยกความรับผิดชอบของแต่ละ layer อย่างชัดเจน

**เหตุผล:**
- **Separation of Concerns**: แต่ละ layer มีหน้าที่เฉพาะ ทำให้ code มี maintainability สูง
- **Testability**: สามารถทำ unit test ได้ง่าย โดยการ mock dependencies
- **Flexibility**: สามารถเปลี่ยน database หรือ framework ได้โดยไม่กระทบ business logic
- **Scalability**: เพิ่ม features ใหม่ได้ง่าย โดยไม่ส่งผลกระทบต่อ code เดิม

### 1.2 Repository Pattern
**หลักการ:** แยก data access logic ออกจาก business logic

**เหตุผล:**
- **Database Independence**: เปลี่ยน database ได้โดยไม่แก้ business logic
- **Testing**: Mock repository ได้ง่ายสำหรับ unit testing
- **Consistency**: มี interface เดียวกันสำหรับการเข้าถึงข้อมูล
- **Caching**: สามารถเพิ่ม caching layer ได้ใน repository

### 1.3 Dependency Injection
**หลักการ:** Inject dependencies ผ่าน constructor

**เหตุผล:**
- **Loose Coupling**: Components ไม่ depend กัน
- **Easy Testing**: สามารถ inject mock objects สำหรับ testing
- **Configuration Flexibility**: เปลี่ยน implementation ได้ง่าย

## 2. Authentication และ Security

### 2.1 JWT Token Strategy
**หลักการ:** ใช้ JWT tokens แยกเป็น Access Token และ Refresh Token

**เหตุผล:**
- **Scalability**: Support horizontal scaling ได้ดี
- **Security**: Access token มี expiry สั้น (15 นาที) ลด risk จาก token theft
- **User Experience**: Refresh token ทำให้ user ไม่ต้อง login บ่อย

### 2.2 Token Versioning System
**หลักการ:** ใช้ timestamp เป็น token version เพื่อ invalidate tokens เก่า

**เหตุผล:**
- **Mass Token Revocation**: สามารถ revoke tokens ทั้งหมดของ user ได้ในครั้งเดียว
- **Security Incident Response**: เมื่อมี security breach สามารถ invalidate tokens ได้ทันที
- **Account Security**: เมื่อ user เปลี่ยน pin / password หรือสงสัยว่า account ถูก hijack

### 2.3 Token Banning System
**หลักการ:** เก็บ banned token IDs ใน Redis cache

**เหตุผล:**
- **Immediate Effect**: Token ถูก ban ทันที ไม่ต้องรอ expiry
- **Performance**: Redis มีความเร็วสูงในการ check ban status

### 2.4 Failed Attempt Protection
**หลักการ:** ล็อค PIN หลังจาก failed attempts 3 ครั้ง

**เหตุผล:**
- **Brute Force Protection**: ป้องกันการทดลอง PIN แบบ brute force
- **Account Security**: ปกป้อง user account จากการเข้าถึงโดยไม่ได้รับอนุญาต


## 3. Database

### 3.1 GORM ORM Framework
**หลักการ:** ใช้ GORM เป็น Object-Relational Mapping (ORM) framework

**เหตุผล:**
- **Type Safety**: รองรับ Go's type system ช่วยลด runtime errors
- **Developer Productivity**: ลดเวลาการเขียน SQL queries แบบ manual
- **Database Migration**: มี built-in migration system ที่ใช้งานง่าย
- **Association Handling**: จัดการ relationships ระหว่าง models ได้อย่างมีประสิทธิภาพ
- **Connection Pooling**: มี connection pooling built-in
- **Database Agnostic**: สามารถเปลี่ยน database ได้ง่าย (MySQL, PostgreSQL, SQLite)
- **Active Community**: มี community support และ documentation ที่ดี

### 3.3 Redis สำหรับ Caching
**หลักการ:** ใช้ Redis เป็น cache layer

**เหตุผล:**
- **Performance**: ความเร็วสูงมากสำหรับ in-memory operations
- **Session Storage**: เหมาะกับการเก็บ temporary data เช่น login sessions
- **Token Management**: เหมาะกับการเก็บ banned tokens และ PIN lockout data
- **Expiration**: รองรับ automatic expiration ของ data
- **Data Structures**: รองรับ data structures ที่หลากหลาย

### 3.4 Database Migration System
**หลักการ:** ใช้ custom migration system ร่วมกับ GORM

**เหตุผล:**
- **Version Control**: Track การเปลี่ยนแปลง database schema
- **Deployment Safety**: Rollback ได้เมื่อมีปัญหา
- **Team Collaboration**: หลายคนสามารถทำงานร่วมกันได้
- **Environment Consistency**: Database schema เหมือนกันทุก environment
- **Data Seeding**: สามารถ seed ข้อมูลเริ่มต้นได้อย่างมีระบบ

## 4. Performance และ Scalability

### 4.1 Fiber Framework
**หลักการ:** ใช้ Fiber เป็น web framework

**เหตุผล:**
- **High Performance**: มี performance สูงเนื่องจากใช้ fasthttp
- **Memory Efficient**: ใช้ memory น้อยกว่า frameworks อื่น
- **Built-in Features**: มี middleware และ utilities ที่จำเป็นครบ

## 5. Security Considerations

### 5.1 Input Validation
**หลักการ:** Validate ข้อมูล input ทุกจุด

**เหตุผล:**
- **SQL Injection Prevention**: ป้องกัน SQL injection attacks
- **Data Integrity**: รับประกันว่าข้อมูลที่เข้าระบบมีรูปแบบถูกต้อง
- **Business Logic Protection**: ป้องกัน business logic จากข้อมูลที่ไม่ถูกต้อง
- **User Experience**: แจ้ง error message ที่ชัดเจนเมื่อข้อมูลไม่ถูกต้อง

### 5.2 Error Handling และ Logging
**หลักการ:** Handle errors อย่างปลอดภัยและ log อย่างครบถ้วน

**เหตุผล:**
- **Information Disclosure Prevention**: ไม่เปิดเผย sensitive information ใน error messages
- **Debugging**: สามารถ debug ปัญหาได้อย่างมีประสิทธิภาพ
- **Monitoring**: ติดตาม health ของระบบ
- **Audit Trail**: เก็บ log สำหรับการตรวจสอบ

## 6. Monitoring และ Observability

### 6.1 Structured Logging
**หลักการ:** ใช้ structured logging ด้วย Zap logger

**เหตุผล:**
- **Machine Readable**: Log format ที่ tools สามารถ parse ได้
- **Efficient**: Performance สูงกว่า traditional logging
- **Searchable**: ค้นหาและ filter logs ได้ง่าย
- **Contextual**: เก็บ context information ที่เป็นประโยชน์

### 6.2 Request ID Tracking
**หลักการ:** ใส่ unique request ID ในทุก request

**เหตุผล:**
- **Request Tracing**: ติดตาม request ได้ตลอด lifecycle
- **Debugging**: เชื่อมโยง logs ที่เกี่ยวข้องกัน
- **Performance Analysis**: วิเคราะห์ performance ของแต่ละ request
- **User Support**: ช่วยในการ support users เมื่อมีปัญหา

### 6.3 Health Checks
**หลักการ:** มี health check endpoints

**เหตุผล:**
- **Load Balancer Integration**: Load balancer ตรวจสอบ health ได้
- **Monitoring Integration**: Monitoring tools ตรวจสอบสถานะได้
- **Automated Recovery**: Auto-restart เมื่อ health check fail
- **Operational Visibility**: Operations team เห็นสถานะระบบได้

