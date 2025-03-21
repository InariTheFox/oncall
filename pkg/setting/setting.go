package setting

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/util"
	"gopkg.in/ini.v1"
)

const (
	RedactedPassword = "*********"
	DefaultHTTPAddr  = "0.0.0.0"
	Dev              = "development"
	Prod             = "production"
	ApplicationName  = "OnCall"
)

var (
	customInitPath = "conf/custom.ini"
	Env            = Dev
	AppUrl         string
	AppSubUrl      string
)

var skipStaticRootValidation = false

type Scheme string

const (
	HTTPScheme   Scheme = "http"
	HTTPSScheme  Scheme = "https"
	HTTP2Scheme  Scheme = "h2"
	SocketScheme Scheme = "socket"
)

type Cfg struct {
	Target []string
	Raw    *ini.File

	DataPath    string
	HomePath    string
	LogsPath    string
	PluginsPath string

	AppSubURL             string
	AppURL                string
	CertFile              string
	CertPassword          string
	Domain                string
	EnforceDomain         bool
	Env                   string
	HTTPAddr              string
	HTTPPort              string
	InstanceName          string
	KeyFile               string
	MinTLSVersion         string
	Protocol              Scheme
	ReadTimeout           time.Duration
	RouterLogging         bool
	ServeFromSubPath      bool
	StaticRootPath        string
	CustomResponseHeaders map[string]string

	RabbitMqExchangeName string
	RabbitMqHost         string
	RabbitMqUsername     string
	RabbitMqPassword     string
	RabbitMqPort         int
	RabbitMqQueueName    string
	RabbitMqVhost        string

	configFiles                  []string
	appliedCommandLineProperties []string
	appliedEnvOverrides          []string
}

type CommandLineArgs struct {
	Config   string
	HomePath string
	Args     []string
}

func EnvKey(sectionName string, keyName string) string {
	sN := strings.ToUpper(strings.ReplaceAll(sectionName, ".", "_"))
	sN = strings.ReplaceAll(sN, "-", "_")
	kN := strings.ToUpper(strings.ReplaceAll(keyName, ".", "_"))
	envKey := fmt.Sprintf("GF_%s_%s", sN, kN)
	return envKey
}

func (cfg *Cfg) Load(args CommandLineArgs) error {
	cfg.setHomePath(args)

	iniFile, err := cfg.loadConfiguration(args)
	if err != nil {
		return err
	}

	err = cfg.parseINIFile(iniFile)
	if err != nil {
		return err
	}

	return nil
}

func NewCfg() *Cfg {
	return &Cfg{
		Env: Dev,
	}
}

func NewCfgFromArgs(args CommandLineArgs) (*Cfg, error) {
	cfg := NewCfg()
	if err := cfg.Load(args); err != nil {
		return nil, err
	}

	return cfg, nil
}

func RedactedValue(key, value string) string {
	if value == "" {
		return ""
	}

	uppercased := strings.ToUpper(key)
	// Sensitive information: password, secrets etc
	for _, pattern := range []string{
		"PASSWORD",
		"SECRET",
		"PROVIDER_CONFIG",
		"PRIVATE_KEY",
		"SECRET_KEY",
		"CERTIFICATE",
		"ACCOUNT_KEY",
		"ENCRYPTION_KEY",
		"VAULT_TOKEN",
		"CLIENT_SECRET",
		"API_DB_PASS",
		"ID_FORWARDING_TOKEN$",
		"AUTHENTICATION_TOKEN$",
		"AUTH_TOKEN$",
		"RENDERER_TOKEN$",
		"API_TOKEN$",
		"WEBHOOK_TOKEN$",
		"INSTALL_TOKEN$",
	} {
		if match, err := regexp.MatchString(pattern, uppercased); match && err == nil {
			return RedactedPassword
		}
	}

	for _, exception := range []string{} {
		if strings.Contains(uppercased, exception) {
			return value
		}
	}

	if u, err := RedactedURL(value); err == nil {
		return u
	}

	return value
}

func RedactedURL(value string) (string, error) {
	// Value could be a list of URLs
	chunks := util.SplitString(value)

	for i, chunk := range chunks {
		var hasTmpPrefix bool
		const tmpPrefix = "http://"

		if !strings.Contains(chunk, "://") {
			chunk = tmpPrefix + chunk
			hasTmpPrefix = true
		}

		u, err := url.Parse(chunk)
		if err != nil {
			return "", err
		}

		redacted := u.Redacted()
		if hasTmpPrefix {
			redacted = strings.Replace(redacted, tmpPrefix, "", 1)
		}

		chunks[i] = redacted
	}

	if strings.Contains(value, ",") {
		return strings.Join(chunks, ","), nil
	}

	return strings.Join(chunks, " "), nil
}

func (cfg *Cfg) applyCommandLineDefaultProperties(props map[string]string, file *ini.File) {
	cfg.appliedCommandLineProperties = make([]string, 0)
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			keyString := fmt.Sprintf("default.%s.%s", section.Name(), key.Name())
			value, exists := props[keyString]
			if exists {
				key.SetValue(value)
				cfg.appliedCommandLineProperties = append(cfg.appliedCommandLineProperties,
					fmt.Sprintf("%s=%s", keyString, RedactedValue(keyString, value)))
			}
		}
	}
}

