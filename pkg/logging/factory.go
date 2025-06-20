package logging

// New は新しいロガーを作成
// 優先順位: logLevel > DEBUG_MODE
func New(debugMode bool, logLevel Level) (Logger, error) {
	config := &Config{
		Level:            logLevel,
		Development:      false,
		Format:           "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// DEBUG_MODEがtrueの場合で、logLevelがInfoLevel（デフォルト）の場合はDebugレベルに変更
	if debugMode && logLevel == InfoLevel {
		config.Level = DebugLevel
		config.Development = true
	}

	// debugレベルの場合は開発モードも有効化
	if config.Level == DebugLevel {
		config.Development = true
	}

	return newZapLogger(config)
}

// NewWithConfig は詳細な設定を指定してロガーを作成
func NewWithConfig(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 現在はzapのみをサポート
	return newZapLogger(config)
}
