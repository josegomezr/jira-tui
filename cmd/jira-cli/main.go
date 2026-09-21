package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	jira "github.com/andygrunwald/go-jira"
	"github.com/josegomezr/jira-tui/internal/format"
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
	content string
	meta    string
	title   string
	ready   bool
	state   int
	width   int
	focus   bool

	currquery    string
	contentpanel viewport.Model
	metapanel    viewport.Model
	query        textinput.Model
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	ctx, _ := context.WithCancel(context.Background())
	cmds = append(cmds, m.executeQueryCommand(ctx, m.currquery))
	return tea.Sequence(cmds...)
}

func (m *model) writeContents() {
	m.contentpanel.SetContent(lipgloss.NewStyle().Render(
		lipgloss.Wrap(m.content, max(0, m.contentpanel.Width()- 2), " "),
	))
	m.metapanel.SetContent(lipgloss.NewStyle().Width(max(m.metapanel.Width() - 2, 0)).Render(m.meta))
}

func (m *model) draw(width, height int) {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight
	m.width = width

	if !m.ready {
		m.query = textinput.New()
		m.query.SetWidth(20)
		m.query.SetValue(m.currquery)
		m.query.Prompt = ": "
		m.query.Placeholder = "Ticket: FFF-123"
		m.query.Validate = func(s string) error {
			if !strings.Contains(s, "-") {
				return fmt.Errorf("Not a ticket")
			}
			return nil
		}

		q := m.query.Styles()
		q.Cursor.Blink = true
		m.query.SetStyles(q)

		m.contentpanel = viewport.New(viewport.WithWidth(width*2/3), viewport.WithHeight(height-verticalMarginHeight))
		m.contentpanel.YPosition = headerHeight

		m.contentpanel.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)

		m.metapanel = viewport.New(viewport.WithWidth(width-m.contentpanel.Width()), viewport.WithHeight(height-verticalMarginHeight))
		m.metapanel.YPosition = headerHeight
		m.metapanel.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)

		m.ready = true
	} else {
		m.contentpanel.SetWidth(width * 2 / 3)
		m.contentpanel.SetHeight(height - verticalMarginHeight)

		m.metapanel.SetWidth(width - m.contentpanel.Width())
		m.metapanel.SetHeight(height - verticalMarginHeight)
	}
}

type issueResult struct {
	search  string
	title   string
	content string
	meta    string
}

type initial struct{}
type loading struct{}
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
			search:  val,
			title:   format.FormatTitle(issue),
			content: format.FormatContent(issue, true),
			meta:    format.FormatMeta(issue),
		}
	}
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
				*cmds = append(*cmds, func() tea.Msg { return loading{} })
				*cmds = append(*cmds, m.executeQueryCommand(ctx, val))
			}
		case "esc":
			m.state = 0
		case "f":
			m.focus = !m.focus
			if m.focus {
				m.contentpanel.SetWidth(m.width)
				m.metapanel.SetWidth(0)
			}else{
				m.contentpanel.SetWidth(m.width*2/3)
				m.metapanel.SetWidth(m.width - m.contentpanel.Width())
			}
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
		} else {
			m.query.Blur()
		}
	}

	if m.state == 1 {
		m.contentpanel.Style = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder(), false, true)
	} else {
		m.contentpanel.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
	}

	if m.state == 2 {
		m.metapanel.Style = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder(), false, true)
	} else {
		m.metapanel.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
	}

	if m.state == 0 {
		q, cmd := m.query.Update(msg)
		m.query = q
		cmds = append(cmds, cmd)
	}

	if m.state == 1 {
		vp, cmd := m.contentpanel.Update(msg)
		m.contentpanel = vp
		cmds = append(cmds, cmd)
	}

	if m.state == 2 {
		vpr, cmd := m.metapanel.Update(msg)
		m.metapanel = vpr
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true // use the full size of the terminal in its "alternate screen buffer"
	// v.MouseMode = tea.MouseModeCellMotion
	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), lipgloss.JoinHorizontal(lipgloss.Top, m.contentpanel.View(), m.metapanel.View()), m.footerView()))
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
		info = infoStyle.Render(fmt.Sprintf("%3.f%%", m.contentpanel.ScrollPercent()*100))
	}
	if m.state == 2 {
		info = infoStyle.Render(fmt.Sprintf("%3.f%%", m.metapanel.ScrollPercent()*100))
	}
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(info)-lipgloss.Width(m.query.View())))
	return lipgloss.JoinHorizontal(lipgloss.Center, m.query.View(), line, info)
}

var jiraClient *jira.Client

func main() {
	tp := jira.PATAuthTransport{
		Token: os.Getenv("PAT"),
	}
	if tp.Token == "" {
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

	// Load some text for our contentpanel
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	if os.Args[len(os.Args)-1] == "--just-print-it" && arg != "" {
		// output to STDOUT
		issue, _, err := jiraClient.Issue.Get(arg, nil)
		if err != nil {
			fmt.Println("Error fetching issue:", arg, err)
			os.Exit(1)
		}

		fmt.Println(format.FormatTitle(issue), issue.Fields.Summary)
		fmt.Println(format.FormatMeta(issue))
		fmt.Println(format.FormatContent(issue, false))
		os.Exit(0)
	}

	p := tea.NewProgram(
		model{
			currquery: arg,
			meta:      "",
			title:     "",
			content:   "Specify a ticket...",
			state:     0,
		},
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
