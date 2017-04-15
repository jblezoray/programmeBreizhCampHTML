package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const jsonProgramFilename = "program.json"

type Talk struct {
	ID          int
	Name        string
	Event_Start string
	Event_End   string
	Event_Type  string
	Format      string
	Venue       string
	VenueID     string
	Speakers    string
	Description string
	// computed fields :
	CDay             int
	CDayMJV          string
	CStartHour       int
	CStartMinutes    int
	CEndHour         int
	CEndMinutes      int
	CDurationTotal   int
	CDurationHours   int
	CDurationMinutes int
	CSpeakers        []string
}

type talkSortByDate []Talk

func (s talkSortByDate) Len() int {
	return len(s)
}
func (s talkSortByDate) Less(i, j int) bool {
	if s[i].CDay < s[j].CDay {
		return true
	} else if s[i].CDay > s[j].CDay {
		return false
	} else if s[i].CStartHour < s[j].CStartHour {
		return true
	} else if s[i].CStartHour > s[j].CStartHour {
		return false
	} else if s[i].CStartMinutes < s[j].CStartMinutes {
		return true
	} else if s[i].CStartMinutes > s[j].CStartMinutes {
		return false
	} else if s[i].Venue < s[j].Venue {
		return true
	} else if s[i].Venue > s[j].Venue {
		return false
	}
	return false // equality ??
}
func (s talkSortByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var talks []Talk
	readFromFile(&talks, jsonProgramFilename)
	parseComputedFields(&talks)
	sort.Sort(talkSortByDate(talks))
	printTalks(talks)
}

func printTalks(talks []Talk) {
	for talkID := range talks {
		t := talks[talkID]
		fmt.Printf("| %s %02d:%02d | %1d:%02d | %-7s | %-14s | %40.40s | %-100.100s | %44s |\n",
			t.CDayMJV, t.CStartHour, t.CStartMinutes,
			t.CDurationHours, t.CDurationMinutes,
			t.Venue, t.Format, t.Event_Type, t.Name,
			strings.Join(t.CSpeakers, ", "))
	}
}

// readFromFile reads file as json.
// TODO dl from a GET on https://api.cfp.io/api/schedule with header "'X-Tenant-Id': 'breizhcamp'"
func readFromFile(talks *[]Talk, filename string) {
	file, err := os.Open(filename)
	check(err)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&talks)
	check(err)
}

func parseComputedFields(talks *[]Talk) {
	var t *Talk
	for talkID := range *talks {
		t = &(*talks)[talkID]
		t.CDay, _ = strconv.Atoi(t.Event_Start[8:10])
		if t.CDay == 19 {
			t.CDayMJV = "Me"
		} else if t.CDay == 20 {
			t.CDayMJV = "Je"
		} else if t.CDay == 21 {
			t.CDayMJV = "Ve"
		}
		t.CStartHour, _ = strconv.Atoi(t.Event_Start[11:13])
		t.CStartMinutes, _ = strconv.Atoi(t.Event_Start[14:16])
		t.CEndHour, _ = strconv.Atoi(t.Event_End[11:13])
		t.CEndMinutes, _ = strconv.Atoi(t.Event_End[14:16])
		t.CDurationTotal = (t.CEndHour-t.CStartHour)*60 - t.CStartMinutes + t.CEndMinutes
		t.CDurationHours = t.CDurationTotal / 60
		t.CDurationMinutes = t.CDurationTotal % 60

		// t.CSpeakers
		speakersSplited := strings.Split(t.Speakers, ",")
		for i := range speakersSplited {
			speaker := strings.TrimSpace(speakersSplited[i])
			if speaker != "null null" && len(speaker) > 0 {
				t.CSpeakers = append(t.CSpeakers, speaker)
			}
		}
	}
}
