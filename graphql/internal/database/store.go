package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/abbeyhrt/keep-up/graphql/internal/models"
	"github.com/abbeyhrt/keep-up/graphql/internal/session"
	log "github.com/sirupsen/logrus"
)

// NewStoreFromClient creates a new SQLstore
func NewStoreFromClient(db *sql.DB) *SQLStore {
	return &SQLStore{db}
}

//SQLStore holds all of the functions that we use on the store
type SQLStore struct {
	db *sql.DB
}

// CreateSession saves a user's session in the DB to make viewer info more accessible.
func (s *SQLStore) CreateSession(ctx context.Context, userID string) (models.Session, error) {
	session := models.Session{}
	err := s.db.QueryRowContext(
		ctx,
		sqlCreateSession,
		userID,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
	)

	return session, err

}

// GetSessionByID retrieves a user's current session by its ID
func (s *SQLStore) GetSessionByID(ctx context.Context, id string) (models.Session, error) {
	session := models.Session{}
	err := s.db.QueryRowContext(ctx, sqlGetSessionByID, id).Scan(
		&session.ID,
		&session.UserID,
	)
	return session, err
}

// GetUserByID finds a user by their ID
func (s *SQLStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	u := models.User{}
	err := s.db.QueryRowContext(ctx, sqlGetUserByID, id).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.HomeID,
		&u.Email,
		&u.AvatarURL,
		&u.Provider,
		&u.ProviderID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	return u, err
}

// GetOrCreateUser finds a user upon logging in or creates the user if that user doesn't exist
func (s *SQLStore) GetOrCreateUser(
	ctx context.Context,
	u *models.User,
) error {
	err := s.db.QueryRowContext(ctx, sqlGetUserByProvider, u.Provider, u.ProviderID).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.HomeID,
		&u.Email,
		&u.AvatarURL,
		&u.Provider,
		&u.ProviderID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		user, err := s.CreateUser(ctx, *u)
		if err != nil {
			return err
		}

		*u = user
		return nil
	}

	return err
}

// CreateUser creates a user
func (s *SQLStore) CreateUser(
	ctx context.Context,
	user models.User,
) (models.User, error) {
	err := s.db.QueryRowContext(
		ctx,
		sqlCreateUser,
		user.FirstName,
		user.LastName,
		user.Email,
		user.AvatarURL,
		user.Provider,
		user.ProviderID,
	).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, fmt.Errorf("error creating user: %v", err)
	}

	return user, nil
}

func (s *SQLStore) GetUsersByName(ctx context.Context, name string) ([]models.User, error) {

	rows, err := s.db.QueryContext(ctx, sqlGetUsersByName, name)
	if err != nil {
		log.Errorf("This is the %s: ", err)
		return nil, err
	}
	defer rows.Close()
	var users []models.User

	for rows.Next() {
		u := models.User{}
		err := rows.Scan(
			&u.ID,
			&u.FirstName,
			&u.LastName,
			&u.HomeID,
			&u.Email,
			&u.AvatarURL,
		)
		if err != nil {
			log.Errorf("This is the %s: ", err)
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

// UpdateUser updates any value in any table based on any identifier.
func (s *SQLStore) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
	user.UpdatedAt = time.Now()
	err := s.db.QueryRowContext(
		ctx,
		sqlUpdateUser,
		user.FirstName,
		user.LastName,
		user.HomeID,
		user.Email,
		user.AvatarURL,
		user.UpdatedAt,
		user.ID,
	).Scan(
		&user.FirstName,
		&user.LastName,
		&user.HomeID,
		&user.Email,
		&user.AvatarURL,
		&user.UpdatedAt,
		&user.ID,
	)

	if err != nil {
		log.Errorf("This is the %s: ", err)
		return user, err
	}
	return user, nil
}

func (s *SQLStore) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, sqlDeleteUser, userID)
	if err != nil {
		log.Errorf("Error deleting user, %s", err)
		return err
	}

	return nil
}

// --------------------- HOME METHODS --------------------- //

// CreateHome function that will be used in the handlers
func (s *SQLStore) CreateHome(ctx context.Context, home models.Home, userID string) (models.Home, error) {
	err := s.db.QueryRowContext(
		ctx,
		sqlCreateHome,
		home.Name,
		home.Description,
		home.AvatarURL,
	).Scan(
		&home.ID,
		&home.Name,
		&home.Description,
		&home.AvatarURL,
		&home.CreatedAt,
		&home.UpdatedAt,
	)
	if err != nil {
		log.Errorf("error creating home: %s", err)
		return home, err
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		log.Errorf("This is the %s: ", err)
		return home, err
	}

	user.HomeID = &home.ID

	u, err := s.UpdateUser(ctx, user)
	if err != nil {
		log.Errorf("This is the %s and the user %v: ", err, u)
		return home, err
	}

	return home, nil
}

//GetHomeByID used in handlers package
func (s *SQLStore) GetHomeByID(ctx context.Context, homeID *string) (models.Home, error) {
	h := models.Home{}
	err := s.db.QueryRowContext(ctx, sqlGetHomeByID, homeID).Scan(
		&h.ID,
		&h.Name,
		&h.Description,
	)

	return h, err
}

func (s *SQLStore) UpdateHome(ctx context.Context, home models.Home) (models.Home, error) {
	home.UpdatedAt = time.Now()
	err := s.db.QueryRowContext(
		ctx,
		sqlUpdateHome,
		home.Name,
		home.Description,
		home.AvatarURL,
		home.UpdatedAt,
		home.ID,
	).Scan(
		&home.Name,
		&home.Description,
		&home.AvatarURL,
		&home.UpdatedAt,
		&home.ID,
	)

	if err != nil {
		log.Errorf("This is the %s: ", err)
		return home, err
	}
	return home, nil
}