func (cfg *Cfg) applyCommandLineProperties(props map[string]string, file *ini.File) {
	for _, section := range file.Sections() {
		sectionName := section.Name() + "."
		if section.Name() == ini.DefaultSection {
			sectionName = ""
		}
		for _, key := range section.Keys() {
			keyString := sectionName + key.Name()
			value, exists := props[keyString]
			if exists {
				cfg.appliedCommandLineProperties = append(cfg.appliedCommandLineProperties, fmt.Sprintf("%s=%s", keyString, value))
				key.SetValue(value)
			}
		}
	}
}

func (cfg *Cfg) applyEnvVariableOverrides(file *ini.File) error {
	cfg.appliedEnvOverrides = make([]string, 0)
	for _, section := range file.Sections() {
		for _, key := range section.Keys() {
			envKey := EnvKey(section.Name(), key.Name())
			envValue := os.Getenv(envKey)

			if len(envValue) > 0 {
				key.SetValue(envValue)
				cfg.appliedEnvOverrides = append(cfg.appliedEnvOverrides, fmt.Sprintf("%s=%s", envKey, RedactedValue(envKey, envValue)))
			}
		}
	}

	return nil
}

func (cfg *Cfg) getCommandLineProperties(args []string) map[string]string {
	props := make(map[string]string)

	for _, arg := range args {
		if !strings.HasPrefix(arg, "cfg:") {
			continue
		}

		trimmed := strings.TrimPrefix(arg, "cfg:")
		parts := strings.Split(trimmed, "=")
		if len(parts) != 2 {
			fmt.Errorf("Invalid command line argument: %s\n", arg)
			os.Exit(1)
		}

		props[parts[0]] = parts[1]
	}
	return props
}

func (cfg *Cfg) loadConfiguration(args CommandLineArgs) (*ini.File, error) {
	// load config defaults
	defaultConfigFile := path.Join(cfg.HomePath, "conf/defaults.ini")
	cfg.configFiles = append(cfg.configFiles, defaultConfigFile)

	// check if config file exists
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		fmt.Println("oncall-server Init Failed: Could not find config defaults, make sure homepath command line parameter is set or working directory is homepath")
		os.Exit(1)
	}

	// load defaults
	parsedFile, err := ini.Load(defaultConfigFile)
	if err != nil {
		fmt.Printf("Failed to parse defaults.ini, %v\n", err)
		os.Exit(1)
		return nil, err
	}

	// command line props
	commandLineProps := cfg.getCommandLineProperties(args.Args)
	// load default overrides
	cfg.applyCommandLineDefaultProperties(commandLineProps, parsedFile)

	// load specified config file
	err = cfg.loadSpecifiedConfigFile(args.Config, parsedFile)
	if err != nil {
		fmt.Errorf("%s\n", err.Error())
		os.Exit(1)
	}

	// apply environment overrides
	err = cfg.applyEnvVariableOverrides(parsedFile)
	if err != nil {
		return nil, err
	}

	// apply command line overrides
	cfg.applyCommandLineProperties(commandLineProps, parsedFile)

	// evaluate config values containing environment variables
	err = expandConfig(parsedFile)
	if err != nil {
		return nil, err
	}

	// update data path and logging config
	dataPath := valueAsString(parsedFile.Section("paths"), "data", "")

	cfg.DataPath = makeAbsolute(dataPath, cfg.HomePath)

	fmt.Printf("Starting %s: Version %s\n", ApplicationName, "1.0.0-alpha")

	return parsedFile, err
}

func (cfg *Cfg) loadSpecifiedConfigFile(configFile string, masterFile *ini.File) error {
	if configFile == "" {
		configFile = filepath.Join(cfg.HomePath, customInitPath)
		// return without error if custom file does not exist
		if !pathExists(configFile) {
			return nil
		}
	}

	userConfig, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to parse %q: %w\n", configFile, err)
	}

	// micro-optimization since we don't need to share this ini file. In
	// general, prefer to leave this flag as true as it is by default to prevent
	// data races
	userConfig.BlockMode = false

	for _, section := range userConfig.Sections() {
		for _, key := range section.Keys() {
			if key.Value() == "" {
				continue
			}

			defaultSec, err := masterFile.GetSection(section.Name())
			if err != nil {
				defaultSec, _ = masterFile.NewSection(section.Name())
			}
			defaultKey, err := defaultSec.GetKey(key.Name())
			if err != nil {
				defaultKey, _ = defaultSec.NewKey(key.Name(), key.Value())
			}
			defaultKey.SetValue(key.Value())
		}
	}

	cfg.configFiles = append(cfg.configFiles, configFile)
	return nil
}

