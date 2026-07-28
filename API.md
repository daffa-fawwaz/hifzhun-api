# Hifzhun API Documentation

Base URL: `/api/v1`

## Buku dalam Kelas (Khusus Guru)

### Buat Buku Langsung di Kelas

**POST** `/classes/{class_id}/books/create`

Hanya guru pemilik kelas yang dapat menggunakan endpoint ini, dan kelas harus memiliki `type: "book"`. Buku dibuat sebagai draft dan langsung ditautkan ke kelas. Karena itu seluruh siswa yang tergabung di kelas dapat melihat buku, modul, dan itemnya melalui endpoint kelas/buku yang ada.

Gunakan `multipart/form-data`. Field `cover_image` wajib berupa file PNG, JPG, JPEG, atau WEBP dengan ukuran maksimal 3 MB. File diunggah ke Supabase Storage menggunakan `SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY` (atau `SUPABASE_ANON_KEY`), dan `SUPABASE_BUCKET`; URL publik Supabase otomatis disimpan sebagai `cover_image` pada buku.

| Field | Tipe | Wajib | Keterangan |
| --- | --- | --- | --- |
| `title` | text | Ya | Judul buku |
| `description` | text | Tidak | Deskripsi buku |
| `order` | number | Tidak | Urutan buku di kelas |
| `cover_image` | file | Ya | Berkas cover buku |

Contoh:

```bash
curl -X POST "$BASE_URL/api/v1/classes/$CLASS_ID/books/create" \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=Modul Tajwid Dasar" \
  -F "description=Materi untuk kelas tajwid" \
  -F "order=1" \
  -F "cover_image=@/path/ke/cover.jpg"
```

Setelah buku dibuat, guru mengelola modul dan item menggunakan endpoint biasa berikut dengan `book_id` dari respons: `POST /books/{book_id}/modules`, `POST /books/{book_id}/items`, dan `POST /books/modules/{module_id}/items`.

## Authentication

### Register
**POST** `/auth/register`

Request Body:
```json
{
  "username": "string",
  "email": "string",
  "password": "string",
  "full_name": "string",
  "school": "string",
  "domicile": "string"
}
```

Response (201):
```json
{
  "success": true,
  "message": "registration success",
  "data": {
    "id": "uuid",
    "email": "string",
    "role": "student"
  }
}
```

---

### Login
**POST** `/auth/login`

Request Body:
```json
{
  "email": "string",
  "password": "string"
}
```

Response (200):
```json
{
  "success": true,
  "message": "login success",
  "data": {
    "id": "uuid",
    "email": "string",
    "name": "string",
    "role": "string",
    "token": "jwt_token"
  }
}
```

---

## User Endpoints (Requires Authentication)

Header: `Authorization: Bearer <token>`

### Request to Become Teacher
**POST** `/user/teacher-request`

Request Body:
```json
{
  "message": "string"
}
```

Response (201):
```json
{
  "success": true,
  "message": "teacher request submitted successfully",
  "data": null
}
```

---

### Get My Teacher Request Status
**GET** `/user/teacher-request/status`

Response (200):
```json
{
  "status": 200,
  "message": "teacher request status fetched successfully",
  "data": {
    "status": "none|pending|approved|rejected",
    "can_request": true,
    "request_id": "uuid"
  }
}
```

Notes:
- `request_id` is only present when the user has submitted a request before
- `can_request` is `true` when the user has no request or their latest request was rejected

---

### Get My Teacher Request
**GET** `/user/teacher-request`

Response (200):
```json
{
  "success": true,
  "message": "teacher request fetched successfully",
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "message": "string",
    "status": "pending|approved|rejected",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

---

## Admin Endpoints (Requires Admin Role)

Header: `Authorization: Bearer <token>`

### Get All Users
**GET** `/admin/users`

Query Parameters:
- `role` (optional): Filter by role (`student`, `teacher`, `admin`)

Response (200):
```json
{
  "success": true,
  "message": "users fetched successfully",
  "data": [
    {
      "id": "uuid",
      "username": "string",
      "email": "string",
      "role": "string",
      "is_active": true,
      "plan": "free",
      "full_name": "string",
      "school": "string",
      "domicile": "string",
      "avatar_url": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

---

### Activate User
**POST** `/admin/users/:id/activate`

Response (200):
```json
{
  "success": true,
  "message": "user activated successfully",
  "data": null
}
```

---

### Deactivate User
**POST** `/admin/users/:id/deactivate`

Response (200):
```json
{
  "success": true,
  "message": "user deactivated successfully",
  "data": null
}
```

---

### Get Pending Teacher Requests
**GET** `/admin/teacher-requests`

Response (200):
```json
{
  "success": true,
  "message": "teacher requests fetched successfully",
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "message": "string",
      "status": "pending",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "user": {
        "id": "uuid",
        "username": "string",
        "email": "string",
        "full_name": "string"
      }
    }
  ]
}
```

---

### Approve Teacher Request
**POST** `/admin/teacher-requests/:id/approve`

Response (200):
```json
{
  "success": true,
  "message": "teacher request approved successfully",
  "data": null
}
```

---

### Reject Teacher Request
**POST** `/admin/teacher-requests/:id/reject`

Response (200):
```json
{
  "success": true,
  "message": "teacher request rejected successfully",
  "data": null
}
```

---

## Error Response Format

```json
{
  "success": false,
  "message": "error message",
  "error_code": "ERROR_CODE",
  "data": null
}
```

Common Error Codes:
- `BAD_REQUEST` - Invalid request body
- `UNAUTHORIZED` - Invalid or missing token
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `REGISTER_FAILED` - Registration failed
- `LOGIN_FAILED` - Login failed
- `REQUEST_FAILED` - Teacher request failed
- `APPROVE_FAILED` - Approve request failed
- `REJECT_FAILED` - Reject request failed
