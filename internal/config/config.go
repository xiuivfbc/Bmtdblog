package config

import (
	"github.com/spf13/viper"
)

// Configuration 应用配置结构体
type Configuration struct {
	Addr          string              `mapstructure:"addr"`
	SignupEnabled bool                `mapstructure:"signup_enabled"`
	Title         string              `mapstructure:"title"`
	SessionSecret string              `mapstructure:"session_secret"`
	Domain        string              `mapstructure:"domain"`
	FileServer    string              `mapstructure:"file_server"`
	NotifyEmails  string              `mapstructure:"notify_emails"`
	PageSize      int                 `mapstructure:"page_size"`
	PublicDir     string              `mapstructure:"public"`
	ViewDir       string              `mapstructure:"view"`
	Dir           string              `mapstructure:"dir"`
	Mysql         Mysql               `mapstructure:"mysql"`
	Seo           Seo                 `mapstructure:"seo"`
	Qiniu         Qiniu               `mapstructure:"qiniu"`
	Smms          Smms                `mapstructure:"smms"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Github        Github              `mapstructure:"github"`
	Smtp          Smtp                `mapstructure:"smtp"`
	TLS           TLSConfig           `mapstructure:"tls"`
	Navigators    []Navigator         `mapstructure:"navigators"`
	Backup        Backup              `mapstructure:"backup"`
	Zap           ZapConfig           `mapstructure:"zap"`
}

// Mysql 数据库配置
type Mysql struct {
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	DbName       string `mapstructure:"db_name"`
	LogLevel     int    `mapstructure:"log_level"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleTime  int    `mapstructure:"max_idle_time"`
}

// Seo SEO配置
type Seo struct {
	Title       string   `mapstructure:"title"`
	Description string   `mapstructure:"description"`
	Author      Author   `mapstructure:"author"`
	Keywords    []string `mapstructure:"keywords"`
}

// Author 作者信息
type Author struct {
	Name  string `mapstructure:"name"`
	Email string `mapstructure:"email"`
}

// Qiniu 七牛云存储配置
type Qiniu struct {
	Enabled    bool   `mapstructure:"enabled"`
	Accesskey  string `mapstructure:"accesskey"`
	Secretkey  string `mapstructure:"secretkey"`
	Bucket     string `mapstructure:"bucket"`
	FileServer string `mapstructure:"fileserver"`
}

// Smms SMMS图床配置
type Smms struct {
	Enabled bool   `mapstructure:"enabled"`
	Apiurl  string `mapstructure:"apiurl"`
	Apikey  string `mapstructure:"apikey"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	URL       string `mapstructure:"url"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	IndexName string `mapstructure:"index_name"`
}

// Github GitHub OAuth配置
type Github struct {
	Enabled      bool   `mapstructure:"enabled"`
	ClientId     string `mapstructure:"clientid"`
	ClientSecret string `mapstructure:"clientsecret"`
	RedirectUrl  string `mapstructure:"redirecturl"`
	AuthUrl      string `mapstructure:"authurl"`
	TokenUrl     string `mapstructure:"tokenurl"`
	Scope        string `mapstructure:"scope"`
}

// Smtp SMTP邮件配置
type Smtp struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
}

// Navigator 导航配置
type Navigator struct {
	Title  string `mapstructure:"title"`
	URL    string `mapstructure:"url"`
	Target string `mapstructure:"target"`
}

// Backup 备份配置
type Backup struct {
	Enabled   bool   `mapstructure:"enabled"`
	BackupKey string `mapstructure:"backup_key"`
}

// TLSConfig TLS/SSL配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	AutoCert bool   `mapstructure:"auto_cert"`
	Domain   string `mapstructure:"domain"`
	Email    string `mapstructure:"email"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CertDir  string `mapstructure:"cert_dir"`
}

// ZapConfig zap日志配置
type ZapConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Level         string `mapstructure:"level"`
	EnableTraceID bool   `mapstructure:"enable_trace_id"`
	Encoding      string `mapstructure:"encoding"`
	OutputPath    string `mapstructure:"output_path"`
	MaxAge        int    `mapstructure:"max_age"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxSize       int    `mapstructure:"max_size"`
}

// 全局变量
var configuration *Configuration

// LoadConfiguration 加载配置文件
func LoadConfiguration(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	var config Configuration
	if err := v.Unmarshal(&config); err != nil {
		return err
	}

	configuration = &config
	return nil
}

// GetConfiguration 获取配置
func GetConfiguration() *Configuration {
	return configuration
}

// GetElasticsearchIndexName 获取ES索引名
func (cfg *Configuration) GetElasticsearchIndexName() string {
	if cfg.Elasticsearch.IndexName != "" {
		return cfg.Elasticsearch.IndexName
	}
	return "bmtdblog_posts" // 默认值
}
