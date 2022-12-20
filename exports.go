package tgauth

import (
	"github.com/b4fun/tg-auth/internal/admission"
	"github.com/b4fun/tg-auth/internal/httpserver"
	"github.com/b4fun/tg-auth/internal/settings"
)

type (
	Settings       = settings.Settings
	SigninSettings = settings.SigninSettings
	BotSettings    = settings.BotSettings
	AuthnSettings  = settings.AuthnSettings
	AuthzSettings  = settings.AuthzSettings

	Admissioner = admission.Admissioner
)

var (
	LoadEnvSettings = settings.LoadEnvSettings

	NewDefaultHTTPServer = httpserver.Default

	NewTelegramChannelAdmissioner = admission.NewTelegramChannelAdmission
	AdmissionerWithCache          = admission.WithCache
	AdmissionerWithSingleFlight   = admission.WithSingleFlight
)
