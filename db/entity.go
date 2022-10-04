package db

type Accretion struct {
	Key   int `gorm:"primaryKey"`
	Value string
}
