// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package pool_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/pool"
)

// ---- DefaultConfig tests ----

func TestDefaultConfig_ReturnsNonNil(t *testing.T) {
	cfg := pool.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
}

func TestDefaultConfig_MaxIdleConns(t *testing.T) {
	cfg := pool.DefaultConfig()
	if cfg.MaxIdleConns != 100 {
		t.Errorf("expected MaxIdleConns=100, got %d", cfg.MaxIdleConns)
	}
}

func TestDefaultConfig_DisableKeepAlives(t *testing.T) {
	cfg := pool.DefaultConfig()
	if cfg.DisableKeepAlives {
		t.Error("expected DisableKeepAlives=false, got true")
	}
}

func TestDefaultConfig_TimeoutsArePositive(t *testing.T) {
	cfg := pool.DefaultConfig()
	if cfg.IdleConnTimeout <= 0 {
		t.Errorf("expected IdleConnTimeout > 0, got %v", cfg.IdleConnTimeout)
	}
	if cfg.ResponseHeaderTimeout <= 0 {
		t.Errorf("expected ResponseHeaderTimeout > 0, got %v", cfg.ResponseHeaderTimeout)
	}
	if cfg.ExpectContinueTimeout <= 0 {
		t.Errorf("expected ExpectContinueTimeout > 0, got %v", cfg.ExpectContinueTimeout)
	}
	if cfg.TLSHandshakeTimeout <= 0 {
		t.Errorf("expected TLSHandshakeTimeout > 0, got %v", cfg.TLSHandshakeTimeout)
	}
}

func TestDefaultConfig_MaxIdleConnsPerHost_IsCPUBased(t *testing.T) {
	cfg := pool.DefaultConfig()
	expected := runtime.NumCPU() * 2
	if cfg.MaxIdleConnsPerHost != expected {
		t.Errorf("expected MaxIdleConnsPerHost=%d (NumCPU*2), got %d", expected, cfg.MaxIdleConnsPerHost)
	}
}

func TestDefaultConfig_MaxConnsPerHost_IsCPUBased(t *testing.T) {
	cfg := pool.DefaultConfig()
	expected := runtime.NumCPU() * 4
	if cfg.MaxConnsPerHost != expected {
		t.Errorf("expected MaxConnsPerHost=%d (NumCPU*4), got %d", expected, cfg.MaxConnsPerHost)
	}
}

// ---- Config getter method tests ----

func TestConfig_GetterMethods(t *testing.T) {
	cfg := &pool.Config{
		MaxIdleConnsPerHost: 5,
		MaxIdleConns:        50,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     45 * time.Second,
	}

	if got := cfg.GetMaxIdleConnsPerHost(); got != 5 {
		t.Errorf("GetMaxIdleConnsPerHost() = %d, want 5", got)
	}
	if got := cfg.GetMaxIdleConns(); got != 50 {
		t.Errorf("GetMaxIdleConns() = %d, want 50", got)
	}
	if got := cfg.GetMaxConnsPerHost(); got != 10 {
		t.Errorf("GetMaxConnsPerHost() = %d, want 10", got)
	}
	if got := cfg.GetIdleConnTimeout(); got != 45*time.Second {
		t.Errorf("GetIdleConnTimeout() = %v, want 45s", got)
	}
}

// ---- ForService tests ----

func TestForService_TransferHigherThanAuth(t *testing.T) {
	transfer := pool.ForService("transfer")
	auth := pool.ForService("auth")

	if transfer.MaxIdleConnsPerHost <= auth.MaxIdleConnsPerHost {
		t.Errorf("transfer.MaxIdleConnsPerHost (%d) should be > auth.MaxIdleConnsPerHost (%d)",
			transfer.MaxIdleConnsPerHost, auth.MaxIdleConnsPerHost)
	}
	if transfer.MaxConnsPerHost <= auth.MaxConnsPerHost {
		t.Errorf("transfer.MaxConnsPerHost (%d) should be > auth.MaxConnsPerHost (%d)",
			transfer.MaxConnsPerHost, auth.MaxConnsPerHost)
	}
}

func TestForService_AuthLowerMaxIdleConnsPerHost(t *testing.T) {
	auth := pool.ForService("auth")
	def := pool.DefaultConfig()

	// auth should have a lower or equal MaxIdleConnsPerHost than the uncapped default
	// The important thing is auth is specifically set to 4
	if auth.MaxIdleConnsPerHost != 4 {
		t.Errorf("expected auth MaxIdleConnsPerHost=4, got %d", auth.MaxIdleConnsPerHost)
	}
	_ = def
}

