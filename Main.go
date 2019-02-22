package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

const jsonProgramOthersFilename = "input/program_others.json"
const jsonProgramFilename = "input/program.json"

const outputFilename = "index.html"

type TalkGroup struct {
	Talks         []Talk
	CDay          int
	CDayMJV       string
	CStartHour    int
	CStartMinutes int
}

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
	CDescriptionHTML string
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
	var talkOthers []Talk
	readFromFile(&talkOthers, jsonProgramOthersFilename)
	for i := range talkOthers {
		talkOther := talkOthers[i]
		talks = append(talks, talkOther)
	}
	parseComputedFields(&talks)
	sort.Sort(talkSortByDate(talks))
	printTalks(talks)
	talksByDateTime := groupByDateTime(talks)
	htmlData := dumpAsHTML(talksByDateTime)
	ioutil.WriteFile(outputFilename, []byte(htmlData), 0644)
}

func groupByDateTime(talks []Talk) []TalkGroup {
	var talkGroups []TalkGroup
	var curGroup *TalkGroup
	for i := range talks {
		if curGroup == nil || !sameDate(talks[i], *curGroup) {
			if curGroup != nil {
				talkGroups = append(talkGroups, *curGroup)
			}
			curGroup = &TalkGroup{}
			curGroup.CDay = talks[i].CDay
			curGroup.CDayMJV = talks[i].CDayMJV
			curGroup.CStartHour = talks[i].CStartHour
			curGroup.CStartMinutes = talks[i].CStartMinutes
		}
		curGroup.Talks = append(curGroup.Talks, talks[i])
	}
	return talkGroups
}

func sameDate(a Talk, tg TalkGroup) bool {
	return a.CDay == tg.CDay && a.CStartHour == tg.CStartHour && a.CStartMinutes == tg.CStartMinutes
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
		if t.CDay == 20 {
			t.CDayMJV = "Me"
		} else if t.CDay == 21 {
			t.CDayMJV = "Je"
		} else if t.CDay == 22 {
			t.CDayMJV = "Ve"
		}
		t.CStartHour, _ = strconv.Atoi(t.Event_Start[11:13])
		t.CStartMinutes, _ = strconv.Atoi(t.Event_Start[14:16])
		t.CEndHour, _ = strconv.Atoi(t.Event_End[11:13])
		t.CEndMinutes, _ = strconv.Atoi(t.Event_End[14:16])
		t.CDurationTotal = (t.CEndHour-t.CStartHour)*60 - t.CStartMinutes + t.CEndMinutes
		t.CDurationHours = t.CDurationTotal / 60
		t.CDurationMinutes = t.CDurationTotal % 60
		t.CDescriptionHTML = "<p class=\"desc\">" + t.Description + "</p>"
		t.CDescriptionHTML = strings.Replace(t.CDescriptionHTML, "\n", "<br/>", -1)
		t.CDescriptionHTML = strings.Replace(t.CDescriptionHTML, "<br/><br/>", "</p><p class=\"desc\">", -1)

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

const htmlTemplate = `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="fr" lang="fr" dir="ltr">
<head>
	<title>Output</title>
	<meta charset="UTF-8">
	<style>
		body {
			font-family: arial;
			font-size: 12px;
		}
		tr.tableHeader td {
			font-weight: bold;
			background-color:#DDDDDD;
		}
		td {
			border:1px solid black;
			padding: 3px;
			margin: 0;
		}
		table {
			width: 100%;
			border-spacing: 0;
			border-collapse: collapse;
		}
		h3 {
			width: 65%;
			float: left;
			margin-top: 20px;
			margin-left: 20px;
		}
		div.facts {
			text-align: right;
			width: 30%;
			margin-top: 20px;
			float: right;
		}
		p.desc {
			clear: both;
			margin-top: 0px;
			margin-left: 40px;
			margin-bottom: 7px;
		}
	</style>
</head>
<body>

<h1>Programme</h1>
<table>
	<tr class="tableHeader">
		<td>Durée</td>
		<td>Lieu</td>
		<td>Format</td>
		<td>Titre</td>
		<td>Speakers</td>
	</tr>
	{{range $talkGroup := .}}
		<tr class="tableHeader">
			<td colspan="5">
				{{printf "%s %02d:%02d" .CDayMJV .CStartHour .CStartMinutes}}
			</td>
		</tr>
		{{range $talk := .Talks}}
			<tr>
				<td style="white-space: nowrap;">
					{{printf "%1d:%02d" .CDurationHours .CDurationMinutes}}
				</td>
				<td>{{.Venue}}</td>
				<td>{{.Format}}</td>
				<td>{{.Name}}</td>
				<td>
					{{if not (eq (len .CSpeakers) 0) }}
						{{index .CSpeakers 0}}
					{{end}}
				</td>
			</tr>
		{{end}}
	{{end}}
</table>

<h1>Détail</h1>

{{range $talkGroup := .}}
	<hr />
	<h2>{{printf "%s %02d:%02d" .CDayMJV .CStartHour .CStartMinutes}}</h2>
	{{range $talk := .Talks}}
		<h3>{{.Name}}</h3>
		<div class="facts">
			[{{.Venue}}]
			({{printf "%1d:%02d" .CDurationHours .CDurationMinutes}})<br/>
			<i>{{.CSpeakers}}</i>
		</div>
		{{.CDescriptionHTML | unescaped}}
		<!--{{.Description}}-->
	{{end}}
{{end}}
</body>
</html>`

func dumpAsHTML(talksByDateTime []TalkGroup) string {
	t := template.Must(
		template.
			New("htmlTemplate").
			Funcs(template.FuncMap{
				"unescaped": func(x string) template.HTML {
					return template.HTML(x)
				}}).
			Parse(htmlTemplate))
	var doc bytes.Buffer
	err := t.Execute(&doc, talksByDateTime)
	check(err)
	// doc.WriteTo(os.Stdout)
	return doc.String()
}
