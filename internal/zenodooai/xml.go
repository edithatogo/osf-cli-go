package zenodooai

import (
	"encoding/xml"
	"strings"
	"time"
)

type oaiEnvelope struct {
	XMLName      xml.Name `xml:"OAI-PMH"`
	ResponseDate string   `xml:"responseDate"`
	Error        struct {
		Code    string `xml:"code,attr"`
		Message string `xml:",chardata"`
	} `xml:"error"`
	ListRecords struct {
		Records []xmlRecord `xml:"record"`
		Token   xmlToken    `xml:"resumptionToken"`
	} `xml:"ListRecords"`
	ListSets struct {
		Sets []struct {
			Spec        string   `xml:"setSpec"`
			Name        string   `xml:"setName"`
			Description innerXML `xml:"setDescription"`
		} `xml:"set"`
		Token xmlToken `xml:"resumptionToken"`
	} `xml:"ListSets"`
	Formats struct {
		Formats []struct {
			Prefix    string `xml:"metadataPrefix"`
			Schema    string `xml:"schema"`
			Namespace string `xml:"metadataNamespace"`
		} `xml:"metadataFormat"`
	} `xml:"ListMetadataFormats"`
}

type xmlRecord struct {
	Header struct {
		Status     string   `xml:"status,attr"`
		Identifier string   `xml:"identifier"`
		Datestamp  string   `xml:"datestamp"`
		SetSpecs   []string `xml:"setSpec"`
	} `xml:"header"`
	Metadata innerXML `xml:"metadata"`
	About    innerXML `xml:"about"`
}

type innerXML struct {
	Value string `xml:",innerxml"`
}

type xmlToken struct {
	Value            string `xml:",chardata"`
	ExpirationDate   string `xml:"expirationDate,attr"`
	Cursor           int    `xml:"cursor,attr"`
	CompleteListSize int    `xml:"completeListSize,attr"`
}

func (token xmlToken) value() (ResumptionToken, error) {
	var expiry time.Time
	if value := strings.TrimSpace(token.ExpirationDate); value != "" {
		var err error
		expiry, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return ResumptionToken{}, err
		}
	}
	return ResumptionToken{Value: strings.TrimSpace(token.Value), ExpiresAt: expiry, Cursor: token.Cursor, CompleteListSize: token.CompleteListSize}, nil
}

func (envelope oaiEnvelope) responseTime() (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(envelope.ResponseDate))
}
