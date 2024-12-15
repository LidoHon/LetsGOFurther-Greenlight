package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/validator"
)


const (
	ScopeActivation ="activation"
)

type Token struct{
	PlainText 	string
	Hash 		[]byte
	UserID 		int64
	Expiry 		time.Time
	Scope 		string
}


func generateToken(userID int64, ttl time.Duration, scope string)(*Token, error){

token := &Token{
	UserID: userID,
	Expiry: time.Now().Add(ttl),
	Scope: scope,
}
// Initialize a zero-valued byte slice with a length of 16 bytes.
randomBytes := make([]byte, 16)

_, err := rand.Read(randomBytes)
if err !=nil {
	return nil, err
}

token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)


// here  sha256.Sum256() function returns an *array* of length 32, so to make it easier to
// work with we convert it to a slice using the [:] operator before storing it
hash := sha256.Sum256([]byte(token.PlainText))
token.Hash = hash[:]
return token, nil

}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string){
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "tokens", "must be 26 bytes long")
}

type TokenModel struct{
	DB *sql.DB
}


func(m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error){
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func(m TokenModel)Insert(token *Token) error{
	query := `INSERT INTO tokens (hash, user_id, expiry,scope) VALUES ($1,$2,$3,$4)`

	args :=[]interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3 *time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err !=nil {
		return err
	}
	return nil
}

func(m TokenModel) DeleteAllForUser(scope string, userID int64) error{
	query :=`DELETE FROM tokens WHERE scope = $1 AND user_id =$2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_,err := m.DB.ExecContext(ctx, query, scope, userID)
	if err != nil{
		return err
	}
	return nil
}