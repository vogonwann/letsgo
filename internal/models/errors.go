package models
import (
  "errors"
)

var ( 
  ErrNoRecord = errors.New("models: no matching record found")

  // Error for wrong user credentials
  ErrInvalidCredentials = errors.New("models: invalid credentials")

  // Error for duplicate user email
  ErrDuplicateEmail = errors.New("models: duplicate email")
)
