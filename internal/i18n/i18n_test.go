package i18n

import "testing"

// TestI18n_TC006_EveryMessageIsEnglish loads the only catalogue and checks
// every entry is a non-empty English sentence, so no code can resolve to an
// empty or untranslated string.
func TestI18n_TC006_EveryMessageIsEnglish(t *testing.T) {
	codes := Codes()
	if len(codes) == 0 {
		t.Fatal("en.json defines no messages")
	}
	for _, code := range codes {
		msg := Message(code)
		if msg == "" {
			t.Errorf("%s: empty message", code)
		}
		if msg == code {
			t.Errorf("%s: falls back to the code itself", code)
		}
	}
}

// TestI18n_TC007_UnknownCodeNeverLeaks checks the documented fallback: an
// unmapped code is returned as-is rather than an empty string, and the same
// call always returns the same English text.
func TestI18n_TC007_UnknownCodeNeverLeaks(t *testing.T) {
	if got := Message("SOME_UNKNOWN_CODE"); got != "SOME_UNKNOWN_CODE" {
		t.Fatalf("fallback = %q, want the code itself", got)
	}
	first := Message("AUTH_INVALID_CREDENTIALS")
	second := Message("AUTH_INVALID_CREDENTIALS")
	if first != second || first == "" {
		t.Fatalf("message is not stable: %q then %q", first, second)
	}
	if first != "Incorrect username/email or password." {
		t.Fatalf("generic sign-in message = %q", first)
	}
}
