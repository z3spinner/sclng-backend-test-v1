package db

// DBError is a custom error type for database errors
type DBError string

func (e DBError) Error() string { return string(e) }
