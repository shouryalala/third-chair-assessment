# Instagram User Processor - Batch Processing Challenge

> **Interview Assignment**: Implement parallel batch processing for Instagram user data collection

## 🎯 Assignment Overview

This is a Go application that processes Instagram user data using RocketAPI. The application currently supports single user processing, and **your task is to implement parallel batch processing capabilities**.

### Current Functionality
- ✅ Single user Instagram data fetching
- ✅ Database storage with PostgreSQL
- ✅ Rate limiting (10 req/sec)
- ✅ Retry logic with exponential backoff
- ✅ Complex database queries and analytics

### Your Challenge
**Implement the `BatchProcessUsersHandler` to process multiple Instagram users in parallel while respecting API rate limits.**

## 🚀 Quick Start

### Prerequisites
- **Go 1.20+** (or any recent version)
- **PostgreSQL** (any recent version 12+)
- **RocketAPI key** from https://rocketapi.io

### Setup Instructions

1. **Clone the repository and install dependencies:**
   ```bash
   git clone <repository-url>
   cd instagram-user-processor
   go mod download
   ```

2. **Set up PostgreSQL database:**
   ```bash
   # Create the database
   createdb instagram_processor

   # Initialize the schema
   psql -d instagram_processor -f init.sql

   # (Optional) Load sample data
   psql -d instagram_processor -f scripts/test_data.sql
   ```

3. **Configure environment variables:**
   ```bash
   # Copy the example configuration
   cp .env.example .env

   # Edit .env and update:
   # 1. ROCKETAPI_KEY - Get from https://rocketapi.io
   # 2. DATABASE_URL - Your PostgreSQL connection string
   ```

4. **Run the application:**
   ```bash
   go run cmd/server/main.go
   ```

5. **Verify setup:**
   ```bash
   # Check health endpoint
   curl http://localhost:8080/health

   # Test single user endpoint
   curl http://localhost:8080/api/v1/instagram/user/musiclover2024
   ```

### Alternative: Use the setup script
```bash
# Run the automated setup
./scripts/setup.sh
```

## 📋 API Endpoints

### Single User Processing (✅ Implemented)
```http
GET /api/v1/instagram/user/{username}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/instagram/user/musiclover2024
```

### Batch Processing (🚧 Your Task)
```http
POST /api/v1/instagram/users/batch
Content-Type: application/json

{
  "usernames": ["user1", "user2", "user3"],
  "max_concurrency": 3
}
```

**Expected Response:**
```json
{
  "job_id": "uuid-here",
  "status": "running",
  "total_users": 3,
  "processed_users": 0,
  "successful_users": 0,
  "failed_users": 0
}
```

## 🎯 Implementation Requirements

### Core Challenge: `BatchProcessUsersHandler`
**Location:** `pkg/api/instagram/handlers.go`

You need to implement parallel processing that:

1. **Creates a worker pool** with `req.MaxConcurrency` workers
2. **Distributes usernames** across workers efficiently
3. **Respects rate limits** (10 requests/second globally)
4. **Handles failures gracefully** without stopping other workers
5. **Updates job status** in real-time in the database
6. **Returns job tracking information** immediately

### Key Constraints
- **Rate Limit:** 10 requests/second across ALL workers
- **Concurrency:** Configurable via `max_concurrency` parameter
- **Fault Tolerance:** One failed user shouldn't break the entire batch
- **Database Updates:** Real-time job progress tracking

### Technical Hints
- Use the existing `WorkerPool` in `pkg/queue/worker.go`
- Leverage `RateLimitedWorkerPool` for API constraints
- Update `processing_jobs` table for progress tracking
- Handle context cancellation and timeouts

## 🏗️ Architecture

```
├── cmd/server/           # Application entry point
├── pkg/
│   ├── api/             # HTTP handlers and routes
│   │   └── instagram/   # Instagram-specific endpoints
│   ├── database/        # Database models and queries
│   ├── external/        # RocketAPI integration
│   ├── queue/           # Worker pool system
│   └── utils/           # Utilities and helpers
├── scripts/             # Setup and utility scripts
├── init.sql             # Database schema
└── .env.example         # Configuration template
```

### Key Components

**Database Schema:**
- `instagram_users` - User profile data
- `instagram_posts` - User posts and engagement
- `instagram_assets` - Media assets and tagged users
- `processing_jobs` - Batch job tracking

