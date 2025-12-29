package traefik_maintenance_plugin

import (
	"context"
	"net/http"
)

type Config struct {
	Enabled      bool     `yaml:"enabled"`
	BypassSecret string   `yaml:"bypassSecret"`
	WhitelistIps []string `yaml:"whitelistIps"`
}

type Maintenance struct {
	name   string
	next   http.Handler
	config *Config
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func CreateConfig() *Config {
	return &Config{
		Enabled:      false,
		BypassSecret: "bypass",
	}
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	//go Inform(config)

	return &Maintenance{
		name:   name,
		next:   next,
		config: config,
	}, nil
}

func (a *Maintenance) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if !a.config.Enabled || a.bypassingHeaders(req) || a.clientIpIsWhitelisted(req) {
		rw := &responseWriter{ResponseWriter: w}
		a.next.ServeHTTP(rw, req)

		return
	}

	bytes := getMaintenanceTemplate()

	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write(bytes)
}

// Maintenance page templates
func getMaintenanceTemplate() []byte {
	return []byte(`<!doctype html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport"
			content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Under maintenance</title>
	<style>
		body {
			text-align: center;
		}

		h1 {
			font-size: 42px;
		}

		body {
			font: 20px Helvetica, sans-serif;
			color: #333;
		}

		article {
			display: block;
			text-align: left;
			margin: auto;
			max-width: 640px;
			min-width: 320px;
			padding: 10% 32px;
		}

		a {
			color: #0047AA;
			text-decoration: none;
		}

		a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
<article>
	<h1>Under maintenance</h1>
	<p><strong>Infos:</strong> <a target="_blank" href="https://siabit.ch/updates/">Updates - siabit AG</a></p>
	<p>Wir sind gerade dabei, unsere Infrastruktur zu aktualisieren und zu verbessern. Diese Website wird bald wieder verfügbar sein!</p>
	<p>Nous sommes en train de mettre à jour et d'améliorer notre infrastructure. Ce site sera bientôt de retour !</p>
	<p>We're currently updating and improving our infrastructure. This website will be back soon!</p>
</article>
</body>
</html>`)
}

func (a *Maintenance) bypassingHeaders(r *http.Request) bool {
	return r.Header.Get("X-Maintenance-Bypass") == a.config.BypassSecret
}

func (a *Maintenance) clientIpIsWhitelisted(r *http.Request) bool {
	for _, ip := range a.config.WhitelistIps {
		if ip == r.RemoteAddr || ip == r.Header.Get("X-Forwarded-For") {
			return true
		}
	}

	return false
}
