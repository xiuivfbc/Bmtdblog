package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/xiuivfbc/bmtdblog/internal/config"
	"golang.org/x/crypto/acme/autocert"
)

// CreateAutoCertManager 创建自动证书管理器
func CreateAutoCertManager(domain, email, certDir string) *autocert.Manager {
	if certDir == "" {
		certDir = "certs"
	}

	// 确保证书目录存在
	if err := os.MkdirAll(certDir, 0755); err != nil {
		config.Logger.Error("创建证书目录失败", "dir", certDir, "err", err)
		return nil
	}

	config.Logger.Info("配置自动证书管理", "domain", domain, "email", email, "certDir", certDir)

	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(certDir),
		Email:      email,
	}
}

// StartWithAutoCert 使用自动证书启动HTTPS服务器
func StartWithAutoCert(srv *http.Server, domain, email, certDir string) error {
	m := CreateAutoCertManager(domain, email, certDir)
	if m == nil {
		return fmt.Errorf("failed to create autocert manager")
	}

	srv.TLSConfig = &tls.Config{
		GetCertificate: m.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1"},
		MinVersion:     tls.VersionTLS12,
	}

	// 启动HTTP挑战服务器（端口80）
	go func() {
		config.Logger.Info("启动HTTP挑战服务器在端口80")
		if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
			config.Logger.Error("HTTP挑战服务器错误", "err", err)
		}
	}()

	config.Logger.Info("启动HTTPS服务器", "addr", srv.Addr, "domain", domain)
	return srv.ListenAndServeTLS("", "")
}

// startTLSServer 启动TLS服务器
func startTLSServer(srv *http.Server, cfg *config.Configuration) error {
	if cfg.TLS.AutoCert && cfg.TLS.Domain != "" {
		// 自动证书模式
		config.Logger.Info("启动自动HTTPS服务", "addr", cfg.Addr, "domain", cfg.TLS.Domain)
		return StartWithAutoCert(srv, cfg.TLS.Domain, cfg.TLS.Email, cfg.TLS.CertDir)
	}

	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		// 手动证书模式
		config.Logger.Info("启动HTTPS服务", "addr", cfg.Addr)
		return srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	}

	// 配置不完整，回退到HTTP
	config.Logger.Error("TLS配置错误: 需要配置自动证书(auto_cert+domain)或手动证书(cert_file+key_file)")
	config.Logger.Info("TLS配置不完整，回退到HTTP服务", "addr", cfg.Addr)
	return srv.ListenAndServe()
}

// StartServer 启动HTTP/HTTPS服务器
func StartServer(srv *http.Server) error {
	cfg := config.GetConfiguration()

	if !cfg.TLS.Enabled {
		config.Logger.Info("启动HTTP服务", "addr", cfg.Addr)
		return srv.ListenAndServe()
	}

	return startTLSServer(srv, cfg)
}
