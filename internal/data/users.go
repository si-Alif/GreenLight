package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.si-Alif.net/internal/validator"
)

// email is the unique identifier for user
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// AnonymousUser pointer holds the address of a User instance . This same address will be passed as the user address to authentication middleware for user instance of unauthenticated users , so when we do IsAnonymous() check and compare user in the request context window with this pointer address , we get true .
var AnonymousUser = &User{}

// hid password and version field from any output
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plainText *string
	hash      []byte
}

// if the user's User instance in the request context is the same address as AnonymousUser ; we get true , else false
func (u *User) IsAnonymous() bool {
	return AnonymousUser == u
}

// hash given pass
func (p *password) Set(plainTextPass string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(plainTextPass), 12)

	if err != nil {
		return err
	}

	p.plainText = &plainTextPass
	p.hash = hashedPass

	return nil

}

// compare with provided password
func (p *password) Matches(plainTextPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPass))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}

	}

	return true, nil

}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

}

func ValidatePasswordPlainText(v *validator.Validator, plainTextPass string) {
	v.Check(plainTextPass != "", "password", "must be provided")
	v.Check(len(plainTextPass) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plainTextPass) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plainText != nil {
		ValidatePasswordPlainText(v, *user.Password.plainText)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

// define a user model that wraps around a sql.DB connection pool to interact with users table in database via methods defined on it
type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name , email , password_hash , activated)
	VALUES ($1 , $2 , $3 , $4)
	RETURNING id , created_at , version
	`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	// define a 3s context window
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id , created_at , name , email , password_hash , activated , version
	FROM users
	WHERE email = $1
	`

	var user User

	// define a 3s context window
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m *UserModel) GetByID(id int64) (*User, error) {
	query := `
	SELECT id , created_at , name , email , password_hash , activated , version
	FROM users
	WHERE id = $1
	`

	var user User

	// define a 3s context window
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m *UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1 , email = $2 , password_hash = $3 , activated = $4 , version = version + 1
	WHERE id = $5 AND version  = $6
	RETURNING version
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	// define a 3s context window
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Version,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflicts
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) GetUserViaToken(tokenScope, tokenPlainText string) (*User, error) {
	hashedToken := sha256.Sum256([]byte(tokenPlainText))

	query := `SELECT users.id , users.created_at , users.name , users.email , users.password_hash , users.activated , users.version
						FROM users
						INNER JOIN tokens
						ON users.id = tokens.user_id
						WHERE tokens.hash = $1
						AND scope = $2
						AND tokens.expiry > $3`

	args := []any{hashedToken[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}

// check if the user is Ano
