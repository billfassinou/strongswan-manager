// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package httpapi

import "testing"

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 12: "12", 105: "105"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Fatalf("itoa(%d)=%s, attendu %s", in, got, want)
		}
	}
}
