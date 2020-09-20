package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/murakmii/exp-h2server/h2server"
)

func main() {
	cert, err := tls.LoadX509KeyPair(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}

	logger := h2server.NewLogger(log.New(os.Stdout, "[h2server] ", log.LstdFlags), h2server.DebugLog)
	sv := h2server.NewServer(&h2server.ServerConfig{
		Logger:      logger,
		Certificate: cert,
		Address:     ":443",
		Preface: func() *h2server.SettingsFrameBuilder {
			return h2server.NewSettingsFrameBuilder()
		},
		Multiplexer: h2server.DefaultMultiplexer(logger),
	})

	if err := sv.ListenAndServe(); err != nil {
		panic(err)
	}
}
