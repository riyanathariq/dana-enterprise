# DANA Enterprise API

REST API untuk integrasi dengan DANA Payment Gateway menggunakan Gapura Hosted Checkout (Redirect).

## 📋 Konfigurasi

### 1. Environment Variables

Buat file `.env` di root project dengan konfigurasi berikut:

```bash
# DANA API Credentials (Required)
DANA_MERCHANT_ID=216620000031042445415
DANA_CLIENT_ID=2025103111305880384385
DANA_CLIENT_SECRET=659598a3e374e77d28d9872e036c2f9e7f3b7526468e73f99f51267ae4eb0913

# Private Key (Required) - Format dengan BEGIN/END markers dan \n untuk newlines
DANA_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDgJ7zOJH1WBapyXXSUn0/IdlKzxMt54BnR5DwyJfug3yStGiKa0KuyYw54PCERgquWEFaWHeneAUIN1LKBu7XaTSkEOH/muvn0sRseHmARBH3cH83lNaDWs65DhMmn10X3V3LAK+kRA+2ZLL6v7kQtgIe7q8YxjpacZfygESGpVrZl7DqlO2YyvhskNVxLDMxVD5pHtref0Os2Nj4lyEfmFSsWeutkxgcD4vAQKfkRcHIi3zQ4nya0VA/ypBNjtzqKxaQTG3clkpV11eW353ny2rZgNzb5nyXwl/6SSH/gqcPbPvoejWe/uAUCZnD3gvYanJwXQT8/by0w5jTZUDSZAgMBAAECggEADCuwLU+vI/lvB6JiIvMy92cQRr090I5dzIcMSytDijivcrwnapk/p1IITjgykfMyupVExEYXxYNzORHGRvPAlBuqKVXhgO9ARNxMZQJgdDAWntn1uZcjtmCfasBWHCg8vuknPH6t2wwH7bAPovkIh+FxjTuuiVCNBi0l7jF77sgh3Sa+IcqN9NBvTLWPA/Y3b6mmY/Ptowwr5LurWT49CPa3hhs+c3XU1gRrmJW+g2bzbhT74elfoF9e77oEwZduOlXFYC2MkNGJfSjFNd87I6dNApQhPppwzz3CGQeWN8wdoRsCKFHxEtC/wOXbwyU1yS/ysmUerLMB7aXnnQRCAQKBgQD2rm67st5Pc9+yjUD7sgcCVNmP5mBoEThiOEurKFgEKqePO/xtKEqBQ3Z1XuUjCtwxA2DRUU3oW9jGNxVJZZeu4R0Dt4hGUdYv74oO22GESj4gCDMWMDnOb4cKOuZ5COSaLc6RNfuwvsf5i6avWBnE+ucmLXOPYHLfYXqAeeCyAQKBgQDon3Zs0ttwV4xuLEFilW1AGEmtWcBP77uMcWg+X5blWv0OEQ/CgSG5r/diS1+JrLmWPPgygvBC/sclykX5mhvs6RXiMrwZiT2Xzr38dZhZZdpY95fwR8b/HrTTOhRyBzgyJFCee3qDDSUP0JkFgnSy3xG6DffN+WS86jj0/gHSmQKBgQDt1gJXoD5tVmggi8ZSpjPR0KMu9cyPqcK2GFcEf9JUuhdxp0FasVUcSkIlKcg8wBTKgNpRFlXKKPvZKHSynmvfpZXG5qZSPkcHUqnGZ0gfN0Gsupse0oJ5gdguSdm6apOV/4JBSU4Q+/BsrnOYbZXy8IH6sinP3AsFSsPEqT22AQKBgQCCCgKNuyHon0hlnl++2IGGPw2Q1odnKEDTteHsXEtiU4b7AhapSL6tquzEChaSQ/hLQDIPKptdGEgDnBuZ+Mh7m6EcHfiA3fOMlYonQyWzc/inm2FYdQeMr2WuXt1nERodDafzsFtAP2zkdlvUdKUACStDsuNARZZG9Th53DTwoQKBgQDI590ZV/v+dNYCxYVgiAXt+jwOvsPuggOEXEKjbqA5xFPISrYISm46EmC3tgHNFf51FY0gSYENRAaGhSv4pBoOO3P79Pn/R2L6YvxCg3ZCu7zh96zWu8EhX5Mge0EJ0mTCFvlGmL8m2u9zz6WTGngqeYvmoQUQX2Rcsv90/tQtnQ==\n-----END PRIVATE KEY-----\n"

# Public Key (Optional, for reference)
DANA_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4Ce8ziR9VgWqcl10lJ9PyHZSs8TLeeAZ0eQ8MiX7oN8krRoimtCrsmMOeDwhEYKrlhBWlh3p3gFCDdSygbu12k0pBDh/5rr59LEbHh5gEQR93B/N5TWg1rOuQ4TJp9dF91dywCvpEQPtmSy+r+5ELYCHu6vGMY6WnGX8oBEhqVa2Zew6pTtmMr4bJDVcSwzMVQ+aR7a3n9DrNjY+JchH5hUrFnrrZMYHA+LwECn5EXByIt80OJ8mtFQP8qQTY7c6isWkExt3JZKVddXlt+d58tq2YDc2+Z8l8Jf+kkh/4KnD2z76Ho1nv7gFAmZw94L2GpycF0E/P28tMOY02VA0mQIDAQAB\n-----END PUBLIC KEY-----\n"

# Sandbox API Configuration
DANA_HOST=api.sandbox.dana.id
DANA_SCHEME=https
DANA_ENV=sandbox

# Optional: X-PARTNER-ID (default: akan menggunakan CLIENT_ID jika tidak diset)
# DANA_X_PARTNER_ID=2025103111305880384385

# Optional: CHANNEL-ID (default: akan diambil dari SDK jika tidak diset)
# DANA_CHANNEL_ID=95221

# Optional: Merchant Category Code (default: 5999 - Miscellaneous)
DANA_MCC=5999

# Optional: Order Title (default: "Order {partnerReferenceNo}")
# DANA_ORDER_TITLE=My Order Title

# Server Configuration
PORT=3150

# Debug Mode (optional, set to "true" untuk melihat request/response detail)
DANA_DEBUG=false

# Gin Mode (optional, set to "debug" untuk development)
# GIN_MODE=release
```

