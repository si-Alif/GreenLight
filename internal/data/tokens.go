package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"greenlight.si-Alif.net/internal/validator"
)

// constant definition for different scopes
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// token info struct
// restructured the Token struct to modify around how it should appear in json body
type Token struct {
	PlainText string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserId    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// function definition to generate Token instance
func generateToken(user_id int64, ttl time.Duration, scope string) *Token { // token is a sensitive info , so rather than moving it by making copies of it in the memory , we prefer to use pointers to transfer it

	token := &Token{
		PlainText: rand.Text(),
		UserId:    user_id,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	// to store in the DB , we need to hash the token using SHA256 hash and then convert it into a slice for ease of work using [:] operator
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token
}

// validations
func ValidPlainTextToken(v *validator.Validator, plainTextToken string) {
	v.Check(plainTextToken != "", "token", "must be provided")
	v.Check(len(plainTextToken) == 26, "token", "invalid token")
}

// TokenModel contains some of the connections from the pool
type TokenModel struct {
	DB *sql.DB
}

// New Generates a token and inserts into the db using generateToken and Insert method and return Token struct instance and error .
// Usage : generates the plaintext token which we'll send to user via activation link and inserts the hashed token in the db at the same time
func (tm *TokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token := generateToken(userId, ttl, scope)

	err := tm.Insert(token)

	return token, err
}

func (tm *TokenModel) Insert(token *Token) error {
	query := `INSERT INTO tokens (hash , user_id , expiry , scope) VALUES ($1 , $2 , $3 , $4)`

	args := []any{
		token.Hash,
		token.UserId,
		token.Expiry,
		token.Scope,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tm.DB.ExecContext(ctx, query, args...)

	// if any error actually occurs , it gets returned or if no error then nil returns , which doesn't affect anything
	return err
}

// delete all the tokens for a specific user of a specific scope
func (tm *TokenModel) DeleteAllTokensForASpecificScopeOfUser(scope string, userID int64) error {
	query := `DELETE FROM tokens WHERE scope=$1 AND user_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := tm.DB.ExecContext(ctx, query, scope, userID)
	return err
}
