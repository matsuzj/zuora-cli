# Zuora API Reference - Extracted Documentation for CLI Implementation

Sources: developer.zuora.com (fetched 2026-03-25)

---

## 1. API Overview & Server Endpoints

The Zuora v1 API (latest minor version: `2025-08-12`) supports REST operations across multiple regional servers with authentication via OAuth 2.0, Bearer tokens, or Basic Auth.

### Server Endpoints

| Region | Endpoint |
|--------|----------|
| US Dev/Sandbox | `https://rest.test.zuora.com` |
| US Sandbox Cloud 1 | `https://rest.sandbox.na.zuora.com` |
| US Sandbox Cloud 2 | `https://rest.apisandbox.zuora.com` |
| US Production Cloud 1 | `https://rest.na.zuora.com` |
| US Production Cloud 2 | `https://rest.zuora.com` |
| EU Dev/Sandbox | `https://rest.test.eu.zuora.com` |
| EU Sandbox | `https://rest.sandbox.eu.zuora.com` |
| EU Production | `https://rest.eu.zuora.com` |
| APAC Dev/Sandbox | `https://rest.test.ap.zuora.com` |
| APAC Production | `https://rest.ap.zuora.com` |

---

## 2. Authentication

### OAuth 2.0 (Recommended)
- Endpoint: `POST /oauth/token`
- Creates bearer tokens for API authentication
- Rate limited by IP address
- Tokens expire after one hour

### Bearer Token
- Header: `Authorization: Bearer {{token}}`

### Organization Header (Optional for multi-org)
- Header: `Zuora-Org-Ids: YOUR_ORG_ID`

---

## 3. API Versioning

### Major Version
- Only `v1` exists currently
- Appears in URL path: `POST https://rest.zuora.com/v1/subscriptions`

### Minor Version
- Format: `YYYY-MM-DD` (e.g., `2024-05-20`, `2025-08-12`)
- Legacy numbered formats (`186.0`, `196.0`, etc.) still supported
- Default version: `186.0` if no version specified and tenant not upgraded
- Header: `Zuora-Version` (applies exclusively to v1 API calls)
- Can be set at tenant level or overridden per-request via header

### Supported Minor Versions
187.0, 188.0, 189.0, 196.0, 206.0, 207.0, 211.0, 214.0, 215.0, 216.0, 223.0, 224.0, 230.0, 239.0, 256.0, 257.0, 309.0, 314.0, 315.0, 329.0, 330.0, 336.0, 337.0, 338.0, 341.0, 2024-05-20, 2025-08-12

---

## 4. Request/Response Formats

### Request Body
- Content-Type: `application/json`
- Decimal values: do not use quotation marks, commas, or spaces. Use `+-0-9.eE` chars
- Object keys can be provided as full 32-digit ID or object number (e.g., `accountKey`)

### Timeout
- If a request does not complete within 120 seconds, Zuora returns a Gateway Timeout error

### Idempotency
- Header: `Idempotency-Key` (UUID v4) for POST/PATCH requests
- If retry with same key returns HTTP 409, the original operation succeeded
- Do NOT use for GET, HEAD, OPTIONS, PUT, DELETE (already idempotent)

### Tracking
- Header: `Zuora-Track-Id` for request correlation (returned in response headers)

---

## 5. Accounts API Endpoints

### Create an Account
- **Method:** `POST /v1/accounts`

### Retrieve an Account
- **Method:** `GET /v1/accounts/{account-key}`
- **Path Parameters:**
  - `account-key` (required): Account number or account ID (32-digit)

### Update an Account
- **Method:** `PUT /v1/accounts/{account-key}`

### Delete an Account
- **Method:** `DELETE /v1/accounts/{account-key}`

### Retrieve Account Summary
- **Method:** `GET /v1/accounts/{account-key}/summary`
- **Path Parameters:**
  - `account-key` (required): Account number or account ID

### List Payment Methods of Account
- **Method:** `GET /v1/accounts/{account-key}/payment-methods`

### Retrieve Default Payment Method
- **Method:** `GET /v1/accounts/{account-key}/payment-methods/default`

### Object Query: List Accounts
- **Method:** `GET /object-query/accounts`
- **Query Parameters:** See Object Query section below

### Object Query: Get Account by Key
- **Method:** `GET /object-query/accounts/{account-key}`
- **Query Parameters:**
  - `expand[]`: Expand related objects (e.g., `billto`, `soldto`, `defaultpaymentmethod`, `subscriptions`)
  - `<object>.fields[]`: Specify which fields to return

---

## 6. Subscriptions API Endpoints

### Preview a Subscription
- **Method:** `POST /v1/subscriptions/preview`

