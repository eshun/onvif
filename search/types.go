package search

import "encoding/xml"

//查询回放概要信息
type GetRecordingSummary struct {
	XMLName xml.Name `xml:"tse:GetRecordingSummary"`
}

type GetRecordingSummaryResponse struct {
	Summary struct{
		DataFrom string
		DataUntil string
		NumberRecordings int
	}
}

//查询回放token
type FindRecordings struct {
	XMLName xml.Name `xml:"tse:FindRecordings"`
	Scope string	`xml:"tse:Scope"`
	KeepaliveTime string `xml:"tse:KeepAliveTime"`
}

type FindRecordingsResponse struct {
	SearchToken string
}

//查询录像记录
type GetRecordingSearchResults struct {
	XMLName xml.Name `xml:"tse:GetRecordingSearchResults"`
	SearchToken string `xml:"tse:SearchToken"`
	MinResults int `xml:"tse:MinResults"`
	MaxResults int `xml:"tse:MaxResults"`
	WaitTime string `xml:"tse:WaitTime"`
}

type GetRecordingSearchResultsResponse struct {
	ResultList struct{
		SearchState string
		RecordingInformation []struct{
			RecordingToken string
			Source struct{
				SourceId string
				Name string
				Location string
				Description string
				Address string
			}
			EarliestRecording string
			LatestRecording string
			Content string
			Track []struct{
				TrackToken string
				TrackType string
				Description string
				DataFrom string
				DataTo string
			}
			RecordingStatus string
		}
	}
}

//查询通道录像信息
type GetRecordingInformation struct {
	XMLName xml.Name `xml:"tse:GetRecordingInformation"`
	RecordingToken string `xml"tse:RecordingToken"`
}

type GetRecordingInformationResponse struct {
	RecordingInformation struct{
		EarliestRecording string
		LatestRecording string
	}
}