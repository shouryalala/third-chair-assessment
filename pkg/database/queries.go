package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// GetUserByUsername retrieves a user by username
func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, full_name, biography, is_verified,
		       is_business_account, is_professional_account, is_private,
		       category_name, followers, following, posts, scraped_at,
		       created_at, updated_at
		FROM instagram_users
		WHERE username = $1
	`

	var user User
	err := DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.FullName, &user.Biography,
		&user.IsVerified, &user.IsBusinessAccount, &user.IsProfessionalAccount,
		&user.IsPrivate, &user.CategoryName, &user.Followers, &user.Following,
		&user.Posts, &user.ScrapedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, username, full_name, biography, is_verified,
		       is_business_account, is_professional_account, is_private,
		       category_name, followers, following, posts, scraped_at,
		       created_at, updated_at
		FROM instagram_users
		WHERE id = $1
	`

	var user User
	err := DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.FullName, &user.Biography,
		&user.IsVerified, &user.IsBusinessAccount, &user.IsProfessionalAccount,
		&user.IsPrivate, &user.CategoryName, &user.Followers, &user.Following,
		&user.Posts, &user.ScrapedAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpsertUser inserts or updates a user
func UpsertUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO instagram_users (
			id, username, full_name, biography, is_verified,
			is_business_account, is_professional_account, is_private,
			category_name, followers, following, posts, scraped_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			full_name = EXCLUDED.full_name,
			biography = EXCLUDED.biography,
			is_verified = EXCLUDED.is_verified,
			is_business_account = EXCLUDED.is_business_account,
			is_professional_account = EXCLUDED.is_professional_account,
			is_private = EXCLUDED.is_private,
			category_name = EXCLUDED.category_name,
			followers = EXCLUDED.followers,
			following = EXCLUDED.following,
			posts = EXCLUDED.posts,
			scraped_at = EXCLUDED.scraped_at,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := DB.ExecContext(ctx, query,
		user.ID, user.Username, user.FullName, user.Biography,
		user.IsVerified, user.IsBusinessAccount, user.IsProfessionalAccount,
		user.IsPrivate, user.CategoryName, user.Followers, user.Following,
		user.Posts, user.ScrapedAt,
	)

	if err != nil {
		log.Error().Err(err).Str("username", user.Username).Msg("failed to upsert user")
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	log.Debug().Str("username", user.Username).Msg("upserted user")
	return nil
}

// GetUserStats retrieves detailed user statistics using complex query
// This is adapted from the actual Hendrix complex query
func GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	// Complex query adapted from Hendrix instagram_user_get.go
	query := `
		SELECT
			row_to_json(u.*) AS user,
			(SELECT coalesce(json_agg(t.*), '[]'::json)
			 FROM (
				SELECT COUNT(p.id) as count,
					   UNNEST(a.tagged_user_usernames) as username,
					   SUM(COALESCE(p.like_count, 0)) as total_likes,
					   SUM(COALESCE(p.comment_count, 0)) as total_comments,
					   SUM(COALESCE(p.play_count, 0)) as total_plays
				FROM instagram_posts p
					LEFT JOIN instagram_assets a ON a.post_id = p.id
				WHERE p.user_id = u.id AND a.tagged_user_usernames IS NOT NULL
				GROUP BY UNNEST(a.tagged_user_usernames)
				ORDER BY count DESC
				LIMIT 10
			 ) t) as tagged_usernames,
			(SELECT coalesce(json_agg(t.*), '[]'::json)
			 FROM (
				SELECT COUNT(p.id) as collaboration_count,
					   p.username as collaborator,
					   SUM(COALESCE(p.like_count, 0)) as total_likes,
					   SUM(COALESCE(p.comment_count, 0)) as total_comments
				FROM instagram_posts p
				WHERE p.user_id = u.id
				GROUP BY p.username
				ORDER BY collaboration_count DESC
				LIMIT 10
			 ) t) as coauthored_usernames,
			(SELECT COUNT(DISTINCT p.id)
			 FROM instagram_posts p
			 WHERE p.user_id = u.id) AS total_posted_count,
			(SELECT COUNT(DISTINCT a.id)
			 FROM instagram_assets a
			 JOIN instagram_posts p ON a.post_id = p.id
			 WHERE a.tagged_user_usernames @> ARRAY[u.username]::text[]) as total_tagged_in_count,
			(SELECT COUNT(DISTINCT p.id)
			 FROM instagram_posts p
			 WHERE p.user_id = u.id AND p.is_ad = false) as total_coauthored_count,
			-- Calculate engagement rate
			CASE
				WHEN u.followers > 0 THEN
					(SELECT AVG(COALESCE(p.like_count, 0) + COALESCE(p.comment_count, 0))
					 FROM instagram_posts p
					 WHERE p.user_id = u.id AND p.posted_at > NOW() - INTERVAL '30 days') / u.followers * 100
				ELSE 0
			END as engagement_rate,
			-- Calculate average posts per week
			(SELECT COUNT(*)::float /
				CASE
					WHEN EXTRACT(days FROM (NOW() - MIN(p.posted_at))) > 0
					THEN EXTRACT(days FROM (NOW() - MIN(p.posted_at))) / 7.0
					ELSE 1
				END
			 FROM instagram_posts p
			 WHERE p.user_id = u.id) as average_posts_per_week
		FROM instagram_users u
		WHERE u.id = $1
	`

	var stats UserStats
	err := DB.QueryRowContext(ctx, query, userID).Scan(
		&stats.User,
		&stats.TaggedUsernames,
		&stats.CoauthoredUsernames,
		&stats.TotalPostedCount,
		&stats.TotalTaggedInCount,
		&stats.TotalCoauthoredCount,
		&stats.EngagementRate,
		&stats.AveragePostsPerWeek,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get user stats")
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

// BatchUpsertUsers efficiently inserts/updates multiple users
func BatchUpsertUsers(ctx context.Context, users []*User) error {
	if len(users) == 0 {
		return nil
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO instagram_users (
			id, username, full_name, biography, is_verified,
			is_business_account, is_professional_account, is_private,
			category_name, followers, following, posts, scraped_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			full_name = EXCLUDED.full_name,
			biography = EXCLUDED.biography,
			is_verified = EXCLUDED.is_verified,
			is_business_account = EXCLUDED.is_business_account,
			is_professional_account = EXCLUDED.is_professional_account,
			is_private = EXCLUDED.is_private,
			category_name = EXCLUDED.category_name,
			followers = EXCLUDED.followers,
			following = EXCLUDED.following,
			posts = EXCLUDED.posts,
			scraped_at = EXCLUDED.scraped_at,
			updated_at = CURRENT_TIMESTAMP
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, user := range users {
		_, err = stmt.ExecContext(ctx,
			user.ID, user.Username, user.FullName, user.Biography,
			user.IsVerified, user.IsBusinessAccount, user.IsProfessionalAccount,
			user.IsPrivate, user.CategoryName, user.Followers, user.Following,
			user.Posts, user.ScrapedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to execute statement for user %s: %w", user.Username, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Int("count", len(users)).Msg("batch upserted users")
	return nil
}

// GetProcessingJobStatus gets the status of a processing job
func GetProcessingJobStatus(ctx context.Context, jobID string) (*ProcessingJob, error) {
	query := `
		SELECT id, status, total_users, processed_users, successful_users,
		       failed_users, max_concurrency, started_at, completed_at,
		       errors, created_at, updated_at
		FROM processing_jobs
		WHERE id = $1
	`

	var job ProcessingJob
	var errorsJSON []byte

	err := DB.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.Status, &job.TotalUsers, &job.ProcessedUsers,
		&job.SuccessfulUsers, &job.FailedUsers, &job.MaxConcurrency,
		&job.StartedAt, &job.CompletedAt, &errorsJSON,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse errors JSON
	if len(errorsJSON) > 0 {
		if err := json.Unmarshal(errorsJSON, &job.Errors); err != nil {
			log.Warn().Err(err).Msg("failed to parse job errors JSON")
			job.Errors = make(map[string]string)
		}
	} else {
		job.Errors = make(map[string]string)
	}

	return &job, nil
}