### Create a Subscription
- **Method:** `POST /v1/subscriptions`

### List Subscriptions by Account
- **Method:** `GET /v1/subscriptions/accounts/{account-key}`
- **Path Parameters:**
  - `account-key` (required): Account number or account ID
- **Query Parameters:**
  - `pageSize`: Number of results per page (max 40 for v1 GET, defaults to 10)
  - `page`: Page number
- **Response includes:** `nextPage` URL when more results exist

### Retrieve Subscription by Key
- **Method:** `GET /v1/subscriptions/{subscription-key}`
- **Path Parameters:**
  - `subscription-key` (required): Subscription number or subscription ID

### Retrieve Subscription by Key and Version
- **Method:** `GET /v1/subscriptions/{subscription-key}/versions/{version}`

### Update a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}`

### Renew a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}/renew`

### Cancel a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}/cancel`

### Suspend a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}/suspend`

### Resume a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}/resume`

### Delete a Subscription
- **Method:** `PUT /v1/subscriptions/{subscription-key}/delete`

### List Subscription Metrics
- **Method:** `GET /v1/subscriptions/subscription-metrics`

---

## 7. Object Query API (Newer Pattern)

### Base URL Pattern
`GET /object-query/{object-type}`

### Query Parameters

#### fields[] - Field Selection
- Syntax: `{object}.fields[]=field1,field2,field3`
- Example: `account.fields[]=id,name,accountNumber,balance`
- Field order in request does not determine output order
- Use `<object>.fields[]` for both primary and associated objects

#### expand[] - Related Object Expansion
- Syntax: `expand[]=relatedObject`
- Example: `expand[]=billto&billto.fields[]=firstname,lastname`
- Available for Accounts: `billto`, `soldto`, `defaultpaymentmethod`, `subscriptions`, `rateplans`, `rateplancharges`
- Up to three levels of nested expansion supported
- Example: `subscriptions.subscription_plans.subscription_items` (3 levels max)
- Pagination targets base objects, not expanded objects

#### filter[] - Filtering
- Syntax: `filter[]=field.operator:value`
- Multiple filters combined with AND logic (OR not supported)
- Only indexed fields and custom indexed fields supported

**Supported Operators:**

| Operator | Purpose | Example |
|----------|---------|---------|
| EQ | Exact match (case-insensitive for strings) | `currency.EQ:EUR` |
| NE | Not equal | `currency.NE:CAN` |
| LT | Less than | `quantity.LT:200` |
| GT | Greater than | `quantity.GT:200` |
| LE | Less than or equal | |
| GE | Greater than or equal | |
| SW | Starts with (strings only) | `name.SW:Acc` |
| IN | Multiple values | `name.IN:[Amy,Bella]` |

**Date Filtering:**
- Transaction dates: `YYYY-MM-DD` format (no timestamps, no quotes)
- Updated dates: ISO 8601 format: `2024-01-01T00:00:00Z`

**Special Values:**
- Null: `<field>.EQ:null`
- Empty strings: `<field>.EQ:%02%03`

**URL Encoding Required:**
- Spaces: `%20`
- Brackets: `%5B` (left), `%5D` (right)
- Example: `filter[]=currency.IN:%5BCAD,GBP%5D`

#### sort[] - Sorting
- Syntax: `sort[]=field.ASC` or `sort[]=field.DESC`
- Cannot sort on more than one field
- Cannot sort on related object fields

---

## 8. Pagination

### Standard v1 API GET Methods
- **Parameter:** `pageSize` (max 40; larger values treated as 40; defaults to 10)
- **Response:** `nextPage` element contains URL for next page when more rows exist
- Arrays for non-paginated data support up to 300 rows

### Object Query API
- **Parameter:** `pageSize` (max 99; defaults to 10)
- **Cursor-based:** Response includes `nextPage` value
- Use `cursor=<nextPage_value>` to retrieve subsequent pages
- When on final page, `nextPage` is absent from response
- Pagination targets base objects only, not expanded objects
- Maximum 99 expanded objects per base object

### v1 API Query Parameters for Pagination
- `pageSize`: 1-50 range, values outside trigger 400 error
- `cursor`: Starting place in a list for cursor-based pagination

---

## 9. Error Handling

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Processed but may contain business logic errors (check `success` flag) |
| 201 | Resource created successfully |
| 204 | Success with no response body |
| 400 | Malformed/missing required fields |
| 401 | Invalid/missing/expired credentials |
| 403 | Authenticated but lacking permissions |
| 404 | Resource doesn't exist |
| 409 | Request conflicts with current state |
| 429 | Rate limit exceeded |
| 500/502/503/504 | Server errors |