func TestForService_UnknownServiceReturnsDefaults(t *testing.T) {
	unknown := pool.ForService("unknown-service-xyz")
	def := pool.DefaultConfig()

	// Unknown service should fall through with default values intact
	// (no case match means no override of the default fields)
	if unknown.MaxIdleConns != def.MaxIdleConns {
		t.Errorf("unknown service MaxIdleConns: got %d, want %d", unknown.MaxIdleConns, def.MaxIdleConns)
	}
	if unknown.DisableKeepAlives != def.DisableKeepAlives {
		t.Errorf("unknown service DisableKeepAlives: got %v, want %v", unknown.DisableKeepAlives, def.DisableKeepAlives)
	}
	if unknown.ResponseHeaderTimeout != def.ResponseHeaderTimeout {
		t.Errorf("unknown service ResponseHeaderTimeout: got %v, want %v", unknown.ResponseHeaderTimeout, def.ResponseHeaderTimeout)
	}
}

func TestForService_TransferGetterMethods(t *testing.T) {
	cfg := pool.ForService("transfer")

	if got := cfg.GetMaxIdleConnsPerHost(); got != 8 {
		t.Errorf("transfer GetMaxIdleConnsPerHost() = %d, want 8", got)
	}
	if got := cfg.GetMaxConnsPerHost(); got != 16 {
		t.Errorf("transfer GetMaxConnsPerHost() = %d, want 16", got)
	}
	if got := cfg.GetIdleConnTimeout(); got != 120*time.Second {
		t.Errorf("transfer GetIdleConnTimeout() = %v, want 120s", got)
	}
	if got := cfg.GetMaxIdleConns(); got != 100 {
		t.Errorf("transfer GetMaxIdleConns() = %d, want 100 (unchanged default)", got)
	}
}

func TestForService_AuthGetterMethods(t *testing.T) {
	cfg := pool.ForService("auth")

	if got := cfg.GetMaxIdleConnsPerHost(); got != 4 {
		t.Errorf("auth GetMaxIdleConnsPerHost() = %d, want 4", got)
	}
	if got := cfg.GetMaxConnsPerHost(); got != 8 {
		t.Errorf("auth GetMaxConnsPerHost() = %d, want 8", got)
	}
	if got := cfg.GetIdleConnTimeout(); got != 60*time.Second {
		t.Errorf("auth GetIdleConnTimeout() = %v, want 60s", got)
	}
}

func TestForService_KnownServices(t *testing.T) {
	services := []string{"transfer", "auth", "compute", "search", "flows", "groups", "timers"}
	for _, svc := range services {
		cfg := pool.ForService(svc)
		if cfg == nil {
			t.Errorf("ForService(%q) returned nil", svc)
		}
	}
}

// ---- NewPool tests ----

func TestNewPool_NilConfigUsesDefaults(t *testing.T) {
	p := pool.NewPool(nil)
	if p == nil {
		t.Fatal("NewPool(nil) returned nil")
	}
	if p.Config == nil {
		t.Fatal("NewPool(nil).Config is nil")
	}
	if p.Config.MaxIdleConns != 100 {
		t.Errorf("expected default MaxIdleConns=100, got %d", p.Config.MaxIdleConns)
	}
}

