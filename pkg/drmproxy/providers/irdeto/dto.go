package irdeto

import "encoding/xml"

// GetContentKeyResponse is a structured representation of Irdeto's response
// to a request for an asset's key information
type GetContentKeyResponse struct {
	XMLName        xml.Name       `xml:"CPIX"`
	Text           string         `xml:",chardata"`
	Ds             string         `xml:"ds,attr"`
	Cpix           string         `xml:"cpix,attr"`
	Xenc           string         `xml:"xenc,attr"`
	Pskc           string         `xml:"pskc,attr"`
	Speke          string         `xml:"speke,attr"`
	ContentId      string         `xml:"contentId,attr"`
	ContentKeyList ContentKeyList `xml:"ContentKeyList"`
	DRMSystemList  struct {
		Text      string `xml:",chardata"`
		DRMSystem struct {
			Text              string `xml:",chardata"`
			SystemId          string `xml:"systemId,attr"`
			Kid               string `xml:"kid,attr"`
			URIExtXKey        string `xml:"URIExtXKey"`
			KeyFormat         string `xml:"KeyFormat"`
			KeyFormatVersions string `xml:"KeyFormatVersions"`
		} `xml:"DRMSystem"`
	} `xml:"DRMSystemList"`
}

type ContentKeyList struct {
	Text       string     `xml:",chardata"`
	ContentKey ContentKey `xml:"ContentKey"`
}

type ContentKey struct {
	Text       string         `xml:",chardata"`
	Kid        string         `xml:"kid,attr"`
	ExplicitIV string         `xml:"explicitIV,attr"`
	Data       ContentKeyData `xml:"Data"`
}

type ContentKeyData struct {
	Text   string           `xml:",chardata"`
	Secret ContentKeySecret `xml:"Secret"`
}

type ContentKeySecret struct {
	Text       string `xml:",chardata"`
	PlainValue string `xml:"PlainValue"`
}

const fairplayReqTempl = `<cpix:CPIX contentId="{{.ContentID}}" xmlns:cpix="urn:dashif:org:cpix" xmlns:pskc="urn:ietf:params:xml:ns:keyprov:pskc" xmlns:speke="urn:aws:amazon:com:speke">

    <cpix:ContentKeyList>
        <cpix:ContentKey kid="{{.KeyID}}">

        </cpix:ContentKey>
    </cpix:ContentKeyList>
    <cpix:DRMSystemList>
        <cpix:DRMSystem systemId="94ce86fb-07ff-4f43-adb8-93d2fa968ca2" kid="{{.KeyID}}"/>
    </cpix:DRMSystemList>
</cpix:CPIX>`
