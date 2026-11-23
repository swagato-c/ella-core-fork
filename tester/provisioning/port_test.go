// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package provisioning_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ellanetworks/core/tester/provisioning"
	"github.com/ellanetworks/core/tester/provisioning/ellacore"
)

func TestDialWithoutAdapterSaysSo(t *testing.T) {
	provisioning.Use(nil)

	if _, err := provisioning.Dial(context.Background(), provisioning.Target{Address: "http://core:5002"}); !errors.Is(err, provisioning.ErrNoAdapter) {
		t.Fatalf("Dial with no adapter: got %v, want ErrNoAdapter", err)
	}
}

func TestDialOpensTheRegisteredAdapter(t *testing.T) {
	provisioning.Use(ellacore.Open)
	t.Cleanup(func() { provisioning.Use(nil) })

	p, err := provisioning.Dial(context.Background(), provisioning.Target{Address: "http://core:5002", Token: "t"})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	if _, err := ellacore.Client(p); err != nil {
		t.Fatalf("the Ella Core adapter should expose its client: %v", err)
	}
}

type otherCore struct{ provisioning.Port }

func TestEllaCoreClientRefusesAnotherCoresAdapter(t *testing.T) {
	if _, err := ellacore.Client(otherCore{}); err == nil {
		t.Fatal("ellacore.Client accepted a non-Ella adapter")
	}
}
