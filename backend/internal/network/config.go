package network

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

type ProxyType string

const (
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

type Config struct {
	ProxyEnabled bool      `mapstructure:"proxy_enabled" json:"proxyEnabled"`
	ProxyType    ProxyType `mapstructure:"proxy_type" json:"proxyType" default:"http"`
	ProxyHost    string    `mapstructure:"proxy_host" json:"proxyHost"`
	ProxyPort    int       `mapstructure:"proxy_port" json:"proxyPort"`
}

type HTTPClientProvider interface {
	HTTPClient(timeout time.Duration) *http.Client
}

type netContextDialer struct {
	dialer net.Dialer
}

func (d netContextDialer) Dial(network, address string) (net.Conn, error) {
	return d.dialer.Dial(network, address)
}

func (d netContextDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, address)
}

type Manager struct {
	mu        sync.RWMutex
	config    Config
	transport *http.Transport
	clients   map[time.Duration]*http.Client
}

func NewManager(config Config) (*Manager, error) {
	manager := &Manager{
		clients: make(map[time.Duration]*http.Client),
	}
	if err := manager.Reload(&config); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) HTTPClient(timeout time.Duration) *http.Client {
	m.mu.RLock()
	client := m.clients[timeout]
	m.mu.RUnlock()
	if client != nil {
		return client
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if client = m.clients[timeout]; client != nil {
		return client
	}
	client = &http.Client{
		Timeout:   timeout,
		Transport: m,
	}
	m.clients[timeout] = client
	return client
}

func (m *Manager) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.currentTransport().RoundTrip(req)
}

func (m *Manager) Reload(config interface{}) error {
	networkConfig, ok := config.(*Config)
	if !ok {
		return errs.NewBadRequest("invalid network config")
	}
	transport, err := newTransport(*networkConfig)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	oldTransport := m.transport
	m.config = *networkConfig
	m.transport = transport
	if oldTransport != nil {
		oldTransport.CloseIdleConnections()
	}
	return nil
}

func (m *Manager) currentTransport() *http.Transport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.transport == nil {
		return http.DefaultTransport.(*http.Transport).Clone()
	}
	return m.transport
}

func Validate(config Config) error {
	if !config.ProxyEnabled {
		return nil
	}
	switch config.ProxyType {
	case "", ProxyTypeHTTP, ProxyTypeSOCKS5:
		_, err := proxyAddress(config)
		return err
	default:
		return errs.NewBadRequest("不支持的代理类型")
	}
}

func newTransport(config Config) (*http.Transport, error) {
	if err := Validate(config); err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !config.ProxyEnabled {
		return transport, nil
	}
	switch config.ProxyType {
	case "", ProxyTypeHTTP:
		proxyURL, err := httpProxyURL(config)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	case ProxyTypeSOCKS5:
		if err := applySOCKS5Proxy(transport, config); err != nil {
			return nil, err
		}
	}
	return transport, nil
}

func httpProxyURL(config Config) (*url.URL, error) {
	address, err := proxyAddress(config)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Scheme: "http",
		Host:   address,
	}, nil
}

func applySOCKS5Proxy(transport *http.Transport, config Config) error {
	address, err := proxyAddress(config)
	if err != nil {
		return err
	}
	forward := netContextDialer{dialer: net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}}
	dialer, err := proxy.SOCKS5("tcp", address, nil, forward)
	if err != nil {
		return errs.NewBadRequest("SOCKS5 代理地址不可用")
	}
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		contextDialer, ok := dialer.(proxy.ContextDialer)
		if ok {
			return contextDialer.DialContext(ctx, network, address)
		}
		return dialer.Dial(network, address)
	}
	return nil
}

func proxyAddress(config Config) (string, error) {
	host, err := normalizeProxyHost(config.ProxyHost)
	if err != nil {
		return "", err
	}
	if config.ProxyPort <= 0 || config.ProxyPort > 65535 {
		return "", errs.NewBadRequest("代理端口必须在 1-65535 之间")
	}
	return net.JoinHostPort(host, strconv.Itoa(config.ProxyPort)), nil
}

func normalizeProxyHost(rawHost string) (string, error) {
	host := strings.TrimSpace(rawHost)
	if host == "" {
		return "", errs.NewBadRequest("代理 IP/Host 不能为空")
	}
	if strings.Contains(host, "://") || strings.ContainsAny(host, "/\\?#") {
		return "", errs.NewBadRequest("代理 IP/Host 不需要填写协议或路径")
	}
	if strings.ContainsAny(host, " \t\r\n") {
		return "", errs.NewBadRequest("代理 IP/Host 不能包含空白字符")
	}
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	}
	if strings.Contains(host, ":") && net.ParseIP(host) == nil {
		return "", errs.NewBadRequest("代理 IP/Host 不需要填写端口号")
	}
	return host, nil
}
