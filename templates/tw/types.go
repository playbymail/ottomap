// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tw

type Layout_t struct {
	Site    Site_t
	Content any
}

type Site_t struct {
	Title string
}

type Clan_t struct {
	Id    string    // id of the player's clan
	Turns []*Turn_t // list of turns that the clan has uploaded reports for
}

type Turn_t struct {
	Id      string
	Reports []*Report_t // list of reports that the clan has uploaded for this turn
}

type Report_t struct {
	Id    string
	Clan  string    // id of the clan that owns the report
	Units []*Unit_t // list of units included in this report
	Map   string    // set when there is a map file
}

type Unit_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}

type ClanDetail_t struct {
	Id    string
	Maps  []string
	Turns []string
}

type TurnList_t struct {
	Turns []string
}

type TurnDetails_t struct {
	Id    string
	Clans []string
}

type TurnReportDetails_t struct {
	Id    string
	Clan  string
	Map   string // set when there is a map file
	Units []UnitDetails_t
}

type UnitDetails_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}