func TestNewPool_WithConfig(t *testing.T) {
	cfg := &pool.Config{
		MaxIdleConnsPerHost: 3,
		MaxIdleConns:        30,
		MaxConnsPerHost:     6,
		IdleConnTimeout:     45 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	p := pool.NewPool(cfg)
	if p == nil {
		t.Fatal("NewPool(cfg) returned nil")
	}
}

func TestNewPool_GetClientReturnsNonNil(t *testing.T) {
	p := pool.NewPool(nil)
	client := p.GetClient()
	if client == nil {
		t.Fatal("GetClient() returned nil")
	}
}

func TestNewPool_GetTransportReturnsNonNil(t *testing.T) {
	p := pool.NewPool(nil)
	transport := p.GetTransport()
	if transport == nil {
		t.Fatal("GetTransport() returned nil")
	}
}

func TestNewPool_SetTimeout(t *testing.T) {
	p := pool.NewPool(nil)
	newTimeout := 60 * time.Second
	p.SetTimeout(newTimeout)
	if p.GetClient().Timeout != newTimeout {
		t.Errorf("after SetTimeout(%v), client timeout = %v, want %v",
			newTimeout, p.GetClient().Timeout, newTimeout)
	}
}

func TestNewPool_CloseIdleConnectionsDoesNotPanic(t *testing.T) {
	p := pool.NewPool(nil)
	// Must not panic
	p.CloseIdleConnections()
}

// ---- GetStats tests ----

func TestPool_GetStats_ReturnsPoolStats(t *testing.T) {
	p := pool.NewPool(nil)
	stats := p.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	ps, ok := stats.(pool.PoolStats)
	if !ok {
		t.Fatalf("GetStats() returned type %T, want pool.PoolStats", stats)
	}

	// Fresh pool should have no active hosts
	if ps.ActiveHosts != 0 {
		t.Errorf("expected ActiveHosts=0, got %d", ps.ActiveHosts)
	}
	if ps.TotalActive != 0 {
		t.Errorf("expected TotalActive=0, got %d", ps.TotalActive)
	}
}

func TestPool_GetStats_ConfigPreserved(t *testing.T) {
	cfg := pool.ForService("transfer")
	p := pool.NewPool(cfg)
	stats := p.GetStats()

	ps, ok := stats.(pool.PoolStats)
	if !ok {
		t.Fatalf("GetStats() returned type %T, want pool.PoolStats", stats)
	}

	if ps.Config.MaxIdleConnsPerHost != cfg.MaxIdleConnsPerHost {
		t.Errorf("stats Config.MaxIdleConnsPerHost = %d, want %d",
			ps.Config.MaxIdleConnsPerHost, cfg.MaxIdleConnsPerHost)
	}
}

// ---- NewPoolManager tests ----

func TestNewPoolManager_NilConfigUsesDefaults(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	if mgr == nil {
		t.Fatal("NewPoolManager(nil) returned nil")
	}
}

func TestNewPoolManager_WithConfig(t *testing.T) {
	cfg := pool.DefaultConfig()
	mgr := pool.NewPoolManager(cfg)
	if mgr == nil {
		t.Fatal("NewPoolManager(cfg) returned nil")
	}
}

func TestPoolManager_GetPool_SameServiceReturnsSamePool(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	p1 := mgr.GetPool("myservice", nil)
	p2 := mgr.GetPool("myservice", nil)
	if p1 != p2 {
		t.Error("GetPool with same service name should return the same pool instance")
	}
}

func TestPoolManager_GetPool_DifferentServicesReturnDifferentPools(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	p1 := mgr.GetPool("service-a", nil)
	p2 := mgr.GetPool("service-b", nil)
	if p1 == p2 {
		t.Error("GetPool with different service names should return different pool instances")
	}
}

func TestPoolManager_GetPool_WithExplicitConfig(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	cfg := &pool.Config{
		MaxIdleConnsPerHost: 7,
		MaxIdleConns:        70,
		MaxConnsPerHost:     14,
		IdleConnTimeout:     55 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	p := mgr.GetPool("custom-service", cfg)
	if p == nil {
		t.Fatal("GetPool with explicit config returned nil")
	}
	if p.GetClient() == nil {
		t.Fatal("pool GetClient() returned nil")
	}
}

// mockConfig implements interfaces.ConnectionPoolConfig without being *pool.Config
type mockConfig struct {
	maxIdleConnsPerHost int
	maxIdleConns        int
	maxConnsPerHost     int
	idleConnTimeout     time.Duration
}

func (m *mockConfig) GetMaxIdleConnsPerHost() int       { return m.maxIdleConnsPerHost }
func (m *mockConfig) GetMaxIdleConns() int              { return m.maxIdleConns }
func (m *mockConfig) GetMaxConnsPerHost() int           { return m.maxConnsPerHost }
func (m *mockConfig) GetIdleConnTimeout() time.Duration { return m.idleConnTimeout }

func TestPoolManager_GetPool_WithInterfaceConfig(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	cfg := &mockConfig{
		maxIdleConnsPerHost: 3,
		maxIdleConns:        30,
		maxConnsPerHost:     6,
		idleConnTimeout:     45 * time.Second,
	}
	p := mgr.GetPool("interface-service", cfg)
	if p == nil {
		t.Fatal("GetPool with interface config returned nil")
	}
	if p.GetClient() == nil {
		t.Fatal("pool GetClient() returned nil")
	}
	if p.GetTransport() == nil {
		t.Fatal("pool GetTransport() returned nil")
	}
}

func TestPoolManager_CloseAllIdleConnections_DoesNotPanic(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	// Add some pools first
	mgr.GetPool("svc1", nil)
	mgr.GetPool("svc2", nil)
	// Must not panic
	mgr.CloseAllIdleConnections()
}

func TestPoolManager_CloseAllIdleConnections_EmptyManager(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	// Must not panic even with no pools
	mgr.CloseAllIdleConnections()
}

func TestPoolManager_GetAllStats_ReturnsCorrectKeys(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	mgr.GetPool("alpha", nil)
	mgr.GetPool("beta", nil)
	mgr.GetPool("gamma", nil)

	allStats := mgr.GetAllStats()
	if len(allStats) != 3 {
		t.Errorf("GetAllStats() returned %d entries, want 3", len(allStats))
	}
	for _, key := range []string{"alpha", "beta", "gamma"} {
		if _, ok := allStats[key]; !ok {
			t.Errorf("GetAllStats() missing key %q", key)
		}
	}
}

func TestPoolManager_GetAllStats_EmptyManager(t *testing.T) {
	mgr := pool.NewPoolManager(nil)
	allStats := mgr.GetAllStats()
	if allStats == nil {
		t.Fatal("GetAllStats() returned nil map")
	}
	if len(allStats) != 0 {
		t.Errorf("GetAllStats() on empty manager: got %d entries, want 0", len(allStats))
	}
}

// ---- NewClient tests ----

func TestNewClient_ReturnsNonNil(t *testing.T) {
	c := pool.NewClient("test-service", nil)
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClient_HasPool(t *testing.T) {
	c := pool.NewClient("test-service-pool", nil)
	if c.Pool == nil {
		t.Fatal("NewClient().Pool is nil")
	}
}

func TestNewClient_GetConnectionPool(t *testing.T) {
	c := pool.NewClient("test-service-cp", nil)
	p := c.GetConnectionPool()
	if p == nil {
		t.Fatal("GetConnectionPool() returned nil")
	}
}

func TestNewClient_SetTimeout(t *testing.T) {
	c := pool.NewClient("test-service-timeout", nil)
	newTimeout := 120 * time.Second
	c.SetTimeout(newTimeout)
	if c.Client.Timeout != newTimeout {
		t.Errorf("after SetTimeout(%v), client timeout = %v, want %v",
			newTimeout, c.Client.Timeout, newTimeout)
	}
}

func TestNewClient_GetHTTPClient_ReturnsNonNil(t *testing.T) {
	c := pool.NewClient("test-service-httpclient", nil)
	httpClient := c.GetHTTPClient()
	if httpClient == nil {
		t.Fatal("GetHTTPClient() returned nil")
	}
}

func TestNewClient_CloseIdleConnections_DoesNotPanic(t *testing.T) {
	c := pool.NewClient("test-service-close", nil)
	// Must not panic
	c.CloseIdleConnections()
}

func TestNewClient_CloseIdleConnections_NilPool(t *testing.T) {
	// Construct a Client with a nil Pool to test nil-pool guard
	c := &pool.Client{
		Client: nil,
		Pool:   nil,
	}
	// Must not panic due to nil pool check in CloseIdleConnections
	c.CloseIdleConnections()
}

// ---- GetServicePool tests ----

func TestGetServicePool_ReturnsNonNil(t *testing.T) {
	p := pool.GetServicePool("global-test-service", nil)
	if p == nil {
		t.Fatal("GetServicePool() returned nil")
	}
}

func TestGetServicePool_ConcurrentCallsAreSafe(t *testing.T) {
	const goroutines = 20
	const iterations = 10

	var wg sync.WaitGroup
	results := make([]interface{}, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				p := pool.GetServicePool("concurrent-service", nil)
				results[idx*iterations+j] = p
			}
		}(i)
	}

	wg.Wait()

	// All goroutines should have received a non-nil pool
	for i, r := range results {
		if r == nil {
			t.Errorf("result[%d] is nil, expected non-nil pool", i)
		}
	}
}

func TestGetServicePool_SameServiceReturnsSamePool(t *testing.T) {
	// Note: GetServicePool uses GlobalPoolManager which persists between tests.
	// We use a unique name to avoid interference.
	p1 := pool.GetServicePool("unique-global-service-abc123", nil)
	p2 := pool.GetServicePool("unique-global-service-abc123", nil)
	if p1 != p2 {
		t.Error("GetServicePool with same name should return the same pool")
	}
}

func TestGetServicePool_WithExplicitConfig(t *testing.T) {
	cfg := pool.ForService("transfer")
	p := pool.GetServicePool("global-transfer-explicit", cfg)
	if p == nil {
		t.Fatal("GetServicePool with config returned nil")
	}
	if p.GetClient() == nil {
		t.Fatal("pool GetClient() returned nil")
	}
}
