// Package i18n resolves every stable error code to the message the API
// returns. English is the only language in v1 and the fallback for every
// missing string; the request's Accept-Language never changes the response.
package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed en.json
var enJSON []byte

var (
	once     sync.Once
	messages map[string]string
	loadErr  error
)

// Message returns the English text for code. A code with no entry is a
// programming error: the catalogue in en.json and the codes the code returns
// are kept in step by TestI18nCatalogueCoversEveryErrorCode.
func Message(code string) string {
	once.Do(load)
	if loadErr != nil {
		return code
	}
	if msg, ok := messages[code]; ok {
		return msg
	}
	return code
}

func load() {
	messages = map[string]string{}
	if err := json.Unmarshal(enJSON, &messages); err != nil {
		loadErr = fmt.Errorf("i18n: parse en.json: %w", err)
	}
}

// Codes returns every code the English catalogue defines.
func Codes() []string {
	once.Do(load)
	out := make([]string, 0, len(messages))
	for code := range messages {
		out = append(out, code)
	}
	return out
}
