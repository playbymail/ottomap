// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlite

type Report struct {
	ID      int64
	Clan    int64
	Year    int64
	Month   int64
	Unit    string
	Hash    string
	Lines   string
	Created int64
}
