package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	tgauth "github.com/b4fun/tg-auth"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

func dumpReqest(req *http.Request) {
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Printf("dump output failed: %s\n", err)
		return
	}
	fmt.Println(string(b))
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	settings, err := tgauth.LoadEnvSettings()
	if err != nil {
		logger.Fatal("load settings", zap.Error(err))
		return
	}

	var admissioner tgauth.Admissioner
	{
		admissioner, err = tgauth.NewTelegramChannelAdmissioner(
			logger, settings.Bot, settings.Authz,
		)
		if err != nil {
			logger.Fatal("create admissioner", zap.Error(err))
			return
		}
		admissioner, err = tgauth.AdmissionerWithCache(admissioner, settings.Authn.SessionTTL)
		if err != nil {
			logger.Fatal("wrap admissioner with cache", zap.Error(err))
			return
		}
		admissioner, err = tgauth.AdmissionerWithSingleFlight(admissioner)
		if err != nil {
			logger.Fatal("wrap admissioner with single flight", zap.Error(err))
			return
		}
	}

	httpServer, err := tgauth.NewDefaultHTTPServer(
		logger, settings, admissioner,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dumpReqest(r)

			w.WriteHeader(http.StatusOK)
		}),
	)
	if err != nil {
		logger.Fatal("create http server", zap.Error(err))
		return
	}

	if err := http.ListenAndServe(":8082", httpServer); err != nil {
		logger.Fatal("http server serve", zap.Error(err))
		return
	}
}