### 2. Format Private Key

**PENTING**: `DANA_PRIVATE_KEY` harus dalam format yang benar:

```bash
DANA_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\nMIIEv...\n-----END PRIVATE KEY-----\n"
```

**Harus ada:**
- `-----BEGIN PRIVATE KEY-----` di awal
- `-----END PRIVATE KEY-----` di akhir
- `\n` untuk setiap newline (bukan literal newline)

**Cara convert dari private key file:**
```bash
# Jika punya file private.key, convert ke format env:
cat private.key | sed ':a;N;$!ba;s/\n/\\n/g' | sed 's/^/DANA_PRIVATE_KEY="/' | sed 's/$/"/'
```

### 3. Konfigurasi di DANA Dashboard

#### A. Aktifkan Payment Gateway

1. Login ke [DANA Dashboard](https://dashboard.dana.id)
2. Pilih merchant Anda
3. Navigasi ke **Settings** → **Payment Gateway**
4. Pastikan **Gapura Hosted Checkout** sudah aktif

#### B. Konfigurasi Merchant Category Code (MCC)

1. Di **Settings** → **Merchant Information**
2. Pastikan **MCC Code** sudah sesuai dengan business type Anda
3. MCC yang umum digunakan:
   - `5999` - Miscellaneous (default)
   - `5411` - Grocery Stores
   - `5812` - Restaurants
   - `5814` - Fast Food Restaurants

#### C. Konfigurasi Webhook URLs

1. Di **Settings** → **Webhook Configuration**
2. Set **Notification URL** (untuk webhook dari DANA)
3. Set **Return URL** (untuk redirect setelah payment)
4. Pastikan URL menggunakan HTTPS (kecuali di localhost untuk testing)

#### D. Verifikasi Credentials

Pastikan di DANA Dashboard:
- **Client ID** sama dengan `DANA_CLIENT_ID`
- **Merchant ID** sama dengan `DANA_MERCHANT_ID`
- **Private Key** sudah di-upload atau match dengan `DANA_PRIVATE_KEY`

### 4. Testing di Sandbox

#### A. Sandbox Credentials

Pastikan menggunakan credentials dari **Sandbox Environment**:
- Sandbox URL: `https://api.sandbox.dana.id`
- Sandbox credentials berbeda dengan Production

#### B. Test Payment Flow

1. **Create Order** → Dapatkan `webRedirectUrl`
2. **Redirect user** ke `webRedirectUrl`
3. **User pilih payment method** di halaman DANA
4. **DANA redirect** ke `PAY_RETURN` URL
5. **DANA kirim webhook** ke `NOTIFICATION` URL

## 🚀 Running

### Install Dependencies

```bash
go mod download
```

### Run Server

```bash
# Development (dengan debug)
DANA_DEBUG=true go run main.go

# Production
go run main.go
```

Server akan berjalan di `http://localhost:3150` (atau sesuai `PORT` env)

## 📡 API Endpoints

### Health Check

```bash
GET /health
```

### Get Merchant Info

```bash
GET /api/v1/merchant/info
GET /api/v1/merchant/info/{merchant_id}
```

### Create Order (Hosted Checkout)

```bash
POST /api/v1/order
Content-Type: application/json

{
  "partner_reference_no": "ORDER-123",
  "amount": {
    "value": "10000.00",
    "currency": "IDR"
  },
  "url_params": [
    {
      "url": "https://yourdomain.com/return",
      "type": "PAY_RETURN",
      "is_deeplink": "N"
    },
    {
      "url": "https://yourdomain.com/webhook",
      "type": "NOTIFICATION",
      "is_deeplink": "N"
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "responseCode": "2005700",
    "responseMessage": "Successful",
    "partnerReferenceNo": "ORDER-123",
    "webRedirectUrl": "https://...",
    "referenceNo": "..."
  }
}
```

### Get Payment Methods

```bash
GET /api/v1/order/payment/method
```

### Get Order Status

```bash
GET /api/v1/order/{partner_reference_no}
```

## 🔍 Troubleshooting

### Error 401: Unauthorized. Invalid Client

**Penyebab:**
- `DANA_CLIENT_ID` atau `DANA_CLIENT_SECRET` salah
- `DANA_X_PARTNER_ID` tidak sesuai (harus sama dengan `CLIENT_ID` untuk auth)
- Private key format salah

**Solusi:**
1. Cek `.env` file, pastikan semua credentials benar
2. Pastikan `DANA_PRIVATE_KEY` format benar (dengan BEGIN/END markers dan `\n`)
3. Pastikan `DANA_X_PARTNER_ID` (jika diset) sama dengan `DANA_CLIENT_ID`

### Error 5005401: Internal Server Error

**Penyebab:**
- Merchant belum dikonfigurasi untuk Payment Gateway di DANA Dashboard
- MCC code tidak valid atau tidak sesuai
- Missing required fields di request

**Solusi:**
1. Cek di DANA Dashboard → Payment Gateway sudah aktif
2. Pastikan MCC code valid (cek di DANA Dashboard)
3. Pastikan semua required fields sudah terisi
4. Hubungi DANA Support dengan:
   - Merchant ID
   - Partner Reference No yang error
   - Error response detail

### Error: Invalid Field Format

**Penyebab:**
- Amount format salah (harus 2 decimal places untuk IDR)
- ValidUpTo format salah
- URL format salah

**Solusi:**
1. Amount harus format: `"10000.00"` (bukan `"10000"` atau `"10000.0"`)
2. ValidUpTo format: `"2025-11-05T12:00:00+07:00"` (Jakarta timezone)
3. URL harus lengkap dengan `http://` atau `https://`

### Webhook Not Received

**Penyebab:**
- URL tidak accessible dari internet
- URL tidak menggunakan HTTPS (kecuali localhost)
- Firewall blocking

**Solusi:**
1. Gunakan public URL (bisa pakai ngrok untuk testing: `ngrok http 3150`)
2. Pastikan endpoint webhook bisa diakses dari internet
3. Test webhook dengan tool seperti webhook.site

## 📝 Notes

- **Sandbox vs Production**: Pastikan credentials dan URL sesuai environment
- **Private Key**: Selalu dalam format PEM dengan BEGIN/END markers
- **Amount**: Selalu 2 decimal places untuk IDR currency
- **Timezone**: Gunakan Jakarta timezone (GMT+7) untuk `validUpTo`
- **Webhook**: Harus HTTPS (kecuali localhost untuk development)

## 🔗 Resources

- [DANA Dashboard](https://dashboard.dana.id)
- [DANA API Documentation](https://dashboard.dana.id/api-docs)
- [DANA Go SDK](https://github.com/dana-id/dana-go)

