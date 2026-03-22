package entity

import "time"

type Server struct {
	Id         uint      `json:"id"`
	Name       string    `json:"name"`
	MacAddress string    `json:"mac_address"`
	IpAddress  string    `json:"ip_address"`
	Port       uint      `json:"port"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
