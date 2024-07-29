package replay

import "encoding/xml"

//查询回放URI
type GetReplayUri struct {
	XMLName xml.Name `xml:"trp1:GetReplayUri"`
	StreamSetup struct{
		Stream string `xml:"trp1:Stream"`
		Transport struct{
			Protocol string `xml:"trp1:Protocol"`
		} `xml:"trp1:Transport"`
	} `xml:"trp1:StreamSetup"`
	RecordingToken string `xml:"trp1:RecordingToken"`
}

type GetReplayUriResponse struct {
	Uri string
}