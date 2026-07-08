package network

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNewTransportWithHTTPProxy(t *testing.T) {
	transport, err := newTransport(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeHTTP,
		ProxyHost:    "30.27.194.25",
		ProxyPort:    7222,
	})
	if err != nil {
		t.Fatalf("newTransport() error = %v", err)
	}

	req := &http.Request{URL: mustParseURL(t, "https://mikanime.tv")}
	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("transport.Proxy() error = %v", err)
	}
	if proxyURL.String() != "http://30.27.194.25:7222" {
		t.Fatalf("proxyURL = %s, want http://30.27.194.25:7222", proxyURL.String())
	}
}

func TestNewTransportWithSOCKS5Proxy(t *testing.T) {
	transport, err := newTransport(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeSOCKS5,
		ProxyHost:    "30.27.194.25",
		ProxyPort:    7221,
	})
	if err != nil {
		t.Fatalf("newTransport() error = %v", err)
	}

	if transport.Proxy != nil {
		t.Fatal("transport.Proxy should be nil for socks5 proxy")
	}
	if transport.DialContext == nil {
		t.Fatal("transport.DialContext should be set for socks5 proxy")
	}
}

func TestManagerReloadUpdatesProxy(t *testing.T) {
	manager, err := NewManager(Config{})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if err := manager.Reload(&Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeHTTP,
		ProxyHost:    "30.27.194.25",
		ProxyPort:    7222,
	}); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	client := manager.HTTPClient(time.Second)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://mikanime.tv", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	transport, ok := client.Transport.(*Manager)
	if !ok {
		t.Fatalf("client.Transport = %T, want *Manager", client.Transport)
	}

	current := transport.currentTransport()
	proxyURL, err := current.Proxy(req)
	if err != nil {
		t.Fatalf("current.Proxy() error = %v", err)
	}
	if proxyURL.String() != "http://30.27.194.25:7222" {
		t.Fatalf("proxyURL = %s, want http://30.27.194.25:7222", proxyURL.String())
	}
}

func TestManagerHTTPClientReusesClientForSameTimeout(t *testing.T) {
	manager, err := NewManager(Config{})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	first := manager.HTTPClient(time.Second)
	second := manager.HTTPClient(time.Second)
	otherTimeout := manager.HTTPClient(2 * time.Second)

	if first != second {
		t.Fatal("HTTPClient() should reuse clients with the same timeout")
	}
	if first == otherTimeout {
		t.Fatal("HTTPClient() should keep separate clients for different timeouts")
	}
}

func TestValidateRejectsEnabledProxyWithoutAddress(t *testing.T) {
	err := Validate(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeHTTP,
	})
	if err == nil {
		t.Fatal("Validate() error is nil, want error")
	}
}

func TestValidateRejectsProxyHostWithScheme(t *testing.T) {
	err := Validate(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeHTTP,
		ProxyHost:    "http://30.27.194.25",
		ProxyPort:    7222,
	})
	if err == nil {
		t.Fatal("Validate() error is nil, want error")
	}
}

func TestValidateRejectsProxyHostWithPort(t *testing.T) {
	err := Validate(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeHTTP,
		ProxyHost:    "30.27.194.25:7222",
		ProxyPort:    7222,
	})
	if err == nil {
		t.Fatal("Validate() error is nil, want error")
	}
}

func TestValidateRejectsProxyPortOutOfRange(t *testing.T) {
	err := Validate(Config{
		ProxyEnabled: true,
		ProxyType:    ProxyTypeSOCKS5,
		ProxyHost:    "30.27.194.25",
		ProxyPort:    70000,
	})
	if err == nil {
		t.Fatal("Validate() error is nil, want error")
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	return parsed
}
