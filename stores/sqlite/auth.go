// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"github.com/google/uuid"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite/sqlc"
	"golang.org/x/crypto/bcrypt"
	"log"
	_ "modernc.org/sqlite"
	"strconv"
	"strings"
	"time"
)

// This file should implement the store for the authentication domain.
// Maybe someday I will understand how to do this.

// UpdateAdministrator updates the administrator's password.
// Like all functions, it assumes that the administrator has user_id of 1.
func (db *DB) UpdateAdministrator(plainTextSecret string, isActive bool) error {
	var err error
	var row sqlc.GetUserHashedPasswordRow
	if plainTextSecret == "" {
		row, err = db.q.GetUserHashedPassword(db.ctx, 1)
	} else {
		row.HashedPassword, err = HashPassword(plainTextSecret)
		if err != nil {
			return err
		}
	}
	log.Printf("db: auth: updateAdministrator: password %q: hashed %q\n", plainTextSecret, row.HashedPassword)
	parms := sqlc.UpdateUserPasswordParams{
		UserID:         1,
		HashedPassword: row.HashedPassword,
	}
	if isActive {
		parms.IsActive = 1
	}
	return db.q.UpdateUserPassword(db.ctx, parms)
}

func (db *DB) CreateUser(email, plainTextSecret, clan string, timezone *time.Location) (*domains.User_t, error) {
	if strings.TrimSpace(email) != email {
		return nil, domains.ErrInvalidEmail
	}
	email = strings.ToLower(email)
	if clanNo, err := strconv.Atoi(clan); err != nil || clanNo < 1 || clanNo > 999 {
		return nil, domains.ErrInvalidClan
	}

	// hash the password. can fail if the password is too long.
	hashedPassword, err := HashPassword(plainTextSecret)
	if err != nil {
		return nil, err
	}

	magicLink := uuid.NewString()

	// lookup the timezone. not sure that can fail, but if it does, default to UTC.
	var tz string
	if timezone != nil {
		tz = timezone.String()
	}
	if tz == "" {
		tz = "UTC"
	}

	//tx, err := db.db.BeginTx(db.ctx, nil)
	//if err != nil {
	//	return 0, err
	//}
	//defer func() {
	//	_ = tx.Rollback()
	//}()
	//qtx := db.q.WithTx(tx)

	// note: we let LastLogin be the zero-value for time.Time, which means never logged in.
	id, err := db.q.CreateUser(db.ctx, sqlc.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
		MagicLink:      magicLink,
		IsActive:       1,
		Clan:           clan,
		Timezone:       tz,
	})
	if err != nil {
		return nil, err
	}

	//err = tx.Commit()
	//if err != nil {
	//	return 0, err
	//}

	return &domains.User_t{
		ID:             domains.ID(id),
		Email:          email,
		Timezone:       timezone,
		HashedPassword: hashedPassword,
		MagicLink:      magicLink,
		Clan:           clan,
		Roles: struct {
			IsActive        bool
			IsAdministrator bool
			IsAuthenticated bool
			IsOperator      bool
			IsUser          bool
		}{
			IsActive: true,
			IsUser:   true,
		},
		Created: time.Now(),
		Updated: time.Now(),
	}, nil
}

func (db *DB) DeleteUserByClan(clan string) error {
	if clanNo, err := strconv.Atoi(clan); err != nil || clanNo < 1 || clanNo > 999 {
		return domains.ErrInvalidClan
	}

	return db.q.DeleteUserByClan(db.ctx, clan)
}

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

func (db *DB) GetUser(userID domains.ID) (*domains.User_t, error) {
	row, err := db.q.GetUser(db.ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	// convert row.Timezone to a time.Location
	loc, err := time.LoadLocation(row.Timezone)
	if err != nil {
		return nil, err
	}

	return &domains.User_t{
		ID:       userID,
		Email:    row.Email,
		Timezone: loc,
		Clan:     row.Clan,
		Roles: struct {
			IsActive        bool
			IsAdministrator bool
			IsAuthenticated bool
			IsOperator      bool
			IsUser          bool
		}{
			IsActive:        row.IsActive == 1,
			IsAdministrator: row.IsAdministrator == 1,
			IsOperator:      row.IsOperator == 1,
			IsUser:          row.IsUser == 1,
		},
		Created:   row.CreatedAt,
		Updated:   row.UpdatedAt,
		LastLogin: row.LastLogin,
	}, nil
}

func (db *DB) GetUserByMagicLink(clan, magicLink string) (*domains.User_t, error) {
	if len(clan) == 0 || clan == "0000" {
		return nil, domains.ErrUnauthorized
	} else if magicLink == "" {
		return nil, domains.ErrUnauthorized
	}

	row, err := db.q.GetUserByClanAndMagicLink(db.ctx, sqlc.GetUserByClanAndMagicLinkParams{
		ClanID:    clan,
		MagicLink: magicLink,
	})
	if err != nil {
		return nil, err
	}
	// convert row.Timezone to a time.Location
	loc, err := time.LoadLocation(row.Timezone)
	if err != nil {
		return nil, err
	}

	return &domains.User_t{
		ID:       domains.ID(row.UserID),
		Email:    row.Email,
		Timezone: loc,
		Clan:     row.Clan,
		Roles: struct {
			IsActive        bool
			IsAdministrator bool
			IsAuthenticated bool
			IsOperator      bool
			IsUser          bool
		}{
			IsActive:        row.IsActive == 1,
			IsAdministrator: row.IsAdministrator == 1,
			IsOperator:      row.IsOperator == 1,
			IsUser:          row.IsUser == 1,
		},
		Created:   row.CreatedAt,
		Updated:   row.UpdatedAt,
		LastLogin: row.LastLogin,
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
