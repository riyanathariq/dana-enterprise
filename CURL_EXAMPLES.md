# cURL Examples - DANA Enterprise API

Kumpulan contoh cURL untuk testing semua endpoint DANA Enterprise API.

## 📋 Base URL

```
http://localhost:3150
```

---

## 🏥 Health Check

```bash
curl -X GET http://localhost:3150/health
```

**Response:**
```json
{
  "status": "ok",
  "message": "Dana Enterprise API is running"
}
```

---

## 👤 Merchant Info

### Get Merchant Info (Default)

```bash
curl -X GET http://localhost:3150/api/v1/merchant/info
```

### Get Merchant Info (Specific Merchant ID)

```bash
curl -X GET http://localhost:3150/api/v1/merchant/info/216620000031042445415
```

---

## 💳 Payment Methods

### Get Available Payment Methods

```bash
curl -X GET http://localhost:3150/api/v1/order/payment/method
```

**Response:**
```json
{
  "success": true,
  "message": "Payment method retrieved successfully",
  "data": {
    "paymentInfos": [
      {
        "payMethod": "VIRTUAL_ACCOUNT",
        "payOption": "VIRTUAL_ACCOUNT_BNI"
      },
      {
        "payMethod": "NETWORK_PAY",
        "payOption": "NETWORK_PAY_PG_QRIS"
      }
    ],
    "responseCode": "2005700",
    "responseMessage": "Successful"
  }
}
```

---

## 🛒 Create Order - Hosted Checkout (Redirect)

**Endpoint:** `POST /api/v1/order`

Hosted Checkout menggunakan redirect ke halaman DANA. User akan memilih payment method di halaman DANA.

### Basic Hosted Checkout

```bash
curl -X POST http://localhost:3150/api/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-HOSTED-001",
    "amount": {
      "value": "10000.00",
      "currency": "IDR"
    },
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "N"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### Hosted Checkout dengan Optional Fields

```bash
curl -X POST http://localhost:3150/api/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-HOSTED-002",
    "merchant_id": "216620000031042445415",
    "amount": {
      "value": "50000.00",
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
    ],
    "valid_up_to": "2025-11-05T15:00:00+07:00",
    "disabled_pay_methods": "VIRTUAL_ACCOUNT"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "responseCode": "2005700",
    "responseMessage": "Successful",
    "partnerReferenceNo": "ORDER-HOSTED-001",
    "webRedirectUrl": "https://m.dana.id/...",
    "referenceNo": "DANA-REF-123456"
  }
}
```

**Note:** Gunakan `webRedirectUrl` untuk redirect user ke halaman pembayaran DANA.

---

## 🛒 Create Order - Custom Checkout (Host-to-Host) - Raw HTTP Request

**Endpoint:** `POST /api/v1/order/custom` atau `POST /api/v1/order` (dengan pay_option_details)

Custom Checkout memerlukan `pay_option_details` untuk menentukan payment method langsung.

**⚠️ Note:** Endpoint ini sekarang menggunakan **raw HTTP request** langsung ke DANA API endpoint `/payment-gateway/v1.0/debit/payment-host-to-host.htm` dengan signature generation manual (tanpa SDK).

### QRIS Payment

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-QRIS-001",
    "amount": {
      "value": "10000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_QRIS",
        "trans_amount": {
          "value": "10000.00",
          "currency": "IDR"
        }
      }
    ],
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "N"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### GoPay Payment

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-GOPAY-001",
    "amount": {
      "value": "25000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_GOPAY",
        "trans_amount": {
          "value": "25000.00",
          "currency": "IDR"
        }
      }
    ],
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "Y"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### Virtual Account BNI

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-VA-BNI-001",
    "amount": {
      "value": "50000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "VIRTUAL_ACCOUNT",
        "pay_option": "VIRTUAL_ACCOUNT_BNI",
        "trans_amount": {
          "value": "50000.00",
          "currency": "IDR"
        }
      }
    ],
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "N"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### Multiple Payment Methods

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-MULTI-001",
    "amount": {
      "value": "75000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_QRIS",
        "trans_amount": {
          "value": "75000.00",
          "currency": "IDR"
        }
      },
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_GOPAY",
        "trans_amount": {
          "value": "75000.00",
          "currency": "IDR"
        }
      },
      {
        "pay_method": "VIRTUAL_ACCOUNT",
        "pay_option": "VIRTUAL_ACCOUNT_BNI",
        "trans_amount": {
          "value": "75000.00",
          "currency": "IDR"
        }
      }
    ],
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "N"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### Custom Checkout dengan Fee

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-FEE-001",
    "amount": {
      "value": "100000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_QRIS",
        "trans_amount": {
          "value": "100000.00",
          "currency": "IDR"
        },
        "fee_amount": {
          "value": "2500.00",
          "currency": "IDR"
        }
      }
    ],
    "url_params": [
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "PAY_RETURN",
        "is_deeplink": "N"
      },
      {
        "url": "https://webhook.site/0f245868-e536-4666-808e-3604fe97b01f",
        "type": "NOTIFICATION",
        "is_deeplink": "N"
      }
    ]
  }'
```