### Error Response Structure (v1 API)

```json
{
  "success": false,
  "processId": "7F2E4C89A1B3C4D5",
  "reasons": [
    {
      "code": 53100320,
      "message": "Invalid value for field termType: must be TERMED or EVERGREEN"
    }
  ]
}
```

### Error Code Structure
- 8-digit code: first 6 digits = resource code, last 2 digits = error category

### Error Category Suffixes

| Suffix | Category | Description |
|--------|----------|-------------|
| 10 | Permission denied | Missing tenant or user permission |
| 11 | Auth failed | Invalid API credentials |
| 20 | Invalid value | Incorrect field format or value |
| 21 | Unknown field | Unknown field in request body |
| 22 | Missing field | Required field absent |
| 23 | Missing param | Required query parameter absent |
| 27 | Invalid param | Incorrect query parameter |
| 30 | Rule violation | Business rule restriction |
| 40 | Not found | Resource cannot be found |
| 45 | Unsupported | HTTP method not supported |
| 50 | Locking | Objects being modified elsewhere |
| 60 | Internal error | Server-side error |
| 61 | Temporary error | Database communication issue |
| 70 | Limit exceeded | Concurrent request limit exceeded |
| 90 | Malformed request | JSON syntax errors |
| 99 | Integration error | External system failure |

### Success Flag Checking

**Must check `success` flag (POST/PUT/PATCH):**
- `/v1/orders`
- `/v1/payment-methods`
- `/v1/accounts`

**Do NOT need to check (GET/Query/DELETE):**
- List/single resource retrieval
- Object Query operations
- DELETE operations (return 204)

### Retryable vs Non-Retryable Errors

**Retry these:**
- Network timeouts
- HTTP 200 + `success: false` with error codes ending in 50, 61, 70, 99
- HTTP 401 (expired token -- get new token, retry once)
- HTTP 429 (rate limit -- respect `Retry-After` header)
- HTTP 500, 502, 503, 504

**Do NOT retry (fix code):**
- Error codes ending in 2X (invalid values)
- Error codes ending in 30 (business rule violations)
- HTTP 400, 403, 404, 409

### Retry Strategy: Exponential Backoff with Jitter
- Formula: `sleep_seconds = random_between(0, min(max_cap, base_delay * (2 ** attempt_number)))`
- Max retries: 5
- Base delay: 1 second
- Max cap: 60 seconds
- Respect `Retry-After` headers from 429/503 responses

---

## 10. Rate and Concurrency Limits

### Rate Limits
- 50,000 requests/minute
- 2.25 million/hour
- 27 million/day

### Concurrency Limits
- Default: 40 simultaneous requests
- Performance Boost: 80
- High-volume: 200 (400 with Performance Booster)
- Object Queries: 80

### Response Headers
- `Rate-Limit-*`: Indicates remaining rate limit capacity
- `Concurrency-Limit-*`: Indicates remaining concurrency capacity
- `Retry-After`: Present on 429 responses, indicates wait time

---

## 11. All Account Endpoints (Complete List)

| Operation | Method | Path |
|-----------|--------|------|
| Create an account | POST | `/v1/accounts` |
| Retrieve an account | GET | `/v1/accounts/{account-key}` |
| Update an account | PUT | `/v1/accounts/{account-key}` |
| Delete an account | DELETE | `/v1/accounts/{account-key}` |
| Retrieve account summary | GET | `/v1/accounts/{account-key}/summary` |
| List payment methods of account | GET | `/v1/accounts/{account-key}/payment-methods` |
| Retrieve cascading payment methods config | GET | `/v1/accounts/{account-key}/payment-methods/cascading` |
| Configure cascading payment methods | PUT | `/v1/accounts/{account-key}/payment-methods/cascading` |
| Retrieve default payment method | GET | `/v1/accounts/{account-key}/payment-methods/default` |

---

## 12. All Subscription Endpoints (Complete List)

| Operation | Method | Path |
|-----------|--------|------|
| Preview a subscription | POST | `/v1/subscriptions/preview` |
| Preview existing subscription | POST | `/v1/subscriptions/{subscription-key}/preview` |
| Create a subscription | POST | `/v1/subscriptions` |
| List subscriptions by account | GET | `/v1/subscriptions/accounts/{account-key}` |
| Retrieve subscription by key | GET | `/v1/subscriptions/{subscription-key}` |
| Retrieve by key and version | GET | `/v1/subscriptions/{subscription-key}/versions/{version}` |
| Update a subscription | PUT | `/v1/subscriptions/{subscription-key}` |
| Renew a subscription | PUT | `/v1/subscriptions/{subscription-key}/renew` |
| Cancel a subscription | PUT | `/v1/subscriptions/{subscription-key}/cancel` |
| Suspend a subscription | PUT | `/v1/subscriptions/{subscription-key}/suspend` |
| Resume a subscription | PUT | `/v1/subscriptions/{subscription-key}/resume` |
| Delete a subscription | PUT | `/v1/subscriptions/{subscription-key}/delete` |
| Update custom fields (version) | PUT | `/v1/subscriptions/{subscriptionNumber}/versions/{version}/customFields` |
| List subscription metrics | GET | `/v1/subscriptions/subscription-metrics` |

