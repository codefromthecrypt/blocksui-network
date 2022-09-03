package server

import (
	"blocksui-node/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func Start(conf *config.Config) {
	router := gin.Default()
	router.Use(cors.Default())

	// Routes
	// GET healthcheck
	// DELETE cache
	// POST /ipfs/api/v0/*
	// POST /
	// GET /assets-worker.js
	// GET *

	router.Run(conf.Port)
}