func (s *SQLStore) DeleteHome(ctx context.Context, ID string) error {
	t, ok := session.FromContext(ctx)
	if !ok {
		return nil
	}

	u, err := s.GetUserByID(ctx, t.User.ID)
	if err != nil {
		log.Errorf("Error finding user %s", err)
		return err
	}

	u.HomeID = nil

	_, err = s.UpdateUser(ctx, u)
	if err != nil {
		log.Errorf("Error updating user %s", err)
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlDeleteHome, ID)
	if err != nil {
		log.Errorf("Error deleting user, %s", err)
		return err
	}

	return nil
}

// --------------------- TASKS METHODS ---------------------- //

// CreateTask creates a task and associates it with the user who made it
func (s *SQLStore) CreateTask(ctx context.Context, task models.Task, userID string) (models.Task, error) {
	err := s.db.QueryRowContext(
		ctx,
		sqlCreateTask,
		userID,
		task.Title,
		task.Description,
	).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		log.Errorf("This is the %s: ", err)
		return task, err
	}

	return task, nil

}

//UpdateTask updates any value in any table based on any identifier.
func (s *SQLStore) UpdateTask(ctx context.Context, task models.Task) (models.Task, error) {
	task.UpdatedAt = time.Now()
	err := s.db.QueryRowContext(
		ctx,
		sqlUpdateTask,
		task.UserID,
		task.Title,
		task.Description,
		task.UpdatedAt,
		task.ID,
	).Scan(
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.UpdatedAt,
		&task.ID,
	)

	if err != nil {
		log.Errorf("This is the %s: ", err)
		return task, err
	}
	return task, nil
}

// GetTasksByUserID returns all of a user's tasks
func (s *SQLStore) GetTasksByUserID(ctx context.Context, userID string) ([]models.Task, error) {
	rows, err := s.db.QueryContext(ctx, sqlGetTasksByUserID, userID)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		t := models.Task{}

		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			log.Errorf("This is the %s: ", err)
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetTaskByID searched the db for a task by its ID
func (s *SQLStore) GetTaskByID(ctx context.Context, id string) (models.Task, error) {
	t := models.Task{}
	err := s.db.QueryRowContext(
		ctx,
		sqlGetTaskByID,
		id,
	).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
	)
	if err != nil {
		log.Errorf("This is the %s: ", err)
		return t, err
	}

	return t, err
}

func (s *SQLStore) DeleteTask(ctx context.Context, ID string) error {
	_, err := s.db.ExecContext(ctx, sqlDeleteTask, ID)
	if err != nil {
		log.Errorf("Error deleting user, %s", err)
		return err
	}

	return nil
}

const (

	// Session Statements

	sqlCreateSession = `
	INSERT into sessions
	(user_id)
	VALUES ($1)
	RETURNING id, user_id, created_at
	`
	sqlGetSessionByID = `
	SELECT id, user_id
	FROM sessions
	WHERE id = $1`

	// User Statements

	sqlCreateUser = `
	INSERT into users
	(first_name, last_name, email, avatar_url, provider, provider_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, first_name, last_name, email, avatar_url, provider, provider_id, created_at, updated_at
	`

	sqlGetUserByProvider = `
	SELECT id, first_name, last_name, home_id, email, avatar_url, provider, provider_id, created_at, updated_at
	FROM users
	WHERE provider = $1 AND provider_id = $2
	`

	sqlGetUserByID = `
	SELECT id, first_name, last_name, home_id, email, avatar_url, provider, provider_id, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	sqlGetUsersByName = `
	SELECT id, first_name, last_name, home_id, email, avatar_url
	FROM users
	WHERE CONCAT(LOWER(first_name), ' ', LOWER(last_name)) LIKE LOWER('%' || $1 || '%')
	`

	sqlUpdateUser = `
	UPDATE users
	SET first_name = $1,
			last_name = $2,
			home_id = $3,
			email = $4,
			avatar_url = $5,
			updated_at = $6
	WHERE id = $7
	RETURNING first_name, last_name, home_id, email, avatar_url, updated_at, id
	`

	sqlDeleteUser = `
	DELETE FROM users
	WHERE id = $1
	`

	// Home Statements

	sqlCreateHome = `
	INSERT into homes
	(name, description, avatar_url)
	VALUES ($1, $2, $3)
	RETURNING id, name, description, avatar_url, created_at, updated_at
	`

	sqlGetHomeByID = `
	SELECT id, name, description
	FROM homes
	WHERE id = $1
	`

	sqlUpdateHome = `
	UPDATE homes
	SET name = $1,
			description = $2,
			avatar_url = $3,
			updated_at = $4
	WHERE id = $5
	RETURNING name, description, avatar_url, updated_at, id
	`

	sqlDeleteHome = `
	DELETE FROM homes
	WHERE id = $1
	`
	// Task Statements

	sqlCreateTask = `
	INSERT into tasks
	(user_id, title, description)
	VALUES ($1, $2, $3)
	RETURNING id, user_id, title, description, created_at, updated_at
	`

	sqlGetTasksByUserID = `
	SELECT id, user_id, title, description, created_at, updated_at
	FROM tasks
	WHERE user_id = $1
	`

	sqlGetTaskByID = `
	SELECT id, user_id, title, description
	FROM tasks
	WHERE id = $1`

	sqlUpdateTask = `
	UPDATE tasks
	SET user_id = $1,
			title = $2,
			description = $3,
			updated_at = $4
	WHERE id = $5
	RETURNING user_id, title, description, updated_at, id`

	sqlDeleteTask = `
	DELETE FROM tasks
	WHERE id = $1
	`
)
