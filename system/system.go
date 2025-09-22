package system

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type (
	Backup struct {
		Enabled   bool   `mapstructure:"enabled"`
		BackupKey string `mapstructure:"backup_key"`
	}

	Database struct {
		Dialect string `mapstructure:"dialect"`
		DSN     string `mapstructure:"dsn"`
	}

	Author struct {
		Name  string `mapstructure:"name"`
		Email string `mapstructure:"email"`
	}

	Seo struct {
		Description string `mapstructure:"description"`
		Author      Author `mapstructure:"author"`
	}

	Qiniu struct {
		Enabled    bool   `mapstructure:"enabled"`
		AccessKey  string `mapstructure:"accesskey"`
		SecretKey  string `mapstructure:"secretkey"`
		FileServer string `mapstructure:"fileserver"`
		Bucket     string `mapstructure:"bucket"`
	}

	Smms struct {
		Enabled bool   `mapstructure:"enabled"`
		ApiUrl  string `mapstructure:"apiurl"`
		ApiKey  string `mapstructure:"apikey"`
	}

	Github struct {
		Enabled      bool   `mapstructure:"enabled"`
		ClientId     string `mapstructure:"clientid"`
		ClientSecret string `mapstructure:"clientsecret"`
		RedirectURL  string `mapstructure:"redirecturl"`
		AuthUrl      string `mapstructure:"authurl"`
		TokenUrl     string `mapstructure:"tokenurl"`
		Scope        string `mapstructure:"scope"`
	}

	Smtp struct {
		Enabled  bool   `mapstructure:"enabled"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
	}

	Navigator struct {
		Title  string `mapstructure:"title"`
		Url    string `mapstructure:"url"`
		Target string `mapstructure:"target"`
	}

	Configuration struct {
		Addr          string      `mapstructure:"addr"`
		SignupEnabled bool        `mapstructure:"signup_enabled"`
		Title         string      `mapstructure:"title"`
		SessionSecret string      `mapstructure:"session_secret"`
		Domain        string      `mapstructure:"domain"`
		FileServer    string      `mapstructure:"file_server"`
		NotifyEmails  string      `mapstructure:"notify_emails"`
		PageSize      int         `mapstructure:"page_size"`
		PublicDir     string      `mapstructure:"public"`
		ViewDir       string      `mapstructure:"view"`
		Dir           string      `mapstructure:"dir"`
		Database      Database    `mapstructure:"database"`
		Seo           Seo         `mapstructure:"seo"`
		Qiniu         Qiniu       `mapstructure:"qiniu"`
		Smms          Smms        `mapstructure:"smms"`
		Github        Github      `mapstructure:"github"`
		Smtp          Smtp        `mapstructure:"smtp"`
		Navigators    []Navigator `mapstructure:"navigators"`
		Backup        Backup      `mapstructure:"backup"`
	}
)

var Logger *slog.Logger

func (a Author) String() string {
	return fmt.Sprintf("%s,%s", a.Name, a.Email)
}

var configuration *Configuration

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

func GetConfiguration() *Configuration {
	return configuration
}
