package authdelivery_test

import (
	"testing"

	"github.com/buger/jsonparser"
	"github.com/cmlight/authdelivery"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseBidRequest(t *testing.T) {
	testCases := []struct {
		desc            string
		inputBidRequest []byte

		wantParsedBidRequest authdelivery.ParsedBidRequest
		wantError            error
	}{
		{
			desc:            "zero-byte input message",
			inputBidRequest: []byte(""),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{},
			wantError:            jsonparser.MalformedObjectError, // FIXME: translate to proper error type.
		},
		{
			desc:            "empty JSON message",
			inputBidRequest: []byte("{}"),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{},
			wantError:            jsonparser.KeyPathNotFoundError, // FIXME: translate to proper error type.
		},
		{
			desc: "bid request without schain",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{},
			wantError:            jsonparser.KeyPathNotFoundError, // FIXME: translate to proper error type.
		},
		{
			desc: "bid request with single node schain",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
				"source": {
					"ext": {
						"schain": {
							"ver": "1.0",
							"complete": 1,
							"nodes": [
								{
									"asi": "directseller.com",
									"sid": "111111",
									"hp": 1
								}
							]
						}
					}
				}
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{
				SignatureMessageFragments: []string{"&schain.[0].asi=directseller.com&schain.[0].sid=111111"},
			},
			wantError: nil,
		},
		{
			desc: "bid request with four node schain",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
				"source": {
					"ext": {
						"schain": {
							"ver": "1.0",
							"complete": 1,
							"nodes": [
								{
									"asi": "directseller.com",
									"sid": "111111",
									"hp": 1
								},
								{
									"asi": "reseller.com",
									"sid": "222222",
									"hp": 1
								},
								{
									"asi": "exchange-1.com",
									"sid": "333333",
									"hp": 1
								},
								{
									"asi": "exchange-2.com",
									"sid": "444444",
									"hp": 1
								}
							]
						}
					}
				}
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{
				SignatureMessageFragments: []string{
					"&schain.[0].asi=directseller.com&schain.[0].sid=111111",
					"&schain.[1].asi=reseller.com&schain.[1].sid=222222",
					"&schain.[2].asi=exchange-1.com&schain.[2].sid=333333",
					"&schain.[3].asi=exchange-2.com&schain.[3].sid=444444",
				},
			},
			wantError: nil,
		},
		{
			desc: "bid request with four node schain and two protected fields",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
				"source": {
					"ext": {
						"schain": {
							"ver": "1.0",
							"complete": 1,
							"nodes": [
								{
									"asi": "directseller.com",
									"sid": "111111",
									"hp": 1,
									"params": "app.bundle"
								},
								{
									"asi": "reseller.com",
									"sid": "222222",
									"hp": 1
								},
								{
									"asi": "exchange-1.com",
									"sid": "333333",
									"hp": 1,
									"params": "app.name"
								},
								{
									"asi": "exchange-2.com",
									"sid": "444444",
									"hp": 1
								}
							]
						}
					}
				}
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{
				SignatureMessageFragments: []string{
					"&schain.[0].asi=directseller.com&schain.[0].sid=111111&app.bundle=com.app.test",
					"&schain.[1].asi=reseller.com&schain.[1].sid=222222",
					"&schain.[2].asi=exchange-1.com&schain.[2].sid=333333&app.name=TestApp",
					"&schain.[3].asi=exchange-2.com&schain.[3].sid=444444",
				},
			},
			wantError: nil,
		},
		{
			desc: "bid request with four node schain and two protected fields, one replacement",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
				"source": {
					"ext": {
						"schain": {
							"ver": "1.0",
							"complete": 1,
							"nodes": [
								{
									"asi": "directseller.com",
									"sid": "111111",
									"hp": 1,
									"params": "app.bundle"
								},
								{
									"asi": "reseller.com",
									"sid": "222222",
									"hp": 1
								},
								{
									"asi": "exchange-1.com",
									"sid": "333333",
									"hp": 1,
									"params": "app.name",
									"replace": "app.bundle=com.app.former"
								},
								{
									"asi": "exchange-2.com",
									"sid": "444444",
									"hp": 1
								}
							]
						}
					}
				}
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{
				SignatureMessageFragments: []string{
					"&schain.[0].asi=directseller.com&schain.[0].sid=111111&app.bundle=com.app.former",
					"&schain.[1].asi=reseller.com&schain.[1].sid=222222",
					"&schain.[2].asi=exchange-1.com&schain.[2].sid=333333&app.name=TestApp&app.bundle=com.app.test",
					"&schain.[3].asi=exchange-2.com&schain.[3].sid=444444",
				},
			},
			wantError: nil,
		},
		{
			desc: "bid request with four node schain and two protected fields, one double-replacement",
			inputBidRequest: []byte(`
			{
				"id": "BidRequest2",
				"app": {
					"bundle": "com.app.test",
					"cat": [
						"IAB22",
						"IAB33",
						"IAB44"
					],
					"domain": "com.app.test",
					"id": "123456",
					"name": "TestApp",
					"publisher": {
						"id": "12345"
					},
					"storeurl": "https://play.google.com/store/apps/details?id=com.app.test"
				},
				"source": {
					"ext": {
						"schain": {
							"ver": "1.0",
							"complete": 1,
							"nodes": [
								{
									"asi": "directseller.com",
									"sid": "111111",
									"hp": 1,
									"params": "app.bundle"
								},
								{
									"asi": "reseller.com",
									"sid": "222222",
									"hp": 1
								},
								{
									"asi": "exchange-1.com",
									"sid": "333333",
									"hp": 1,
									"params": "app.name",
									"replace": "app.bundle=com.app.original"
								},
								{
									"asi": "exchange-2.com",
									"sid": "444444",
									"hp": 1,
									"replace": "app.bundle=com.app.former"
								}
							]
						}
					}
				}
			}`),

			wantParsedBidRequest: authdelivery.ParsedBidRequest{
				SignatureMessageFragments: []string{
					"&schain.[0].asi=directseller.com&schain.[0].sid=111111&app.bundle=com.app.original",
					"&schain.[1].asi=reseller.com&schain.[1].sid=222222",
					"&schain.[2].asi=exchange-1.com&schain.[2].sid=333333&app.name=TestApp&app.bundle=com.app.former",
					"&schain.[3].asi=exchange-2.com&schain.[3].sid=444444&app.bundle=com.app.test",
				},
			},
			wantError: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotParsedBidRequest, gotError := authdelivery.ParseBidRequest(tC.inputBidRequest)
			if gotError != tC.wantError {
				t.Fatalf("error mismatch for scenario %q: \n\tgot:  %v\n\twant: %v", tC.desc, gotError, tC.wantError)
			}
			if diff := cmp.Diff(gotParsedBidRequest, tC.wantParsedBidRequest, protocmp.Transform()); diff != "" {
				t.Errorf("parsed bid request mismatch: %s", diff)
			}
		})
	}
}
