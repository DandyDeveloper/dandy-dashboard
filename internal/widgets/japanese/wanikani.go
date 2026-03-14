package japanese

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const wkBaseURL = "https://api.wanikani.com/v2"

// --- WaniKani API types ---

type wkAssignmentData struct {
	SubjectID int `json:"subject_id"`
}

type wkAssignment struct {
	Data wkAssignmentData `json:"data"`
}

type wkPages struct {
	NextURL string `json:"next_url"`
}

type wkAssignmentResp struct {
	Data  []wkAssignment `json:"data"`
	Pages wkPages        `json:"pages"`
}

type wkReading struct {
	Reading string `json:"reading"`
	Primary bool   `json:"primary"`
}

type wkMeaning struct {
	Meaning string `json:"meaning"`
	Primary bool   `json:"primary"`
}

type wkSentence struct {
	En string `json:"en"`
	Ja string `json:"ja"`
}

type wkSubjectData struct {
	Characters       string       `json:"characters"`
	Level            int          `json:"level"`
	Readings         []wkReading  `json:"readings"`
	Meanings         []wkMeaning  `json:"meanings"`
	ContextSentences []wkSentence `json:"context_sentences"`
	PartsOfSpeech    []string     `json:"parts_of_speech"`
}

type wkSubjectResp struct {
	Data wkSubjectData `json:"data"`
}

// --- Client ---

type wkClient struct {
	token      string
	httpClient *http.Client
}

func (c *wkClient) get(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Wanikani-Revision", "20170710")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wanikani request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("wanikani: invalid API token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wanikani: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

// fetchVocabSubjectIDs returns the subject IDs of all vocabulary assignments,
// following pagination until every page has been consumed.
func (c *wkClient) fetchVocabSubjectIDs() ([]int, error) {
	var ids []int
	nextURL := wkBaseURL + "/assignments?subject_types=vocabulary"

	for nextURL != "" {
		var page wkAssignmentResp
		if err := c.get(nextURL, &page); err != nil {
			return nil, err
		}
		for _, a := range page.Data {
			ids = append(ids, a.Data.SubjectID)
		}
		nextURL = page.Pages.NextURL
	}

	return ids, nil
}

// fetchSubject fetches a single vocabulary subject by ID and maps it to a WordEntry.
func (c *wkClient) fetchSubject(id int) (*WordEntry, error) {
	var resp wkSubjectResp
	if err := c.get(fmt.Sprintf("%s/subjects/%d", wkBaseURL, id), &resp); err != nil {
		return nil, err
	}
	return subjectToEntry(resp.Data), nil
}

// subjectToEntry maps WaniKani subject data to the shared WordEntry shape.
func subjectToEntry(d wkSubjectData) *WordEntry {
	entry := &WordEntry{
		Word:  d.Characters,
		Level: fmt.Sprintf("WK Lv. %d", d.Level),
	}

	for _, r := range d.Readings {
		if r.Primary {
			entry.Reading = r.Reading
			break
		}
	}

	// Primary meaning first, then secondaries.
	for _, m := range d.Meanings {
		if m.Primary {
			entry.Meanings = append([]string{m.Meaning}, entry.Meanings...)
		} else {
			entry.Meanings = append(entry.Meanings, m.Meaning)
		}
	}
	if len(entry.Meanings) > 5 {
		entry.Meanings = entry.Meanings[:5]
	}

	for i, s := range d.ContextSentences {
		if i >= 3 {
			break
		}
		entry.Examples = append(entry.Examples, Example{Japanese: s.Ja, English: s.En})
	}

	return entry
}
