package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
	"github.com/sst/opencode/internal/viewport"
)

// OpsMenuItem represents a single ops workflow item
type OpsMenuItem struct {
	Title       string
	Description string
	Action      string // Command to send to the AI
	Keybind     string // Optional keybind hint
}

type opsDialog struct {
	width     int
	height    int
	modal     *modal.Modal
	app       *app.App
	viewport  viewport.Model
	items     []OpsMenuItem
	selected  int
}

func (o *opsDialog) Init() tea.Cmd {
	return o.viewport.Init()
}

func (o *opsDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.width = msg.Width
		o.height = msg.Height
		// Set viewport size with some padding for the modal, but cap at reasonable width
		maxWidth := min(80, msg.Width-8)
		o.viewport = viewport.New(viewport.WithWidth(maxWidth-4), viewport.WithHeight(msg.Height-6))
		
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if o.selected > 0 {
				o.selected--
			}
		case "down", "j":
			if o.selected < len(o.items)-1 {
				o.selected++
			}
		case "enter":
			if o.selected < len(o.items) {
				item := o.items[o.selected]
				// Send the action as a prompt to the AI
				return o, tea.Sequence(
					util.CmdHandler(modal.CloseModalMsg{}),
					util.CmdHandler(app.SendPrompt{Text: item.Action}),
				)
			}
		case "esc":
			return o, util.CmdHandler(modal.CloseModalMsg{})
		}
	}

	// Update viewport content
	o.viewport.SetContent(o.renderContent())

	// Update viewport
	var vpCmd tea.Cmd
	o.viewport, vpCmd = o.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return o, tea.Batch(cmds...)
}

func (o *opsDialog) renderContent() string {
	t := theme.CurrentTheme()
	
	var builder strings.Builder
	
	// Render each menu item
	for i, item := range o.items {
		itemStyle := styles.NewStyle().
			Background(t.BackgroundPanel()).
			Foreground(t.Text()).
			PaddingLeft(2)
			
		if i == o.selected {
			itemStyle = itemStyle.
				Background(t.Primary()).
				Foreground(t.BackgroundPanel()).
				Bold(true)
		}
		
		titleStyle := itemStyle
		descStyle := itemStyle.
			Foreground(t.TextMuted()).
			PaddingLeft(4)
		
		if item.Keybind != "" {
			keybindStyle := itemStyle.
				Foreground(t.TextMuted()).
				PaddingLeft(1)
			builder.WriteString(titleStyle.Render(item.Title))
			builder.WriteString(keybindStyle.Render(" (" + item.Keybind + ")"))
		} else {
			builder.WriteString(titleStyle.Render(item.Title))
		}
		builder.WriteString("\n")
		builder.WriteString(descStyle.Render(item.Description))
		if i < len(o.items)-1 {
			builder.WriteString("\n\n")
		}
	}
	
	return builder.String()
}

func (o *opsDialog) View() string {
	t := theme.CurrentTheme()
	// Add background color to viewport
	bg := t.BackgroundPanel()
	content := o.viewport.View()
	
	// Apply background to entire viewport area
	return styles.NewStyle().
		Background(bg).
		Width(o.viewport.Width()).
		Height(o.viewport.Height()).
		Render(content)
}

func (o *opsDialog) Render(background string) string {
	return o.modal.Render(o.View(), background)
}

func (o *opsDialog) Close() tea.Cmd {
	return nil
}

type OpsDialog interface {
	layout.Modal
}

func NewOpsDialog(app *app.App) OpsDialog {
	vp := viewport.New(viewport.WithHeight(12))
	
	// Define ops workflow menu items
	items := []OpsMenuItem{
		{
			Title:       "Check Alerts",
			Description: "Review current alerts and anomalies in AppSignals",
			Action:      "Check my AppSignals alerts and summarize any current issues or anomalies",
		},
		{
			Title:       "Service Health",
			Description: "Get overview of service health and performance",
			Action:      "Show me the current health status and performance metrics for all my services",
		},
		{
			Title:       "Recent Deployments",
			Description: "View recent deployments and their status",
			Action:      "List recent deployments and their current status, including any issues",
		},
		{
			Title:       "Error Analysis",
			Description: "Analyze recent errors and exceptions",
			Action:      "Analyze recent errors and exceptions across services, group by type and show trends",
		},
		{
			Title:       "Performance Insights",
			Description: "Get performance insights and recommendations",
			Action:      "Provide performance insights and optimization recommendations based on current metrics",
		},
		{
			Title:       "Cost Analysis",
			Description: "Review AWS costs and optimization opportunities",
			Action:      "Analyze current AWS costs and suggest optimization opportunities",
		},
		{
			Title:       "Security Scan",
			Description: "Check for security issues and vulnerabilities",
			Action:      "Run a security scan and check for any vulnerabilities or misconfigurations",
		},
		{
			Title:       "Incident Report",
			Description: "Generate incident report template",
			Action:      "Generate an incident report template for the most recent issue",
		},
	}
	
	return &opsDialog{
		app:      app,
		modal:    modal.New(modal.WithTitle("Ops Menu"), modal.WithMaxWidth(80)),
		viewport: vp,
		items:    items,
		selected: 0,
	}
}