# External Activity API Documentation

## Endpoints

### 1. Community Activity (All Messages)
```
GET /api/external/community/{community_id}/activity
```
Returns message count grouped by time intervals for all messages in the community.

### 2. Community FUD Activity (FUD Users Only)
```
GET /api/external/community/{community_id}/fud-activity
```
Returns message count grouped by time intervals for messages from FUD users only (where `is_fud_user = 1`).

---

## Request Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `community_id` | string | Yes | Twitter community ID | `1914102634241577036` |
| `timestamp_from` | int64 | No* | Start of period in milliseconds | `1704067200000` |
| `timestamp_to` | int64 | No* | End of period in milliseconds | `1706745600000` |
| `period` | string | No | Grouping interval | `day` |

\* If not specified, defaults to last 30 days

---

## Period Values

| Value | Description | Interval |
|-------|-------------|----------|
| `hour` | Hourly | 1 hour |
| `2hour` | Every 2 hours | 2 hours |
| `4hour` | Every 4 hours | 4 hours |
| `6hour` | Every 6 hours | 6 hours |
| `day` | Daily (default) | 1 day |
| `week` | Weekly | 1 week |
| `month` | Monthly | 1 month |

---

## Response Format

### Success Response

```json
{
  "status": "ok",
  "data": [
    {
      "timestamp": 1704067200000,
      "message_count": 42
    },
    {
      "timestamp": 1704153600000,
      "message_count": 38
    },
    {
      "timestamp": 1704240000000,
      "message_count": 51
    }
  ]
}
```

**Fields:**
- `status` - Response status (`"ok"` or `"error"`)
- `data` - Array of activity data points
  - `timestamp` - Unix timestamp in milliseconds
  - `message_count` - Number of messages in this interval

### Error Response

```json
{
  "status": "error",
  "message": "Invalid timestamp_from parameter",
  "error": "strconv.ParseInt: parsing \"invalid\": invalid syntax"
}
```

---

## Examples

### Example 1: Last 30 days by day (default)
```bash
GET /api/external/community/1914102634241577036/activity
```

**Response:**
```json
{
  "status": "ok",
  "data": [
    {"timestamp": 1704067200000, "message_count": 42},
    {"timestamp": 1704153600000, "message_count": 38},
    {"timestamp": 1704240000000, "message_count": 51}
  ]
}
```

---

### Example 2: Specific period by hour
```bash
GET /api/external/community/1914102634241577036/activity?timestamp_from=1704067200000&timestamp_to=1704153600000&period=hour
```

**Response:**
```json
{
  "status": "ok",
  "data": [
    {"timestamp": 1704067200000, "message_count": 5},
    {"timestamp": 1704070800000, "message_count": 3},
    {"timestamp": 1704074400000, "message_count": 8},
    {"timestamp": 1704078000000, "message_count": 2}
  ]
}
```

---

### Example 3: FUD activity for last month by day
```bash
GET /api/external/community/1914102634241577036/fud-activity?timestamp_from=1704067200000&timestamp_to=1706745600000&period=day
```

**Response:**
```json
{
  "status": "ok",
  "data": [
    {"timestamp": 1704067200000, "message_count": 5},
    {"timestamp": 1704153600000, "message_count": 3},
    {"timestamp": 1704240000000, "message_count": 8}
  ]
}
```

---

### Example 4: Activity by 4-hour intervals
```bash
GET /api/external/community/1914102634241577036/activity?timestamp_from=1704067200000&timestamp_to=1706745600000&period=4hour
```

**Response:**
```json
{
  "status": "ok",
  "data": [
    {"timestamp": 1704067200000, "message_count": 15},
    {"timestamp": 1704081600000, "message_count": 12},
    {"timestamp": 1704096000000, "message_count": 18}
  ]
}
```

---

### Example 5: Weekly statistics for 3 months
```bash
GET /api/external/community/1914102634241577036/activity?timestamp_from=1704067200000&timestamp_to=1712102400000&period=week
```

**Response:**
```json
{
  "status": "ok",
  "data": [
    {"timestamp": 1703980800000, "message_count": 245},
    {"timestamp": 1704585600000, "message_count": 198},
    {"timestamp": 1705190400000, "message_count": 312}
  ]
}
```

---

## Timestamp Examples

### JavaScript/TypeScript
```javascript
// Current date
const now = Date.now(); // 1735689600000

// Month ago
const monthAgo = Date.now() - (30 * 24 * 60 * 60 * 1000); // 1733011200000

// Week ago
const weekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000); // 1735084800000

// Yesterday
const yesterday = Date.now() - (24 * 60 * 60 * 1000); // 1735603200000

// Specific date: January 1, 2024 00:00:00
const specificDate = new Date('2024-01-01T00:00:00Z').getTime(); // 1704067200000
```

### Python
```python
import time
from datetime import datetime, timedelta

# Current date
now = int(time.time() * 1000)  # 1735689600000

# Month ago
month_ago = int((datetime.now() - timedelta(days=30)).timestamp() * 1000)

# Week ago
week_ago = int((datetime.now() - timedelta(days=7)).timestamp() * 1000)

# Specific date: January 1, 2024 00:00:00
specific_date = int(datetime(2024, 1, 1, 0, 0, 0).timestamp() * 1000)  # 1704067200000
```

### Go
```go
import "time"

// Current date
now := time.Now().UnixMilli() // 1735689600000

// Month ago
monthAgo := time.Now().AddDate(0, -1, 0).UnixMilli()

// Week ago
weekAgo := time.Now().AddDate(0, 0, -7).UnixMilli()

// Specific date: January 1, 2024 00:00:00
specificDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli() // 1704067200000
```

---

## Full URL Examples

### Development (localhost)
```
# All messages for last month by day
http://localhost:3333/grufender/api/external/community/1914102634241577036/activity?period=day

# FUD activity for last 7 days by hour
http://localhost:3333/grufender/api/external/community/1914102634241577036/fud-activity?timestamp_from=1735084800000&timestamp_to=1735689600000&period=hour

# Activity for December 2024 by 4-hour intervals
http://localhost:3333/grufender/api/external/community/1914102634241577036/activity?timestamp_from=1701388800000&timestamp_to=1704067200000&period=4hour
```

### Production
```
# All messages for last month by day
https://your-domain.com/grufender/api/external/community/1914102634241577036/activity?period=day

# FUD activity for last 7 days by hour
https://your-domain.com/grufender/api/external/community/1914102634241577036/fud-activity?timestamp_from=1735084800000&timestamp_to=1735689600000&period=hour
```

---

## Testing

Run the test suite to verify the endpoints:

```bash
# Run activity API tests
go test -v -run TestExternalCommunityActivityEndpoints

# With custom community ID
TEST_COMMUNITY_ID=1914102634241577036 go test -v -run TestExternalCommunityActivityEndpoints
```

---

## Notes

- All timestamps are in **milliseconds** (JavaScript `Date.now()` format)
- Default period is `day` if not specified
- Default date range is last 30 days if not specified
- FUD activity endpoint filters messages from users with `is_fud_user = 1` in `user_reports` table
- Regular activity endpoint includes all messages with `source_type='community'`
- Response is always sorted by timestamp in ascending order
