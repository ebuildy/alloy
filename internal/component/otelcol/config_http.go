package otelcol

import (
	"time"

	"github.com/alecthomas/units"
	"github.com/grafana/agent/internal/component/otelcol/auth"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfigauth "go.opentelemetry.io/collector/config/configauth"
	otelconfighttp "go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	otelextension "go.opentelemetry.io/collector/extension"
)

// HTTPServerArguments holds shared settings for components which launch HTTP
// servers.
type HTTPServerArguments struct {
	Endpoint string `alloy:"endpoint,attr,optional"`

	TLS *TLSServerArguments `alloy:"tls,block,optional"`

	CORS *CORSArguments `alloy:"cors,block,optional"`

	// TODO(rfratto): auth
	//
	// Figuring out how to do authentication isn't very straightforward here. The
	// auth section links to an authenticator extension.
	//
	// We will need to generally figure out how we want to provide common
	// authentication extensions to all of our components.

	MaxRequestBodySize units.Base2Bytes `alloy:"max_request_body_size,attr,optional"`
	IncludeMetadata    bool             `alloy:"include_metadata,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *HTTPServerArguments) Convert() *otelconfighttp.HTTPServerSettings {
	if args == nil {
		return nil
	}

	return &otelconfighttp.HTTPServerSettings{
		Endpoint:           args.Endpoint,
		TLSSetting:         args.TLS.Convert(),
		CORS:               args.CORS.Convert(),
		MaxRequestBodySize: int64(args.MaxRequestBodySize),
		IncludeMetadata:    args.IncludeMetadata,
	}
}

// CORSArguments holds shared CORS settings for components which launch HTTP
// servers.
type CORSArguments struct {
	AllowedOrigins []string `alloy:"allowed_origins,attr,optional"`
	AllowedHeaders []string `alloy:"allowed_headers,attr,optional"`

	MaxAge int `alloy:"max_age,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *CORSArguments) Convert() *otelconfighttp.CORSSettings {
	if args == nil {
		return nil
	}

	return &otelconfighttp.CORSSettings{
		AllowedOrigins: args.AllowedOrigins,
		AllowedHeaders: args.AllowedHeaders,

		MaxAge: args.MaxAge,
	}
}

// HTTPClientArguments holds shared HTTP settings for components which launch
// HTTP clients.
type HTTPClientArguments struct {
	Endpoint string `alloy:"endpoint,attr"`

	Compression CompressionType `alloy:"compression,attr,optional"`

	TLS TLSClientArguments `alloy:"tls,block,optional"`

	ReadBufferSize  units.Base2Bytes  `alloy:"read_buffer_size,attr,optional"`
	WriteBufferSize units.Base2Bytes  `alloy:"write_buffer_size,attr,optional"`
	Timeout         time.Duration     `alloy:"timeout,attr,optional"`
	Headers         map[string]string `alloy:"headers,attr,optional"`
	// CustomRoundTripper  func(next http.RoundTripper) (http.RoundTripper, error) TODO (@tpaschalis)
	MaxIdleConns        *int           `alloy:"max_idle_conns,attr,optional"`
	MaxIdleConnsPerHost *int           `alloy:"max_idle_conns_per_host,attr,optional"`
	MaxConnsPerHost     *int           `alloy:"max_conns_per_host,attr,optional"`
	IdleConnTimeout     *time.Duration `alloy:"idle_conn_timeout,attr,optional"`
	DisableKeepAlives   bool           `alloy:"disable_keep_alives,attr,optional"`

	// Auth is a binding to an otelcol.auth.* component extension which handles
	// authentication.
	Auth *auth.Handler `alloy:"auth,attr,optional"`
}

// Convert converts args into the upstream type.
func (args *HTTPClientArguments) Convert() *otelconfighttp.HTTPClientSettings {
	if args == nil {
		return nil
	}

	// Configure the authentication if args.Auth is set.
	var auth *otelconfigauth.Authentication
	if args.Auth != nil {
		auth = &otelconfigauth.Authentication{AuthenticatorID: args.Auth.ID}
	}

	opaqueHeaders := make(map[string]configopaque.String)
	for headerName, headerVal := range args.Headers {
		opaqueHeaders[headerName] = configopaque.String(headerVal)
	}

	return &otelconfighttp.HTTPClientSettings{
		Endpoint: args.Endpoint,

		Compression: args.Compression.Convert(),

		TLSSetting: *args.TLS.Convert(),

		ReadBufferSize:  int(args.ReadBufferSize),
		WriteBufferSize: int(args.WriteBufferSize),
		Timeout:         args.Timeout,
		Headers:         opaqueHeaders,
		// CustomRoundTripper: func(http.RoundTripper) (http.RoundTripper, error) { panic("not implemented") }, TODO (@tpaschalis)
		MaxIdleConns:        args.MaxIdleConns,
		MaxIdleConnsPerHost: args.MaxIdleConnsPerHost,
		MaxConnsPerHost:     args.MaxConnsPerHost,
		IdleConnTimeout:     args.IdleConnTimeout,

		Auth: auth,
	}
}

// Extensions exposes extensions used by args.
func (args *HTTPClientArguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	m := make(map[otelcomponent.ID]otelextension.Extension)
	if args.Auth != nil {
		m[args.Auth.ID] = args.Auth.Extension
	}
	return m
}
