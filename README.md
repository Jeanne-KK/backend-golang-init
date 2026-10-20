# initbackgo

REST API backend สำหรับระบบยืนยันตัวตน (สมัคร / เข้าสู่ระบบ) และจัดการผู้ใช้ สร้างด้วย Go, Gin, GORM และ MySQL

## คุณสมบัติ

- สมัครสมาชิกและเข้าสู่ระบบด้วยอีเมลและรหัสผ่าน
- รหัสผ่านเข้ารหัสด้วย Argon2id
- Access token (JWT) ส่งใน response body
- Refresh token เก็บใน HttpOnly cookie และเก็บ hash ในฐานข้อมูล
- CORS และตรวจสอบ origin สำหรับ endpoint refresh
- Role-based access: `admin` / `user`

## ความต้องการของระบบ

- Go 1.26+
- MySQL 8+
- goose สำหรับ migration

## เริ่มต้นใช้งาน

### 1. Clone และติดตั้ง dependencies

```bash
go mod download
```

### 2. ตั้งค่า environment

```bash
cp .env.example .env
```

แก้ไข `.env` โดยเฉพาะ:

| ตัวแปร | คำอธิบาย |
|--------|----------|
| `JWT_ACCESS_SECRET` | สตริงสุ่มยาวอย่างน้อย 32 ตัวอักษร (บังคับ) |
| `DB_*` | การเชื่อมต่อ MySQL |
| `CORS_ALLOW_ORIGINS` | origin ของ frontend (คั่นด้วย comma ได้หลายค่า) |
| `GOOSE_DBSTRING` | DSN สำหรับ goose (ควรตรงกับ `DB_*`) |

สร้าง database ก่อนรัน migration:

```sql
CREATE DATABASE backend CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. Migration

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir ./migration mysql "$GOOSE_DBSTRING" up
```

หรือ export จาก `.env`:

```bash
export $(grep -v '^#' .env | xargs)
goose -dir ./migration mysql "$GOOSE_DBSTRING" up
```

### 4. รันเซิร์ฟเวอร์

```bash
go run ./cmd/server
```

เซิร์ฟเวอร์ฟังที่ `SERVER_ADDR` (ค่าเริ่มต้น `:8080`)

### 5. ทดสอบ

```bash
go test ./...
```

## การยืนยันตัวตน

1. **Register / Login** — ได้ `access_token` ใน JSON และ refresh token ใน cookie (`COOKIE_REFRESH_NAME`, path ตาม `COOKIE_REFRESH_PATH`)
2. **API ที่ต้องล็อกอิน** — ส่ง header `Authorization: Bearer <access_token>`
3. **Refresh** — `POST /api/auth/refresh` พร้อม cookie refresh (frontend ต้องส่ง `credentials: 'include'` และ origin ต้องอยู่ใน `CORS_ALLOW_ORIGINS`)
4. **Logout** — ลบ refresh ในฐานข้อมูล (ถ้ามี access token จะ revoke ตาม user) และล้าง cookie

ใน production ตั้ง `COOKIE_SECURE=true` และใช้ HTTPS

## API

Base path: `/api`

### Auth (ไม่ต้องมี token ยกเว้นที่ระบุ)

| Method | Path | คำอธิบาย |
|--------|------|----------|
| POST | `/auth/register` | สมัครสมาชิก |
| POST | `/auth/login` | เข้าสู่ระบบ |
| POST | `/auth/refresh` | ต่ออายุ access token (cookie refresh) |
| POST | `/auth/logout` | ออกจากระบบ (auth แบบ optional) |
| GET | `/auth/me` | ข้อมูลผู้ใช้ปัจจุบัน (ต้องมี Bearer token) |

**Register body**

```json
{
  "name": "ชื่อ",
  "email": "user@example.com",
  "password": "password123"
}
```

**Login body**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Auth success response**

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "name": "ชื่อ",
    "email": "user@example.com",
    "created_at": "..."
  }
}
```

### Users (ต้องมี Bearer token + role `admin`)

| Method | Path | คำอธิบาย |
|--------|------|----------|
| GET | `/users` | รายชื่อผู้ใช้ทั้งหมด |

ผู้ใช้ใหม่จาก register ได้ role `user` โดยค่าเริ่มต้น — ต้องอัปเดต role เป็น `admin` ในฐานข้อมูลเองเพื่อใช้ `/users`

## โครงสร้างโปรเจกต์

```
cmd/server/          entrypoint
internal/
  config/            โหลด .env
  database/          MySQL + GORM
  handler/           HTTP handlers
  middleware/        auth, CORS, role
  model/             entities
  repository/        data access
  service/           business logic
  router/            route definitions
  dto/               request/response types
  pkg/               jwt, password, cookie, tokenhash
migration/           SQL migrations (goose)
```
