package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Ottieni la directory corrente per costruire il percorso assoluto
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Il file config.json dovrebbe essere in ../config/config.json rispetto alla cartella test
	configPath := filepath.Join(cwd, "..", "config", "config.json")

	// Verifica che il file esista
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist at %s", configPath)
	}

	// Carica la configurazione
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verifica che la configurazione non sia nil
	assert.NotNil(t, cfg)

	// Verifica alcuni valori di default o dal file di esempio
	// Nota: questi valori dipendono dal contenuto di config.example.json
	assert.Equal(t, "glm-4.7", cfg.Agents.Defaults.Model)
	assert.Equal(t, 8192, cfg.Agents.Defaults.MaxTokens)
	
	// Verifica che le variabili d'ambiente vengano caricate (se impostate nel .env)
	// Qui stiamo testando solo il caricamento del file json e l'integrazione di base
}
