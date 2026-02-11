// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"testing"

	"github.com/playbymail/ottomap/internal/coords"
	schema "github.com/playbymail/ottomap/internal/tniif"
)

func TestSortOrder_OwningClanLast(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")

	events := []obsEvent_t{
		{Turn: "0901-01", Clan: "0987", Unit: "0987", Loc: loc},
		{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc},
		{Turn: "0901-01", Clan: "0987", Unit: "1987e1", Loc: loc},
		{Turn: "0901-01", Clan: "0138", Unit: "1138e1", Loc: loc},
	}
	sortEvents(events, "0987")

	for i, ev := range events {
		if ev.Clan == "0987" {
			for j := 0; j < i; j++ {
				if events[j].Clan == "0987" {
					continue
				}
			}
			for j := i + 1; j < len(events); j++ {
				if events[j].Clan != "0987" {
					t.Errorf("owning clan event at index %d followed by non-owning clan event at index %d", i, j)
				}
			}
			break
		}
	}

	for i := 0; i < len(events)-1; i++ {
		if events[i].Clan == "0138" && events[i+1].Clan == "0987" {
			continue
		}
		if events[i].Clan == events[i+1].Clan && events[i].Unit > events[i+1].Unit {
			t.Errorf("units out of order within clan %s: %s > %s", events[i].Clan, events[i].Unit, events[i+1].Unit)
		}
	}
}

func TestSortOrder_MultipleTurns(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")

	events := []obsEvent_t{
		{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc},
		{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc},
		{Turn: "0902-01", Clan: "0138", Unit: "0138", Loc: loc},
		{Turn: "0901-01", Clan: "0987", Unit: "0987", Loc: loc},
	}
	sortEvents(events, "0987")

	for i := 0; i < len(events)-1; i++ {
		if events[i].Turn > events[i+1].Turn {
			t.Fatalf("events not sorted by turn: event[%d].Turn=%s > event[%d].Turn=%s",
				i, events[i].Turn, i+1, events[i+1].Turn)
		}
	}

	type expected struct {
		Turn schema.TurnID
		Clan schema.ClanID
	}
	want := []expected{
		{"0901-01", "0138"},
		{"0901-01", "0987"},
		{"0902-01", "0138"},
		{"0902-01", "0987"},
	}
	for i, w := range want {
		if events[i].Turn != w.Turn || events[i].Clan != w.Clan {
			t.Errorf("event[%d]: got (turn=%s, clan=%s), want (turn=%s, clan=%s)",
				i, events[i].Turn, events[i].Clan, w.Turn, w.Clan)
		}
	}
}
