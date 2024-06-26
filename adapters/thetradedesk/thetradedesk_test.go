package thetradedesk

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/adapterstest"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestJsonSamples(t *testing.T) {
	bidder, err := Builder(
		openrtb_ext.BidderTheTradeDesk,
		config.Adapter{Endpoint: "https://direct.adsrvr.org/bid/bidder", ExtraAdapterInfo: "ttd"},
		config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "1"},
	)
	assert.Nil(t, err)
	adapterstest.RunJSONBidderTest(t, "thetradedesktest", bidder)
}

func TestGetBidType(t *testing.T) {
	type args struct {
		markupType openrtb2.MarkupType
	}
	tests := []struct {
		name              string
		args              args
		markupType        openrtb2.MarkupType
		expectedBidTypeId openrtb_ext.BidType
		wantErr           bool
	}{
		{
			name: "getBidType banner",
			args: args{
				markupType: openrtb2.MarkupBanner,
			},
			expectedBidTypeId: openrtb_ext.BidTypeBanner,
			wantErr:           false,
		},
		{
			name: "getBidType video",
			args: args{
				markupType: openrtb2.MarkupVideo,
			},
			expectedBidTypeId: openrtb_ext.BidTypeVideo,
			wantErr:           false,
		},
		{
			name: "getBidType audio",
			args: args{
				markupType: openrtb2.MarkupAudio,
			},
			expectedBidTypeId: openrtb_ext.BidTypeAudio,
			wantErr:           false,
		},
		{
			name: "getBidType native",
			args: args{
				markupType: openrtb2.MarkupNative,
			},
			expectedBidTypeId: openrtb_ext.BidTypeNative,
			wantErr:           false,
		},
		{
			name: "getBidType invalid",
			args: args{
				markupType: -1,
			},
			expectedBidTypeId: "",
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bidType, err := getBidType(tt.args.markupType)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedBidTypeId, bidType)
		})
	}
}

func TestGetPublisherId(t *testing.T) {
	type args struct {
		impressions []openrtb2.Imp
	}
	tests := []struct {
		name                string
		args                args
		expectedPublisherId string
		wantErr             bool
	}{
		{
			name: "valid publisher Id",
			args: args{
				impressions: []openrtb2.Imp{
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{"publisherId":"1"}}`),
					},
				},
			},
			expectedPublisherId: "1",
			wantErr:             false,
		},
		{
			name: "multiple valid publisher Id",
			args: args{
				impressions: []openrtb2.Imp{
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{"publisherId":"1"}}`),
					},
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{"publisherId":"2"}}`),
					},
				},
			},
			expectedPublisherId: "1",
			wantErr:             false,
		},
		{
			name: "not publisherId present",
			args: args{
				impressions: []openrtb2.Imp{
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{}}`),
					},
				},
			},
			expectedPublisherId: "",
			wantErr:             false,
		},
		{
			name: "nil publisherId present",
			args: args{
				impressions: []openrtb2.Imp{
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{"publisherId":""}}`),
					},
				},
			},
			expectedPublisherId: "",
			wantErr:             false,
		},
		{
			name: "no impressions",
			args: args{
				impressions: []openrtb2.Imp{},
			},
			expectedPublisherId: "",
			wantErr:             false,
		},
		{
			name: "invalid bidder object",
			args: args{
				impressions: []openrtb2.Imp{
					{
						Video: &openrtb2.Video{},
						Ext:   json.RawMessage(`{"bidder":{"doesnotexistprop":""}}`),
					},
				},
			},
			expectedPublisherId: "",
			wantErr:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisherId, err := getPublisherId(tt.args.impressions)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expectedPublisherId, publisherId)
		})
	}
}

func TestTheTradeDeskAdapter_MakeRequests(t *testing.T) {
	type fields struct {
		URI string
	}
	type args struct {
		request *openrtb2.BidRequest
		reqInfo *adapters.ExtraRequestInfo
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedReqData []*adapters.RequestData
		wantErr         bool
	}{
		{
			name: "invalid bidderparams",
			args: args{
				request: &openrtb2.BidRequest{Ext: json.RawMessage(`{"prebid":{"bidderparams":{:"123"}}}`)},
			},
			wantErr: true,
		},
		{
			name: "request with App",
			args: args{
				request: &openrtb2.BidRequest{
					App: &openrtb2.App{},
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":"123"}}}`),
				},
			},
			wantErr: false,
		},
		{
			name: "request with App and publisher",
			args: args{
				request: &openrtb2.BidRequest{
					App: &openrtb2.App{Publisher: &openrtb2.Publisher{}},
					Ext: json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":"123"}}}`),
				},
			},
			wantErr: false,
		},
		{
			name: "request with Site",
			args: args{
				request: &openrtb2.BidRequest{
					Site: &openrtb2.Site{},
					Ext:  json.RawMessage(`{"prebid":{"bidderparams":{"wrapper":"123"}}}`),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{
				URI: tt.fields.URI,
			}
			gotReqData, gotErr := a.MakeRequests(tt.args.request, tt.args.reqInfo)
			assert.Equal(t, tt.wantErr, len(gotErr) != 0)
			if tt.wantErr == false {
				assert.NotNil(t, gotReqData)
			}
		})
	}
}

