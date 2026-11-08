// Copyright 2026 The core-geth Authors
// This file is part of the core-geth library.
//
// The core-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The core-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the core-geth library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestHashrateArgUnmarshal(t *testing.T) {
	tests := []struct {
		input   string
		want    uint64
		wantErr bool
	}{
		// Accepted by hexutil.Uint64 too.
		{input: `"0x0"`, want: 0},
		{input: `"0x500000"`, want: 0x500000},
		{input: `"0X500000"`, want: 0x500000},
		{input: `"0xffffffffffffffff"`, want: 0xffffffffffffffff},
		// Leading zeros, rejected by hexutil.Uint64.
		{input: `"0x00"`, want: 0},
		{input: `"0x0500000"`, want: 0x500000},
		{input: `"0x0000000000000000000000000000000000000000000000000000000000500000"`, want: 0x500000},
		{input: `"0x0000000000000000000000000000000000000000000000000000000000000000"`, want: 0},
		{input: `"0x000000000000000000000000000000000000000000000000ffffffffffffffff"`, want: 0xffffffffffffffff},
		// Rejected.
		{input: `5242880`, wantErr: true},
		{input: `null`, wantErr: true},
		{input: `""`, wantErr: true},
		{input: `"0x"`, wantErr: true},
		{input: `"500000"`, wantErr: true},
		{input: `"0xg00000"`, wantErr: true},
		// More than 64 bits, unpadded and padded.
		{input: `"0x10000000000000000"`, wantErr: true},
		{input: `"0x0000000000000000000000000000000000000000000000010000000000000000"`, wantErr: true},
		// 65 hex digits.
		{input: `"0x00000000000000000000000000000000000000000000000000000000000500000"`, wantErr: true},
	}
	for _, tt := range tests {
		var h HashrateArg
		err := json.Unmarshal([]byte(tt.input), &h)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got %d", tt.input, h)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.input, err)
			continue
		}
		if uint64(h) != tt.want {
			t.Errorf("%s: got %#x, want %#x", tt.input, uint64(h), tt.want)
		}
	}
}

// TestSubmitHashrateRPCLeadingZeros submits a hash rate left-padded to 32 bytes
// through the JSON-RPC layer, as reported in core-geth #494.
func TestSubmitHashrateRPCLeadingZeros(t *testing.T) {
	ethash := NewTester(nil, false)
	defer ethash.Close()

	server := rpc.NewServer()
	defer server.Stop()
	for _, api := range ethash.APIs(nil) {
		if err := server.RegisterName(api.Namespace, api.Service); err != nil {
			t.Fatal(err)
		}
	}
	client := rpc.DialInProc(server)
	defer client.Close()

	id := common.HexToHash("0x59daa26581d0acd1fce254fb7e85952f4c09d0915afd33d3886cd914bc7d283c")
	for _, method := range []string{"eth_submitHashrate", "ethash_submitHashrate"} {
		var ok bool
		if err := client.Call(&ok, method, "0x0000000000000000000000000000000000000000000000000000000000500000", id); err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		if !ok {
			t.Fatalf("%s: returned false", method)
		}
	}
	if got := ethash.Hashrate(); got != float64(0x500000) {
		t.Fatalf("hashrate: got %v, want %v", got, float64(0x500000))
	}
}
