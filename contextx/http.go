package contextx

import (
	"context"
)

const (
	HeaderUserAgent       = "User-Agent"
	HeaderAcceptLanguage  = "Accept-Language"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderAuthorization   = "Authorization"
	HeaderAccept          = "Accept"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderReferer         = "Referer"
	HeaderOrigin          = "Origin"
	HeaderHost            = "Host"
	HeaderCookie          = "Cookie"
	HeaderSessionID       = "Session-Id"
	HeaderCacheControl    = "Cache-Control"
	HeaderCountry         = "Country"
	HeaderApplicationID   = "Application-Id"
	HeaderRequestID       = "Request-Id"
	HeaderAPIKey          = "Api-Key"
	HeaderVendorKey       = "Vendor-Key"
	HeaderLocale          = "Locale"
	HeaderTimezone        = "Timezone"
	HeaderCSRFToken       = "CSRF-Token"
	HeaderNonce           = "Nonce"
	HeaderUserID          = "User-Id"
	HeaderAnonymousUserID = "Anonymous-User-Id"
)

func GetHeaderUserAgent(ctx context.Context) string {
	return GetString(ctx, HeaderUserAgent)
}

func GetHeaderAcceptLanguage(ctx context.Context) string {
	return GetString(ctx, HeaderAcceptLanguage)
}

func GetHeaderAcceptEncoding(ctx context.Context) string {
	return GetString(ctx, HeaderAcceptEncoding)
}

func GetHeaderAuthorization(ctx context.Context) string {
	return GetString(ctx, HeaderAuthorization)
}

func GetHeaderAccept(ctx context.Context) string {
	return GetString(ctx, HeaderAccept)
}

func GetHeaderContentType(ctx context.Context) string {
	return GetString(ctx, HeaderContentType)
}

func GetHeaderContentLength(ctx context.Context) string {
	return GetString(ctx, HeaderContentLength)
}

func GetHeaderReferer(ctx context.Context) string {
	return GetString(ctx, HeaderReferer)
}

func GetHeaderOrigin(ctx context.Context) string {
	return GetString(ctx, HeaderOrigin)
}

func GetHeaderHost(ctx context.Context) string {
	return GetString(ctx, HeaderHost)
}

func GetHeaderCookie(ctx context.Context) string {
	return GetString(ctx, HeaderCookie)
}

func GetHeaderSessionID(ctx context.Context) string {
	return GetString(ctx, HeaderSessionID)
}

func GetHeaderCacheControl(ctx context.Context) string {
	return GetString(ctx, HeaderCacheControl)
}

func GetHeaderCountry(ctx context.Context) string {
	return GetString(ctx, HeaderCountry)
}

func GetHeaderApplicationID(ctx context.Context) string {
	return GetString(ctx, HeaderApplicationID)
}

func GetHeaderRequestID(ctx context.Context) string {
	return GetString(ctx, HeaderRequestID)
}

func GetHeaderAPIKey(ctx context.Context) string {
	return GetString(ctx, HeaderAPIKey)
}

func GetHeaderVendorKey(ctx context.Context) string {
	return GetString(ctx, HeaderVendorKey)
}

func GetHeaderLocale(ctx context.Context) string {
	return GetString(ctx, HeaderLocale)
}

func GetHeaderTimezone(ctx context.Context) string {
	return GetString(ctx, HeaderTimezone)
}

func GetHeaderCSRFToken(ctx context.Context) string {
	return GetString(ctx, HeaderCSRFToken)
}

func GetHeaderNonce(ctx context.Context) string {
	return GetString(ctx, HeaderNonce)
}

func GetHeaderUserID(ctx context.Context) string {
	return GetString(ctx, HeaderUserID)
}

func GetHeaderAnonymousUserID(ctx context.Context) string {
	return GetString(ctx, HeaderAnonymousUserID)
}
