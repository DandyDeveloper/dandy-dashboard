package japanese

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dandydeveloper/dandy-dashboard/internal/store"
)

//go:embed wordlist.json
var wordlistJSON []byte

// KV store bucket and key constants.
const (
	bucket        = "japanese"
	keyToday      = "today"       // WordEntry JSON for the current day
	keyUsedPrefix = "used:"       // "used:<word>" or "used:wk:<id>" — marks items seen this cycle
	keyWKPool     = "wk_pool"     // JSON []int of WaniKani subject IDs (cached pool)
	keyWKPoolTS   = "wk_pool_ts"  // RFC3339 timestamp of last pool fetch
	wkPoolTTL     = 6 * time.Hour // How long before refreshing the WK assignment pool
)

// WordEntry is the response returned to the frontend.
type WordEntry struct {
	Word     string    `json:"word"`
	Reading  string    `json:"reading"`
	Meanings []string  `json:"meanings"`
	Level    string    `json:"level"`    // e.g. "WK Lv. 8" or "N3"
	Examples []Example `json:"examples"`
	Date     string    `json:"date"`
	Source   string    `json:"source"` // "wanikani" or "wordlist"
}

// Example is a sentence pair.
type Example struct {
	Japanese string `json:"japanese"`
	English  string `json:"english"`
}

// --- Service ---

type Service struct {
	wordlist   []string
	wk         *wkClient // nil when WaniKani is not configured
	store      store.Store
	httpClient *http.Client
}

func NewService(s store.Store, wkToken string) (*Service, error) {
	var words []string
	if err := json.Unmarshal(wordlistJSON, &words); err != nil {
		return nil, fmt.Errorf("parsing wordlist: %w", err)
	}

	svc := &Service{
		wordlist:   words,
		store:      s,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}

	if wkToken != "" {
		svc.wk = &wkClient{token: wkToken, httpClient: svc.httpClient}
	}

	return svc, nil
}

// GetWordOfDay returns today's word. If WaniKani is configured, the word is
// drawn from the user's active vocabulary assignments. Otherwise it falls back
// to the embedded wordlist + Jotoba.
func (s *Service) GetWordOfDay() (*WordEntry, error) {
	today := time.Now().UTC().Format("2006-01-02")

	if entry, err := s.loadToday(today); err == nil && entry != nil {
		return entry, nil
	}

	var (
		entry *WordEntry
		err   error
	)
	if s.wk != nil {
		entry, err = s.wkWordOfDay(today)
	} else {
		entry, err = s.wordlistWordOfDay(today)
	}
	if err != nil {
		return nil, err
	}

	if err := s.saveToday(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// ── WaniKani path ─────────────────────────────────────────────────────────────

func (s *Service) wkWordOfDay(today string) (*WordEntry, error) {
	pool, err := s.wkPool()
	if err != nil {
		return nil, err
	}

	id, err := s.pickWKSubject(pool)
	if err != nil {
		return nil, err
	}

	entry, err := s.wk.fetchSubject(id)
	if err != nil {
		return nil, err
	}

	entry.Date = today
	entry.Source = "wanikani"

	return entry, s.markWKUsed(id)
}

// wkPool returns the cached subject ID pool, refreshing it if stale.
func (s *Service) wkPool() ([]int, error) {
	raw, _ := s.store.Get(bucket, keyWKPool)
	tsRaw, _ := s.store.Get(bucket, keyWKPoolTS)

	if raw != nil && tsRaw != nil {
		if ts, err := time.Parse(time.RFC3339, string(tsRaw)); err == nil {
			if time.Since(ts) < wkPoolTTL {
				var ids []int
				if err := json.Unmarshal(raw, &ids); err == nil {
					return ids, nil
				}
			}
		}
	}

	// Pool is missing or stale — fetch from WaniKani.
	ids, err := s.wk.fetchVocabSubjectIDs()
	if err != nil {
		return nil, fmt.Errorf("fetching WaniKani assignments: %w", err)
	}
	if err := s.saveWKPool(ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Service) saveWKPool(ids []int) error {
	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	if err := s.store.Set(bucket, keyWKPool, raw); err != nil {
		return err
	}
	return s.store.Set(bucket, keyWKPoolTS, []byte(time.Now().UTC().Format(time.RFC3339)))
}

func (s *Service) pickWKSubject(pool []int) (int, error) {
	usedKeys, err := s.store.Keys(bucket)
	if err != nil {
		return 0, err
	}

	used := make(map[int]bool, len(usedKeys))
	prefix := keyUsedPrefix + "wk:"
	for _, k := range usedKeys {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			if id, err := strconv.Atoi(k[len(prefix):]); err == nil {
				used[id] = true
			}
		}
	}

	remaining := make([]int, 0, len(pool))
	for _, id := range pool {
		if !used[id] {
			remaining = append(remaining, id)
		}
	}

	if len(remaining) == 0 {
		if err := s.resetWKCycle(); err != nil {
			return 0, err
		}
		remaining = pool
	}

	return remaining[rand.IntN(len(remaining))], nil
}

func (s *Service) markWKUsed(id int) error {
	return s.store.Set(bucket, keyUsedPrefix+"wk:"+strconv.Itoa(id), []byte("1"))
}

func (s *Service) resetWKCycle() error {
	keys, err := s.store.Keys(bucket)
	if err != nil {
		return err
	}
	prefix := keyUsedPrefix + "wk:"
	for _, k := range keys {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			if err := s.store.Delete(bucket, k); err != nil {
				return err
			}
		}
	}
	return nil
}

// ── Wordlist + Jotoba path ────────────────────────────────────────────────────

func (s *Service) wordlistWordOfDay(today string) (*WordEntry, error) {
	word, err := s.pickWordlistWord()
	if err != nil {
		return nil, err
	}

	entry, err := s.fetchJotoba(word)
	if err != nil {
		return nil, err
	}

	entry.Date = today
	entry.Source = "wordlist"

	return entry, s.markWordlistUsed(word)
}

func (s *Service) pickWordlistWord() (string, error) {
	usedKeys, err := s.store.Keys(bucket)
	if err != nil {
		return "", err
	}

	prefix := keyUsedPrefix + "wl:"
	used := make(map[string]bool, len(usedKeys))
	for _, k := range usedKeys {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			used[k[len(prefix):]] = true
		}
	}

	remaining := make([]string, 0, len(s.wordlist))
	for _, w := range s.wordlist {
		if !used[w] {
			remaining = append(remaining, w)
		}
	}

	if len(remaining) == 0 {
		if err := s.resetWordlistCycle(); err != nil {
			return "", err
		}
		remaining = s.wordlist
	}

	return remaining[rand.IntN(len(remaining))], nil
}

