package zenodooai

import "testing"

func FuzzDecodeEnvelope(f *testing.F) {
	for _, seed := range []string{
		`<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><responseDate>2026-07-15T00:00:00Z</responseDate><ListRecords/></OAI-PMH>`,
		`<OAI-PMH xmlns="http://www.openarchives.org/OAI/2.0/"><error code="badArgument">bad</error></OAI-PMH>`,
		`<OAI-PMH>`,
	} {
		f.Add([]byte(seed), "ListRecords")
	}
	f.Fuzz(func(t *testing.T, data []byte, verb string) {
		_, _ = decodeEnvelope(data, verb)
	})
}
