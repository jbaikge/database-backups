package api

import "time"

type Database struct {
	Id            int        `json:"id"`
	ServerId      int        `json:"server_id"`
	Backup        bool       `json:"backup"`
	Name          string     `json:"name"`
	OnlyTables    string     `json:"only_tables"`
	ExcludeTables string     `json:"exclude_tables"`
	Added         time.Time  `json:"added"`
	Removed       *time.Time `json:"removed"`
}

type NewDatabaseRequest struct {
	ServerId int    `json:"server_id"`
	Name     string `json:"name"`
}

type NewServerRequest struct {
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ProxyHost     string `json:"proxy_host"`
	ProxyUsername string `json:"proxy_username"`
	ProxyIdentity string `json:"proxy_identity"`
}

type Server struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ProxyHost     string `json:"proxy_host"`
	ProxyUsername string `json:"proxy_username"`
	ProxyIdentity string `json:"proxy_identity"`
}

type Tree struct {
	Server    Server     `json:"server"`
	Databases []Database `json:"databases"`
}

type UpdateDatabaseRequest struct {
	ServerId      int    `json:"server_id"`
	Name          string `json:"name"`
	Backup        bool   `json:"backup"`
	OnlyTables    string `json:"only_tables"`
	ExcludeTables string `json:"exclude_tables"`
}