func (s *Service) markWordlistUsed(word string) error {
	return s.store.Set(bucket, keyUsedPrefix+"wl:"+word, []byte("1"))
}

func (s *Service) resetWordlistCycle() error {
	keys, err := s.store.Keys(bucket)
	if err != nil {
		return err
	}
	prefix := keyUsedPrefix + "wl:"
	for _, k := range keys {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			if err := s.store.Delete(bucket, k); err != nil {
				return err
			}
		}
	}
	return nil
}

// ── Store helpers ─────────────────────────────────────────────────────────────

func (s *Service) loadToday(today string) (*WordEntry, error) {
	raw, err := s.store.Get(bucket, keyToday)
	if err != nil || raw == nil {
		return nil, err
	}
	var entry WordEntry
	if err := json.Unmarshal(raw, &entry); err != nil || entry.Date != today {
		return nil, nil
	}
	return &entry, nil
}

func (s *Service) saveToday(entry *WordEntry) error {
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.store.Set(bucket, keyToday, raw)
}

// ── Jotoba API ────────────────────────────────────────────────────────────────

type jotobaReading struct {
	Kana  string `json:"kana"`
	Kanji string `json:"kanji"`
}

type jotobaSense struct {
	Glosses []string `json:"glosses"`
}

type jotobaWord struct {
	Reading jotobaReading `json:"reading"`
	Senses  []jotobaSense `json:"senses"`
	JLPT    *int          `json:"jlpt"`
}

type jotobaSentence struct {
	Content     string `json:"content"`
	Translation string `json:"translation"`
}

type jotobaResponse struct {
	Words     []jotobaWord     `json:"words"`
	Sentences []jotobaSentence `json:"sentences"`
}

func (s *Service) fetchJotoba(word string) (*WordEntry, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://jotoba.de/api/search/words?q="+url.QueryEscape(word)+"&language=English",
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dandy-dashboard/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jotoba request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jresp jotobaResponse
	if err := json.Unmarshal(raw, &jresp); err != nil {
		return nil, fmt.Errorf("parsing jotoba response: %w", err)
	}

	entry := &WordEntry{Word: word}
	if len(jresp.Words) > 0 {
		w := jresp.Words[0]
		entry.Reading = w.Reading.Kana
		if w.Reading.Kanji != "" {
			entry.Word = w.Reading.Kanji
		}
		if w.JLPT != nil {
			entry.Level = fmt.Sprintf("N%d", *w.JLPT)
		}
		entry.Meanings = collectMeanings(w.Senses, 5)
	}
	entry.Examples = extractExamples(jresp.Sentences, 3)

	return entry, nil
}

func collectMeanings(senses []jotobaSense, limit int) []string {
	var out []string
	for _, sense := range senses {
		out = append(out, sense.Glosses...)
		if len(out) >= limit {
			return out[:limit]
		}
	}
	return out
}

func extractExamples(sentences []jotobaSentence, limit int) []Example {
	out := make([]Example, 0, limit)
	for _, s := range sentences {
		out = append(out, Example{Japanese: s.Content, English: s.Translation})
		if len(out) >= limit {
			break
		}
	}
	return out
}
