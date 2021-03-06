package database

import (
	"context"
	"database/sql"

	"github.com/bocanada/rest-ws/models"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db}, nil
}

func (repo *PostgresRepository) InsertUser(ctx context.Context, user *models.User) error {
	_, err := repo.db.ExecContext(ctx, "INSERT INTO users (id, email, password) VALUES ($1, $2, $3)", user.ID, user.Email, user.Password)
	return err
}

func (repo *PostgresRepository) GetUserById(ctx context.Context, id string) (*models.User, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, email FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user models.User
	for rows.Next() {
		if err = rows.Scan(&user.ID, &user.Email); err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, email, password FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user models.User
	for rows.Next() {
		if err = rows.Scan(&user.ID, &user.Email, &user.Password); err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *PostgresRepository) InsertPost(ctx context.Context, post *models.Post) error {
	_, err := repo.db.ExecContext(ctx, "INSERT INTO posts (id, post_content, user_id) VALUES ($1, $2, $3)",
		post.Id,
		post.PostContent,
		post.UserId)
	return err
}

func (repo *PostgresRepository) GetPostById(ctx context.Context, id string) (*models.Post, error) {
	rows, err := repo.db.QueryContext(ctx, "SELECT id, post_content, user_id, created_at FROM posts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var post models.Post
	for rows.Next() {
		if err = rows.Scan(&post.Id, &post.PostContent, &post.UserId, &post.CreatedAt); err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &post, nil
}

func (repo *PostgresRepository) UpdatePost(ctx context.Context, post *models.Post) error {
	res, err := repo.db.ExecContext(ctx, "UPDATE posts SET post_content = $1 WHERE id = $2 AND user_id = $3",
		post.PostContent,
		post.Id,
		post.UserId)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *PostgresRepository) DeletePost(ctx context.Context, post *models.Post) error {
	res, err := repo.db.ExecContext(ctx, "DELETE FROM posts WHERE id = $1 AND user_id = $2",
		post.Id,
		post.UserId)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *PostgresRepository) ListPosts(ctx context.Context, limit uint64, after string) ([]*models.Post, error) {
	var rows *sql.Rows
	var err error
	if after == "" {
		rows, err = repo.db.QueryContext(ctx,
			"SELECT id, post_content, user_id, created_at FROM posts ORDER BY id ASC LIMIT $1",
			limit)
	} else {
		rows, err = repo.db.QueryContext(ctx,
			"SELECT id, post_content, user_id, created_at FROM posts WHERE id > $2  ORDER BY id ASC LIMIT $1",
			limit,
			after)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err = rows.Scan(&post.Id, &post.PostContent, &post.UserId, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (repo *PostgresRepository) Close() error {
	return repo.db.Close()
}
