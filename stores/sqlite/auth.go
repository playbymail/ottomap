// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite/sqlc"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"strconv"
	"strings"
	"time"
)

// This file should implement the store for the authentication domain.
// Maybe someday I will understand how to do this.

func (db *DB) AuthenticateUser(email, plainTextPassword string) (domains.ID, error) {
	if strings.TrimSpace(email) != email {
		return 0, domains.ErrInvalidEmail
	}

	row, err := db.q.GetUserByEmail(db.ctx, email)
	if err != nil {
		return 0, err
	} else if !CheckPassword(plainTextPassword, row.HashedPassword) {
		return 0, domains.ErrUnauthorized
	}

	// update the last login time, ignoring any errors
	_ = db.q.UpdateUserLastLogin(db.ctx, row.UserID)

	return domains.ID(row.UserID), nil
}

func (db *DB) CreateOperator(plainTextSecret string) error {
	hashedPassword, err := HashPassword(plainTextSecret)
	if err != nil {
		return err
	}
	return db.q.CreateOperator(db.ctx, hashedPassword)
}

func (db *DB) CreateUser(email, plainTextSecret, clan, role string, timezone *time.Location) (domains.ID, error) {
	if strings.TrimSpace(email) != email {
		return 0, domains.ErrInvalidEmail
	} else if clanNo, err := strconv.Atoi(clan); err != nil || clanNo < 1 || clanNo > 999 {
		return 0, domains.ErrInvalidClan
	}

	// note: we let LastLogin be the zero-value for time.Time, which means never logged in.
	parms := sqlc.CreateUserParams{
		Email:    strings.ToLower(email),
		IsActive: 1,
		Clan:     clan,
	}

	// adapt values and provide defaults as needed

	// hash the password. can fail if the password is too long.
	if hashedPassword, err := HashPassword(plainTextSecret); err != nil {
		return 0, err
	} else {
		parms.HashedPassword = hashedPassword
	}
	if role == "admin" || role == "operator" {
		parms.Role = role
	} else {
		parms.Role = "user"
	}
	if timezone == nil {
		parms.Timezone = "UTC"
	} else {
		parms.Timezone = timezone.String()
	}

	id, err := db.q.CreateUser(db.ctx, parms)
	return domains.ID(id), err
}

func (db *DB) GetUser(userID domains.ID) (*domains.User, error) {
	row, err := db.q.GetUser(db.ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	// convert row.Timezone to a time.Location
	loc, err := time.LoadLocation(row.Timezone)
	if err != nil {
		return nil, err
	}
	isActive := row.IsActive == 1
	roleMap := map[string]bool{"active": isActive, "authenticated": false}
	if isActive {
		switch row.Role {
		case "admin":
			roleMap["admin"] = true
		case "operator":
			roleMap["operator"] = true
		case "user":
			roleMap["user"] = true
		}
	}
	return &domains.User{
		ID:       userID,
		Email:    row.Email,
		Timezone: loc,
		IsActive: isActive,
		Clan:     row.Clan,
		Roles:    roleMap,
	}, nil
}

//type User struct {
//	ID       int
//	Username string
//	Timezone string
//}
//
//
//func handler(w http.ResponseWriter, r *http.Request) {
//	// Assuming you have a function to get the user ID from the request
//	userID := getUserIDFromRequest(r)
//
//	db, err := sql.Open("sqlite3", "./your-database.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	user, err := getUser(db, userID)
//	if err != nil {
//		http.Error(w, "User not found", http.StatusNotFound)
//		return
//	}
//
//	// Load the time.Location using the timezone value
//	loc, err := time.LoadLocation(user.Timezone)
//	if err != nil {
//		http.Error(w, "Invalid timezone", http.StatusInternalServerError)
//		return
//	}
//
//	// Use the time.Location (loc) as needed
//	fmt.Fprintf(w, "User: %s, Timezone: %s, Location: %v", user.Username, user.Timezone, loc)
//}
//
//func main() {
//	http.HandleFunc("/", handler)
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}
//
//// Dummy function to simulate extracting user ID from the request
//func getUserIDFromRequest(r *http.Request) int {
//	return 1 // Replace with actual logic
//}

// simple password functions inspired by https://www.gregorygaines.com/blog/how-to-properly-hash-and-salt-passwords-in-golang-bcrypt/

// CheckPassword returns true if the plain text password matches the hashed password.
func CheckPassword(plainTextPassword, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextPassword)) == nil
}

// HashPassword uses the cheapest bcrypt cost to hash the password because we are not going to use
// it for anything other than authentication in non-production environments.
func HashPassword(plainTextPassword string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordBytes), err
}