### Custom Checkout dengan Optional Fields

```bash
curl -X POST http://localhost:3150/api/v1/order/custom \
  -H "Content-Type: application/json" \
  -d '{
    "partner_reference_no": "ORDER-CUSTOM-FULL-001",
    "merchant_id": "216620000031042445415",
    "amount": {
      "value": "200000.00",
      "currency": "IDR"
    },
    "pay_option_details": [
      {
        "pay_method": "NETWORK_PAY",
        "pay_option": "NETWORK_PAY_PG_QRIS",
        "trans_amount": {
          "value": "200000.00",
          "currency": "IDR"
        }
      }
    ],
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
    ],
    "sub_merchant_id": "SUB123",
    "external_store_id": "STORE456",
    "valid_up_to": "2025-11-05T16:00:00+07:00",
    "disabled_pay_methods": "VIRTUAL_ACCOUNT"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Order created successfully (Custom Checkout)",
  "data": {
    "responseCode": "2005700",
    "responseMessage": "Successful",
    "partnerReferenceNo": "ORDER-CUSTOM-QRIS-001",
    "referenceNo": "DANA-REF-123456",
    "additionalInfo": {
      "paymentCode": "QRIS_CODE_123456"
    }
  }
}
```

**Note:** Untuk QRIS dan Virtual Account, response akan berisi `paymentCode` yang bisa ditampilkan ke user.

---

## 📊 Get Order Status

```bash
curl -X GET http://localhost:3150/api/v1/order/ORDER-CUSTOM-QRIS-001
```

**Response:**
```json
{
  "success": true,
  "message": "Order retrieved successfully",
  "data": {
    "responseCode": "2005700",
    "responseMessage": "Successful",
    "originalPartnerReferenceNo": "ORDER-CUSTOM-QRIS-001",
    "referenceNo": "DANA-REF-123456",
    "status": "SUCCESS"
  }
}
```

---

## 📝 Notes

### Hosted Checkout vs Custom Checkout

**Hosted Checkout (Redirect):**
- ✅ Tidak perlu `pay_option_details`
- ✅ User pilih payment method di halaman DANA
- ✅ Response berisi `webRedirectUrl` untuk redirect
- ✅ Lebih mudah implementasi
- ❌ User harus di-redirect ke DANA

**Custom Checkout (Host-to-Host):**
- ✅ Wajib `pay_option_details`
- ✅ Payment method ditentukan langsung
- ✅ Bisa langsung generate QRIS/VA tanpa redirect
- ✅ Response berisi `paymentCode` untuk QRIS/VA
- ❌ Lebih kompleks implementasi

### URL Params

- **PAY_RETURN**: URL untuk redirect user setelah payment selesai
  - `is_deeplink`: "Y" untuk mobile app, "N" untuk web
- **NOTIFICATION**: URL untuk webhook dari DANA (server-to-server)
  - `is_deeplink`: Harus "N" (webhook tidak bisa deeplink)

### Amount Format

- **Wajib**: 2 decimal places untuk IDR
- ✅ `"10000.00"` (benar)
- ❌ `"10000"` (salah)
- ❌ `"10000.0"` (salah)

### Valid Up To Format

- Format: `YYYY-MM-DDTHH:mm:ss+07:00`
- Timezone: Jakarta (GMT+7)
- Maksimal: 1 minggu dari sekarang
- Contoh: `"2025-11-05T15:00:00+07:00"`

### Testing Webhook

Gunakan [webhook.site](https://webhook.site) untuk testing webhook:
1. Buka https://webhook.site
2. Copy URL yang diberikan
3. Gunakan sebagai `NOTIFICATION` URL
4. Monitor webhook yang masuk di webhook.site

---

## 🔍 Troubleshooting

### Error: pay_option_details is required

**Solusi:** Pastikan request ke `/api/v1/order/custom` atau `/api/v1/order` dengan `pay_option_details` tidak kosong.

### Error: Invalid Field Format amount

**Solusi:** Pastikan amount menggunakan format 2 decimal places: `"10000.00"` bukan `"10000"`.

### Error: Invalid URL format

**Solusi:** Pastikan URL di `url_params` lengkap dengan `http://` atau `https://`.

### Error: 500 Internal Server Error

**Solusi:**
1. Cek `.env` file sudah benar
2. Cek merchant sudah dikonfigurasi di DANA Dashboard
3. Enable `DANA_DEBUG=true` untuk melihat detail error
4. Hubungi DANA Support jika masih error
