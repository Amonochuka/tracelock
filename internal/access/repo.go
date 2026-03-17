package access

import "database/sql"

type ZoneRepo struct{
	db *sql.DB
}

func NewZoneRepo(db *sql.DB)*ZoneRepo{
	return &ZoneRepo{db: db}
}

func(z *ZoneRepo)
-Createevent
GetLastEventHash
GetLastEventHash
CountActiveUsers
CreateSession
DeleteSession