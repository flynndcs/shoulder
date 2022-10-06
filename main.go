package main

import (
	internal "shoulder/internal"
)

const (
	EXCHANGE_TYPE = "fanout"
)

func main() {
	swagger, shoulderConfig := internal.GetConfig()
	channel, q := internal.GetChannel(shoulderConfig, EXCHANGE_TYPE)
	db := internal.GetDb(shoulderConfig)

	internal.InitServer(shoulderConfig, channel, db, q, swagger)
}
