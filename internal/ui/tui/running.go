package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/viper"
)

type RunningModel struct {
	config   RunConfig
	spinner  spinner.Model
	viewport viewport.Model
	logs     []string
	running  bool
	done     bool
	cancel   context.CancelFunc
	width    int
	height   int
}

func NewRunningModel() RunningModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = successStyle

	return RunningModel{
		spinner: s,
		logs:    []string{},
	}
}

func (m RunningModel) SetConfig(config RunConfig) RunningModel {
	m.config = config
	return m
}

func (m RunningModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startExecution(),
	)
}

func (m RunningModel) Update(msg tea.Msg) (RunningModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+x":
			if m.cancel != nil {
				m.cancel()
				m.logs = append(m.logs, "Cancelling...")
			}
		case "esc", "enter":
			if m.done {
				return m, func() tea.Msg { return NavigateMsg{To: ScreenHome} }
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10

	case spinner.TickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case logMsg:
		m.logs = append(m.logs, string(msg))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case GenerateCompleteMsg:
		m.done = true
		m.running = false
		return m, func() tea.Msg { return msg }

	case AnalyzeCompleteMsg:
		m.done = true
		m.running = false
		return m, func() tea.Msg { return msg }
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m RunningModel) View() string {
	var b strings.Builder

	if m.config.Mode == "generate" {
		b.WriteString(titleStyle.Render("⚡ Generating Tests"))
	} else {
		b.WriteString(titleStyle.Render("📊 Analyzing Codebase"))
	}
	b.WriteString("\n\n")

	if !m.done {
		b.WriteString(fmt.Sprintf("%s Running...\n\n", m.spinner.View()))
	} else {
		b.WriteString(successStyle.Render("✔ Complete"))
		b.WriteString("\n\n")
	}

	// Logs viewport
	b.WriteString(boxStyle.Render(strings.Join(m.logs, "\n")))
	b.WriteString("\n\n")

	if m.done {
		b.WriteString(helpStyle.Render("enter: continue • esc: home"))
	} else {
		b.WriteString(helpStyle.Render("ctrl+x: cancel"))
	}

	return b.String()
}

type logMsg string

func (m RunningModel) startExecution() tea.Cmd {
	return func() tea.Msg {
		if m.config.Mode == "generate" {
			return m.runGenerate()
		}
		return m.runAnalyze()
	}
}

func (m *RunningModel) runGenerate() tea.Msg {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	defer cancel()

	service := app.NewService()
	response, err := service.Generate(ctx, app.GenerateRequest{
		Path:        m.config.Path,
		File:        m.config.File,
		Recursive:   m.config.Recursive,
		TestTypes:   m.config.Types,
		DryRun:      m.config.DryRun,
		Validate:    m.config.Validate,
		Parallelism: m.config.Parallel,
		Provider:    viper.GetString("llm.provider"),
	})
	if err != nil {
		return GenerateCompleteMsg{Err: err}
	}

	return GenerateCompleteMsg{Results: response.Results}
}

func (m *RunningModel) runAnalyze() tea.Msg {
	service := app.NewService()
	result, err := service.Analyze(context.Background(), app.AnalyzeRequest{
		Path:         m.config.Path,
		Recursive:    m.config.Recursive,
		CostEstimate: m.config.CostEst,
		Detail:       m.config.Detail,
	})
	if err != nil {
		return AnalyzeCompleteMsg{Err: err}
	}

	return AnalyzeCompleteMsg{Result: result}
}
