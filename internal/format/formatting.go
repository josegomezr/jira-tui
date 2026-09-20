package format

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	jira "github.com/andygrunwald/go-jira"
)

func FormatTitle(issue *jira.Issue) string {
	return fmt.Sprintf("[%s] [%s]", issue.Fields.Type.Name, issue.Key)
}

//go:embed content.tpl
var contentTpl []byte

//go:embed meta.tpl
var metaTpl []byte

func makeTplEngine() *template.Template {
	funcs := sprig.FuncMap()
	delete(funcs, "env")
	delete(funcs, "expandenv")
	funcs["fromjiratime"] = func(t jira.Time) time.Time {
		return time.Time(t)
	}
	funcs["jiratomd"] = func(s string) string {
		cvt, err := NewConverter(strings.NewReader(s)).Convert()
		if err != nil {
			return s
		}
		return cvt
	}
	funcs["agoshort"] = func(date any) string {
		// Drop resolution to seconds
		var t time.Time

		switch date := date.(type) {
		default:
			t = time.Now()
		case time.Time:
			t = date
		case int64:
			t = time.Unix(date, 0)
		case int:
			t = time.Unix(int64(date), 0)
		}
		// Drop resolution to seconds
		duration := int(time.Since(t).Round(time.Second).Seconds())

		factors := []struct {
			labels  []string
			divisor int
		}{
			{[]string{"second", "seconds"}, 60},
			{[]string{"minute", "minutes"}, 60},
			{[]string{"hour", "hours"}, 24},
			{[]string{"day", "days"}, 30},
			{[]string{"month", "months"}, 12},
			{[]string{"year", "years"}, 1},
		}

		pieces := []string{}
		for _, tripl := range factors {
			mn := duration
			if tripl.divisor > 0 {
				mn %= tripl.divisor
			}
			if mn > 0 {
				lbl := tripl.labels[0]
				if mn > 1 {
					lbl = tripl.labels[1]
				}

				pieces = append(pieces, fmt.Sprintf("%d %s", mn, lbl))
				duration -= mn
			}
			if tripl.divisor > 0 {
				duration /= tripl.divisor
			}
		}
		min := 0
		if len(pieces)-2 > 0 {
			min = len(pieces) - 2
		}
		return strings.Join(pieces[min:len(pieces)], ", ")
	}

	return template.New("test").Funcs(funcs)
}

func FormatContent(issue *jira.Issue) string {
	bldr := &strings.Builder{}
	engine := makeTplEngine()
	tmpl, err := engine.Parse(string(contentTpl))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(bldr, issue)
	if err != nil {
		panic(err)
	}
	return bldr.String()
}
func FormatMeta(issue *jira.Issue) string {
	bldr := &strings.Builder{}
	engine := makeTplEngine()
	tmpl, err := engine.Parse(string(metaTpl))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(bldr, issue)
	if err != nil {
		panic(err)
	}
	return bldr.String()
}
