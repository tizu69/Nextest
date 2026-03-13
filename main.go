package main

import (
	"errors"
	"log/slog"
	"net/http"

	"g.tizu.dev/Nextest/api"
	"g.tizu.dev/Nextest/config"
)

func main() {
	cfg, err1 := config.Load("./nextest.toml")
	if err := errors.Join(err1); err != nil {
		panic(err)
	}

	a := api.NewAPI(cfg)
	const addr = ":8080"
	slog.Info("Starting Nextest", "addr", addr)
	if err := http.ListenAndServe(addr, a.Routes()); err != nil {
		panic(err)
	}
}
