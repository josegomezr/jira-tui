package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"fmt"
	"os"
	"context"
	"time"
	_ "embed"
	"strings"
	"text/template"

	"charm.land/bubbles/v2/viewport"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Masterminds/sprig/v3"
	jira "github.com/andygrunwald/go-jira"
)

var (
	titleStyle = func() lipgloss.Style {
		return lipgloss.NewStyle()
	}()

	infoStyle = func() lipgloss.Style {
		return titleStyle
	}()
)

type model struct {
	content  string
	meta  string
	title  string
	ready    bool
	state int
	width int
	
	currquery string
	viewport viewport.Model
	leftvp viewport.Model
	query textinput.Model
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	ctx, _ := context.WithCancel(context.Background())
	cmds = append(cmds, m.executeQueryCommand(ctx, m.currquery))
	return tea.Sequence(cmds...)
}

func(m *model) writeContents(){
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()-2).Render(m.content))
	m.leftvp.SetContent(lipgloss.NewStyle().Width(m.leftvp.Width()-2).Render(m.meta))
}

func(m *model) draw(width, height int){
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	m.width = width

	if !m.ready {
		m.query = textinput.New()
		m.query.SetWidth(20)
		m.query.Prompt = ": "
		m.query.Placeholder = "Ticket: FFF-123"
		m.query.Validate = func(s string) error {
			if !strings.Contains(s, "-"){
				return fmt.Errorf("Not a ticket")
			}
			return nil
		}

		q := m.query.Styles()
		q.Cursor.Blink = true
		m.query.SetStyles(q)

		m.viewport = viewport.New(viewport.WithWidth(width*2/3), viewport.WithHeight(height-verticalMarginHeight))
		m.viewport.YPosition = headerHeight
	
		m.viewport.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)

		m.leftvp = viewport.New(viewport.WithWidth(width - m.viewport.Width()), viewport.WithHeight(height-verticalMarginHeight))
		m.leftvp.YPosition = headerHeight
		m.leftvp.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
		
		m.ready = true
	} else {
		m.viewport.SetWidth(width*2/3)
		m.viewport.SetHeight(height - verticalMarginHeight)

		m.leftvp.SetWidth(m.viewport.Width())
		m.leftvp.SetHeight(height - verticalMarginHeight)
	}
}

type issueResult struct {
	search string
	title string
	content string
	meta string
}

type initial struct {}
type loading struct {}
type errFetch struct {
	val string
	err error
}

func (m *model) executeQueryCommand(ctx context.Context, val string) tea.Cmd {
	return func() tea.Msg {
		if val == "" {
			return initial{}
		}
		issue, _, err := jiraClient.Issue.Get(val, nil)
		if err != nil {
			return errFetch{
				val: val,
				err: err,
			}
		}

		return issueResult{
			search: val,
			title: formatTitle(issue),
			content: formatContent(issue),
			meta: formatMeta(issue),
		}
	}
}

func formatTitle(issue *jira.Issue) string {
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

		factors := []struct{
			labels []string
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
		for _, tripl := range factors{
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
			if tripl.divisor > 0{
				duration /= tripl.divisor
			}
		}
		min := 0
		if len(pieces)-2 > 0 {
			min = len(pieces)-2
		}
		return strings.Join(pieces[min:len(pieces)], ", ")
	}

	return template.New("test").Funcs(funcs)
}

func formatContent(issue *jira.Issue) string {
	bldr := &strings.Builder{}
	engine := makeTplEngine()
	tmpl, err := engine.Parse(string(contentTpl))
	if err != nil { panic(err) }
	err = tmpl.Execute(bldr, issue)
	if err != nil { panic(err) }
	return bldr.String()
}
func formatMeta(issue *jira.Issue) string {
	bldr := &strings.Builder{}
	engine := makeTplEngine()
	tmpl, err := engine.Parse(string(metaTpl))
	if err != nil { panic(err) }
	err = tmpl.Execute(bldr, issue)
	if err != nil { panic(err) }
	return bldr.String()
}

func (m *model) handleMessage(msg tea.Msg, cmds *[]tea.Cmd) {
	switch msg := msg.(type) {
	case issueResult:
		m.currquery = msg.search
		m.meta = msg.meta
		m.title = msg.title
		m.content = msg.content
		m.state = 1
	case errFetch:
		m.meta = ""
		m.title = ""
		m.content = fmt.Sprintf("err, that's not an issue...: %s: %s", msg.val, msg.err)
	case initial:
		m.meta = ""
		m.title = ""
		m.content = "Specify an issue..."
	case loading:
		m.meta = ""
		m.title = ""
		m.content = "Loading..."
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			*cmds = append(*cmds, tea.Quit)
			return
		case "enter":
			if val := m.query.Value(); val != m.currquery {
				ctx, _ := context.WithCancel(context.Background())
				*cmds = append(*cmds, func() tea.Msg { return loading{}})
				*cmds = append(*cmds, m.executeQueryCommand(ctx, val))
			}
		case "esc":
			m.state = 0
		case "tab":
			m.state += 1
			if m.state > 2 {
				m.state = 2
			}
		case "shift+tab":
			m.state -= 1
			if m.state < 0 {
				m.state = 0
			}
		}
	case tea.WindowSizeMsg:
		m.draw(msg.Width, msg.Height)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	m.handleMessage(msg, &cmds)
	m.writeContents()

	if m.ready {
		if m.state == 0 {
			m.query.Focus()
		}else{
			m.query.Blur()
		}
	}

	if m.state == 1 {
		m.viewport.Style = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder(), false, true)
	}else{
		m.viewport.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
	}

	if m.state == 2 {
		m.leftvp.Style = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder(), false, true)
	}else{
		m.leftvp.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
	}

	if m.state == 0 {
		q, cmd := m.query.Update(msg)
		m.query = q
		cmds = append(cmds, cmd)
	}

	if m.state == 1 {
		vp, cmd := m.viewport.Update(msg)
		m.viewport = vp
		cmds = append(cmds, cmd)
	}

	if m.state == 2 {
		vpr, cmd := m.leftvp.Update(msg)
		m.leftvp = vpr
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true                    // use the full size of the terminal in its "alternate screen buffer"
	v.MouseMode = tea.MouseModeCellMotion // turn on mouse support so we can track the mouse wheel
	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), lipgloss.JoinHorizontal(lipgloss.Top, m.viewport.View(), m.leftvp.View()), m.footerView()))
	}
	return v
}

func (m model) headerView() string {
	title := titleStyle.Render(m.title)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render("")
	if m.state == 1 {
		info = infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	}
	if m.state == 2 {
		info = infoStyle.Render(fmt.Sprintf("%3.f%%", m.leftvp.ScrollPercent()*100))
	}
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(info)-lipgloss.Width(m.query.View())))
	return lipgloss.JoinHorizontal(lipgloss.Center, m.query.View(), line, info)
}


var jiraClient *jira.Client
func main() {
	tp := jira.PATAuthTransport{
		Token: os.Getenv("PAT"),
	}
	if tp.Token == ""{
		fmt.Println("Please provide a jira PAT via the PAT env var")
		os.Exit(1)
	}
	
	instance := os.Getenv("JIRA")
	if instance == "" {
		fmt.Println("Please provide a jira instance via the JIRA env var")
		os.Exit(1)
	}

	c, _ := jira.NewClient(tp.Client(), fmt.Sprintf("https://%s", instance))
	jiraClient = c

	// Load some text for our viewport
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	p := tea.NewProgram(
		model{
			currquery: arg,
			meta: "",
			title: "",
			content: "Specify a ticket...",
			state: 0,
		},
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