**Worker Pool System:**
- `WorkerPool` - Generic concurrent worker management
- `RateLimitedWorkerPool` - API rate limit compliance
- Job queuing and result aggregation

**External Integration:**
- RocketAPI client with retry logic
- Exponential backoff for transient failures
- Request/response logging and metrics

## 🧪 Testing Your Implementation

### Manual Testing
```bash
# Test single user (should work)
curl http://localhost:8080/api/v1/instagram/user/musiclover2024

# Test batch processing (your implementation)
curl -X POST http://localhost:8080/api/v1/instagram/users/batch \
  -H "Content-Type: application/json" \
  -d '{
    "usernames": ["musiclover2024", "techguru_sarah", "foodie_adventures"],
    "max_concurrency": 2
  }'
```

### Database Verification
```sql
-- Check job progress
SELECT * FROM processing_jobs ORDER BY created_at DESC LIMIT 5;

-- Check processed users
SELECT username, scraped_at FROM instagram_users
WHERE scraped_at > NOW() - INTERVAL '1 hour';

-- Complex analytics query
SELECT * FROM get_top_users_by_engagement(5);
```

### Load Testing
```bash
# Test with larger batches
curl -X POST http://localhost:8080/api/v1/instagram/users/batch \
  -H "Content-Type: application/json" \
  -d '{
    "usernames": ["user1", "user2", "user3", "user4", "user5", "user6"],
    "max_concurrency": 3
  }'
```

## 📊 Success Criteria

### Functional Requirements
- [ ] Batch endpoint accepts multiple usernames
- [ ] Parallel processing with configurable concurrency
- [ ] Rate limiting respected (≤10 req/sec)
- [ ] Job progress tracked in database
- [ ] Failed users don't stop batch processing
- [ ] Proper error handling and logging

### Performance Requirements
- [ ] Processes 10+ users efficiently
- [ ] Handles concurrent batch requests
- [ ] Memory usage remains stable
- [ ] No goroutine leaks

### Code Quality
- [ ] Clean, readable implementation
- [ ] Proper error handling
- [ ] Appropriate use of channels/goroutines
- [ ] Database transactions where needed

## 🛠️ Development Tools

**PostgreSQL Admin:**
- Use any PostgreSQL client (pgAdmin, TablePlus, DBeaver, etc.)
- Or use psql command line:
  ```bash
  psql -d instagram_processor
  ```

**Health Check:** http://localhost:8080/health

**Sample Data:** Pre-populated with 15 test users and posts (if you loaded test_data.sql)

## 🐛 Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
pg_isready

# Test connection
psql -d instagram_processor -c "SELECT 1"

# Check your connection string in .env
grep DATABASE_URL .env
```

### RocketAPI Issues
```bash
# Verify your API key is set
grep ROCKETAPI_KEY .env

# Monitor application logs
go run cmd/server/main.go 2>&1 | grep -i error
```

### Port Conflicts
```bash
# Check if port 8080 is available
lsof -i :8080

# Use a different port
PORT=8081 go run cmd/server/main.go
```

## 📝 Implementation Notes

### Worker Pool Pattern
```go
// Example pattern for your implementation
pool := queue.NewRateLimitedWorkerPool(opts, 10) // 10 req/sec

for _, username := range req.Usernames {
    pool.Submit(func() error {
        // Process individual user
        return processUser(username)
    })
}

results := pool.Wait() // Collect all results
```

### Database Updates
```go
// Update job progress periodically
job.ProcessedUsers++
if success {
    job.SuccessfulUsers++
} else {
    job.FailedUsers++
    // Add error to job.Errors JSON field
}
db.UpdateProcessingJob(job)
```

### Rate Limiting
```go
// Respect global rate limit across all workers
if err := rateLimiter.Wait(ctx); err != nil {
    return fmt.Errorf("rate limit wait cancelled: %w", err)
}
```

## 🎉 Good Luck!

This assignment tests your ability to:
- Design concurrent systems
- Handle external API constraints
- Manage database state
- Write production-ready Go code

**Time Estimate:** 2-3 hours for a solid implementation

**Questions?** Check the existing code for patterns and examples. The single user implementation shows the expected structure and error handling approach.