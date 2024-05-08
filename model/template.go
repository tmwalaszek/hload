package model

import "time"

type Template struct {
	Name         string    `db:"name" json:"name"`
	CreationDate time.Time `db:"create_date" json:"creation_date"`
	UpdateDate   time.Time `db:"update_date" json:"update_date"`
	Content      string    `db:"content" json:"content"`
}