func makeAbsolute(path string, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func (cfg *Cfg) parseINIFile(iniFile *ini.File) error {
	cfg.Raw = iniFile

	cfg.Env = valueAsString(iniFile.Section(""), "app_mode", "development")

	if err := cfg.readServerSettings(iniFile); err != nil {
		return err
	}

	if err := cfg.readRabbitMqSettings(iniFile); err != nil {
		return err
	}

	return nil
}

func (cfg *Cfg) parseAppUrlAndSubUrl(section *ini.Section) (string, string, error) {
	appUrl := valueAsString(section, "root_url", "http://localhost:3100/")

	if appUrl[len(appUrl)-1] != '/' {
		appUrl += "/"
	}

	// Check if has app suburl.
	url, err := url.Parse(appUrl)
	if err != nil {
		fmt.Errorf("Invalid root_url %s: %w\n", appUrl, err)
		os.Exit(1)
	}

	appSubUrl := strings.TrimSuffix(url.Path, "/")
	return appUrl, appSubUrl, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (cfg *Cfg) readRabbitMqSettings(iniFile *ini.File) error {
	rabbit := iniFile.Section("rabbit")
	var err error

	cfg.RabbitMqHost = rabbit.Key("hostname").String()
	port, err := rabbit.Key("port").Int()
	if err != nil {
		return err
	}
	cfg.RabbitMqPort = port
	cfg.RabbitMqUsername = rabbit.Key("username").String()
	cfg.RabbitMqPassword = rabbit.Key("password").String()
	cfg.RabbitMqVhost = rabbit.Key("vhost").String()
	cfg.RabbitMqQueueName = rabbit.Key("queue_name").String()
	cfg.RabbitMqExchangeName = rabbit.Key("exchange_name").String()

	return err
}

func (cfg *Cfg) readServerSettings(iniFile *ini.File) error {
	server := iniFile.Section("server")
	var err error
	AppUrl, AppSubUrl, err = cfg.parseAppUrlAndSubUrl(server)
	if err != nil {
		return err
	}

	cfg.AppURL = AppUrl
	cfg.AppSubURL = AppSubUrl
	cfg.Protocol = HTTPScheme
	cfg.ServeFromSubPath = server.Key("serve_from_sub_path").MustBool(false)

	protocolStr := valueAsString(server, "protocol", "http")

	if protocolStr == "https" {
		cfg.Protocol = HTTPSScheme
		cfg.CertFile = server.Key("cert_file").String()
		cfg.KeyFile = server.Key("cert_key").String()
		cfg.CertPassword = server.Key("cert_pass").String()
	}

	if protocolStr == "h2" {
		cfg.Protocol = HTTP2Scheme
		cfg.CertFile = server.Key("cert_file").String()
		cfg.KeyFile = server.Key("cert_key").String()
		cfg.CertPassword = server.Key("cert_pass").String()
	}

	cfg.MinTLSVersion = valueAsString(server, "min_tls_version", "TLS1.2")
	if cfg.MinTLSVersion == "TLS1.0" || cfg.MinTLSVersion == "TLS1.1" {
		return fmt.Errorf("TLS version not configured correctly:%v, allowed values are TLS1.2 and TLS1.3\n", cfg.MinTLSVersion)
	}

	cfg.Domain = valueAsString(server, "domain", "localhost")
	cfg.HTTPAddr = valueAsString(server, "http_addr", DefaultHTTPAddr)
	cfg.HTTPPort = valueAsString(server, "http_port", "3100")
	cfg.RouterLogging = server.Key("router_logging").MustBool(false)

	cfg.EnforceDomain = server.Key("enforce_domain").MustBool(false)
	staticRoot := valueAsString(server, "static_root_path", "")
	cfg.StaticRootPath = makeAbsolute(staticRoot, cfg.HomePath)

	if err := cfg.validateStaticRootPath(); err != nil {
		return err
	}

	cfg.ReadTimeout = server.Key("read_timeout").MustDuration(0)

	headersSection := cfg.Raw.Section("server.custom_response_headers")
	keys := headersSection.Keys()
	cfg.CustomResponseHeaders = make(map[string]string, len(keys))

	for _, key := range keys {
		cfg.CustomResponseHeaders[key.Name()] = key.Value()
	}

	return nil
}

func (cfg *Cfg) setHomePath(args CommandLineArgs) {
	if args.HomePath != "" {
		cfg.HomePath = args.HomePath
		return
	}

	var err error
	cfg.HomePath, err = filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	// check if homepath is correct
	if pathExists(filepath.Join(cfg.HomePath, "conf/defaults.ini")) {
		return
	}

	// try down one path
	if pathExists(filepath.Join(cfg.HomePath, "../conf/defaults.ini")) {
		cfg.HomePath = filepath.Join(cfg.HomePath, "../")
	}
}

func (cfg *Cfg) validateStaticRootPath() error {
	if skipStaticRootValidation {
		return nil
	}

	if _, err := os.Stat(path.Join(cfg.StaticRootPath, "build")); err != nil {
		fmt.Errorf("Failed to detect generated javascript files in public/build\n")
	}

	return nil
}

func valueAsString(section *ini.Section, keyName string, defaultValue string) string {
	return section.Key(keyName).MustString(defaultValue)
}
