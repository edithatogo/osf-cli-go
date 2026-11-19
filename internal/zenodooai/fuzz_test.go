package zenodooai

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func FuzzDecodeEnvelope(f *testing.F) {
	for _, seed := range []string{
		`<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><responseDate>2026-07-15T00:00:00Z</responseDate><ListRecords/></OAI-PMH>`,
		`<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><error code="badArgument">bad</error></OAI-PMH>`,
		`<OAI-PMH>`,
	} {
		f.Add([]byte(seed), "ListRecords")
	}
	f.Fuzz(func(t *testing.T, data []byte, verb string) {
		if len(data) > 64*1024 || len(verb) > 128 {
			t.Skip()
		}
		first, err := decodeEnvelope(data, verb)
		if err != nil {
			return
		}
		if first.XMLName.Local != "OAI-PMH" || first.XMLName.Space != "http://www.openarchives.org/OAI/2.0/" {
			t.Fatal("accepted an envelope outside the OAI-PMH namespace")
		}
		second, err := decodeEnvelope(data, verb)
		if err != nil || !reflect.DeepEqual(first, second) {
			t.Fatal("decoding the same envelope changed the result")
		}
	})
}

func FuzzResumptionToken(f *testing.F) {
	f.Add(" next-page ", "2026-11-20T00:00:00Z", 0, 1)
	f.Add("", "", -1, 0)
	f.Add("synthetic", "not-a-date", 1, 2)
	f.Fuzz(func(t *testing.T, value, expiry string, cursor, total int) {
		if len(value) > 4096 || len(expiry) > 256 {
			t.Skip()
		}
		input := xmlToken{Value: value, ExpirationDate: expiry, Cursor: cursor, CompleteListSize: total}
		got, err := input.value()
		if err != nil {
			return
		}
		if got.Value != strings.TrimSpace(value) || got.Cursor != cursor || got.CompleteListSize != total {
			t.Fatal("token conversion lost cursor, count or opaque token data")
		}
		if strings.TrimSpace(expiry) == "" {
			if !got.ExpiresAt.IsZero() {
				t.Fatal("missing expiry acquired a timestamp")
			}
		} else {
			want, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(expiry))
			if parseErr != nil || !got.ExpiresAt.Equal(want) {
				t.Fatal("accepted expiry does not preserve its instant")
			}
		}
	})
}