func TestTheTradeDeskAdapter_MakeBids(t *testing.T) {
	type fields struct {
		URI string
	}
	type args struct {
		internalRequest *openrtb2.BidRequest
		externalRequest *adapters.RequestData
		response        *adapters.ResponseData
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  []error
		wantResp *adapters.BidderResponse
	}{
		{
			name: "happy path, valid response with all bid params",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"mtype": 1, "id": "7706636740145184841", "impid": "test-imp-id", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["ttd.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":{"dspid": 6, "deal_channel": 1, "prebiddealpriority": 1}}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"ttd.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`{"dspid": 6, "deal_channel": 1, "prebiddealpriority": 1}`),
							MType:   openrtb2.MarkupBanner,
						},
						BidType: openrtb_ext.BidTypeBanner,
					},
				},
				Currency: "USD",
			},
		},
		{
			name: "ignore invalid prebiddealpriority",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"id": "test-request-id", "seatbid":[{"seat": "958", "bid":[{"mtype": 2, "id": "7706636740145184841", "impid": "test-imp-id", "price": 0.500000, "adid": "29681110", "adm": "some-test-ad", "adomain":["ttd.com"], "crid": "29681110", "h": 250, "w": 300, "dealid": "testdeal", "ext":{"dspid": 6, "deal_channel": 1, "prebiddealpriority": -1}}]}], "bidid": "5778926625248726496", "cur": "USD"}`),
				},
			},
			wantErr: nil,
			wantResp: &adapters.BidderResponse{
				Bids: []*adapters.TypedBid{
					{
						Bid: &openrtb2.Bid{
							ID:      "7706636740145184841",
							ImpID:   "test-imp-id",
							Price:   0.500000,
							AdID:    "29681110",
							AdM:     "some-test-ad",
							ADomain: []string{"ttd.com"},
							CrID:    "29681110",
							H:       250,
							W:       300,
							DealID:  "testdeal",
							Ext:     json.RawMessage(`{"dspid": 6, "deal_channel": 1, "prebiddealpriority": -1}`),
							MType:   openrtb2.MarkupVideo,
						},
						BidType: openrtb_ext.BidTypeVideo,
					},
				},
				Currency: "USD",
			},
		},
		{
			name: "no content response",
			args: args{
				response: &adapters.ResponseData{
					StatusCode: http.StatusNoContent,
					Body:       nil,
				},
			},
			wantErr:  nil,
			wantResp: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &adapter{
				URI: tt.fields.URI,
			}
			gotResp, gotErr := a.MakeBids(tt.args.internalRequest, tt.args.externalRequest, tt.args.response)
			assert.Equal(t, tt.wantErr, gotErr, gotErr)
			assert.Equal(t, tt.wantResp, gotResp)
		})
	}
}
