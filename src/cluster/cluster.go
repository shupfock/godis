package cluster

import (
	"godis/src/config"
	"godis/src/db"
	"godis/src/lib/consistenthash"
)

type Cluster struct {
	self string

	peerPicker *consistenthash.Map

	db *db.DB
}

func NewCluster() *Cluster {
	cluster := &Cluster{
		self: config.Properties.Self,
		peerPicker: consistenthash.New()

		db: db.NewDB(),
	}
}