---

## 13. All Other Key Endpoints

### Orders

| Operation | Method | Path |
|-----------|--------|------|
| Preview an order | POST | `/v1/orders/preview` |
| List all orders | GET | `/v1/orders` |
| Create an order | POST | `/v1/orders` |
| Retrieve an order | GET | `/v1/orders/{orderNumber}` |
| Update an order | PUT | `/v1/orders/{orderNumber}` |
| Delete an order | DELETE | `/v1/orders/{orderNumber}` |
| List by subscription owner | GET | `/v1/orders/subscriptionOwner/{accountNumber}` |
| List by subscription number | GET | `/v1/orders/subscription/{subscriptionNumber}` |
| List by invoice owner | GET | `/v1/orders/invoiceOwner/{accountNumber}` |

### Contacts

| Operation | Method | Path |
|-----------|--------|------|
| Create a contact | POST | `/v1/contacts` |
| Retrieve a contact | GET | `/v1/contacts/{contactId}` |
| Update a contact | PUT | `/v1/contacts/{contactId}` |
| Delete a contact | DELETE | `/v1/contacts/{contactId}` |

### Usage

| Operation | Method | Path |
|-----------|--------|------|
| Upload usage file | POST | `/v1/usage` |
| Create usage record | POST | `/v1/object/usage` |
| Retrieve usage record | GET | `/v1/object/usage/{id}` |

---

## 14. Sample Account Response (from Object Query)

```json
{
  "accountNumber": "A00000029",
  "balance": 149.97,
  "creditBalance": 0.0,
  "autoPay": true,
  "billCycleDay": 1,
  "paymentTerm": "Net 30",
  "currency": "USD",
  "status": "Active",
  "name": "Amy Lawrence",
  "createdDate": "...",
  "updatedDate": "...",
  "billTo": {
    "firstName": "Amy",
    "lastName": "Lawrence",
    "state": "California"
  }
}
```

---

## 15. Combined Object Query Example

```bash
curl --request GET \
  --url 'https://rest.apisandbox.zuora.com/object-query/accounts?account.fields[]=name,accountNumber,balance&expand[]=billto&billto.fields[]=firstname,lastname&filter[]=currency.EQ:CAD&sort[]=name.DESC&pageSize=20' \
  -H "Authorization: Bearer $ztoken" \
  -H "Content-Type: application/json"
```

### Response with Pagination

```json
{
  "data": [
    {
      "accountNumber": "A00024791",
      "balance": 0.0,
      "name": "Oh Canada",
      "billTo": {
        "firstName": "...",
        "lastName": "..."
      }
    }
  ],
  "nextPage": "cursor_value_here"
}
```

When `nextPage` is present, append `cursor=<nextPage_value>` to get the next page. When on the final page, `nextPage` is absent.

---

## 16. Request Headers Reference

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | `Bearer {token}` |
| `Content-Type` | Yes (for POST/PUT) | `application/json` |
| `Zuora-Version` | No | API minor version (e.g., `2025-08-12`) |
| `Zuora-Track-Id` | No | Correlation ID for request tracking |
| `Zuora-Org-Ids` | No | Organization ID for multi-org |
| `Idempotency-Key` | No | UUID v4 for POST/PATCH idempotency |

---

## 17. Async Operations

Some operations support async execution:

| Operation | Method | Path |
|-----------|--------|------|
| Create order async | POST | `/v1/async/orders` |
| Preview order async | POST | `/v1/async/orders/preview` |
| Delete order async | DELETE | `/v1/async/orders/{orderNumber}` |
| Check async job status | GET | `/v1/async-jobs/{jobId}` |

### Sync vs Async Limits
- Synchronous: up to 50 subscriptions, 50 order actions per call
- Asynchronous: up to 300 subscriptions, 300 order actions per call

---

## 18. Support & Debugging

When contacting Zuora support, provide:
- Error code
- `Zuora-Request-Id` response header
- `processId` from error body
- Timestamp
- Operation attempted
