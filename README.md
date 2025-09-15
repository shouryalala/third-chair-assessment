# Instagram User Processor - Batch Processing Challenge

> **Interview Assignment**: Implement parallel batch processing for Instagram user data collection

## 🎯 Assignment Overview

This is a self-contained Go application that processes Instagram user data using RocketAPI. The application currently supports single user processing, and **your task is to implement parallel batch processing capabilities**.

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
- Go 1.23+
- Docker & Docker Compose
- RocketAPI key from https://rocketapi.io

### Setup
1. **Clone and setup the environment:**
   ```bash
   ./scripts/setup.sh
   ```

2. **Add your RocketAPI key to `.env`:**
   ```bash
   # Edit .env file
   ROCKETAPI_KEY=your_actual_key_here
   ```

3. **Start the application:**
   ```bash
   go run cmd/server/main.go
   ```

4. **Verify setup:**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/api/v1/instagram/user/musiclover2024
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
└── docker-compose.yml   # Local development setup
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

**Database Admin:** http://localhost:8081 (adminer)
- Server: `postgres`
- Username: `user`
- Password: `password`
- Database: `instagram_processor`

**Health Check:** http://localhost:8080/health

**Sample Data:** Pre-populated with 15 test users and posts

## 🐛 Common Issues

**Database Connection:**
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View logs
docker-compose logs postgres
```

**RocketAPI Issues:**
```bash
# Check your API key in .env
grep ROCKETAPI_KEY .env

# Monitor API calls
tail -f logs/app.log | grep "RocketAPI"
```

**Port Conflicts:**
```bash
# Check what's using port 8080
lsof -i :8080

# Use different port
export PORT=8081
go run cmd/server/main.go
```

## 📝 Implementation Notes

### Worker Pool Pattern
```go
// Example pattern for your implementation
pool := queue.NewRateLimitedWorkerPool(req.MaxConcurrency, rateLimiter)

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