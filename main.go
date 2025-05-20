package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TrafficLightState int

const (
	Red TrafficLightState = iota
	Green
	Yellow
)

func (s TrafficLightState) String() string {
	switch s {
	case Red:
		return "🔴 Vermelho"
	case Green:
		return "🟢 Verde"
	case Yellow:
		return "🟡 Amarelo"
	default:
		return "❓ Desconhecido"
	}
}

func (s TrafficLightState) ColorChar() string {
	switch s {
	case Red:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("🟥")
	case Green:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("🟩")
	case Yellow:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("🟨")
	default:
		return "⬛"
	}
}

type TrafficLight struct {
	Avenue    string
	State     TrafficLightState
	Timer     int
	Group     int
	Passed    int
	Waiting   int
	Countdown int
}

type Vehicle struct {
	Avenue string
	Pos    int
	Speed  int
	ID     int
}

type model struct {
	lights        []TrafficLight
	frame         int
	vehicles      []Vehicle
	width         int
	cycle         int
	syncCountdown int // tempo até próximo grupo
	transitioning bool
}

type tickMsg time.Time

func initialModel() model {
	return model{
		lights: []TrafficLight{
			{"Av. Brasil", Red, 0, 1, 0, 0, 0},
			{"Av. Paulista", Red, 0, 2, 0, 0, 0},
			{"Av. Rebouças", Red, 0, 1, 0, 0, 0},
		},
		vehicles:      []Vehicle{},
		width:         50,
		cycle:         1,
		syncCountdown: 2, // atraso entre trocas de grupo
		transitioning: true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.frame++

		// controle de grupo sincronizado
		if m.transitioning {
			m.syncCountdown--
			if m.syncCountdown <= 0 {
				m.transitioning = false
				for i := range m.lights {
					if m.lights[i].Group == m.cycle {
						m.lights[i].State = Green
						m.lights[i].Countdown = 5
					}
				}
			}
		} else {
			allExpired := true
			for i := range m.lights {
				light := &m.lights[i]
				if light.Group == m.cycle {
					if light.Countdown > 0 {
						light.Countdown--
						allExpired = false
					} else {
						light.State = Yellow
						light.Countdown = 2
					}
				} else if light.State != Red {
					light.Countdown--
					if light.Countdown <= 0 {
						light.State = Red
						light.Countdown = 0
					}
				}
			}
			if allExpired {
				m.transitioning = true
				m.syncCountdown = 2
				m.cycle = 3 - m.cycle
			}
		}

		if rand.Intn(10) > 6 {
			avenues := []string{"Av. Brasil", "Av. Paulista", "Av. Rebouças"}
			v := Vehicle{
				Avenue: avenues[rand.Intn(len(avenues))],
				Pos:    0,
				Speed:  rand.Intn(2) + 1,
				ID:     rand.Intn(999),
			}
			m.vehicles = append(m.vehicles, v)
		}

		active := []Vehicle{}
		for _, v := range m.vehicles {
			light := getLightByAvenue(m.lights, v.Avenue)
			if light != nil && light.State == Green {
				v.Pos += v.Speed
			}
			if v.Pos < m.width {
				active = append(active, v)
			}
		}
		m.vehicles = active

		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func getLightByAvenue(lights []TrafficLight, avenue string) *TrafficLight {
	for i := range lights {
		if lights[i].Avenue == avenue {
			return &lights[i]
		}
	}
	return nil
}

func (m model) renderIntersectionView() string {
	var horizontal, vertical string
	for _, l := range m.lights {
		if l.Group == 1 {
			horizontal = l.State.ColorChar()
		} else if l.Group == 2 {
			vertical = l.State.ColorChar()
		}
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("     %s     \n", vertical))
	sb.WriteString("     │     \n")
	sb.WriteString(fmt.Sprintf("%s───┼───%s\n", horizontal, horizontal))
	sb.WriteString("     │     \n")
	sb.WriteString(fmt.Sprintf("     %s     ", vertical))

	return sb.String()
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("🚦 Simulador de Trânsito com Fluxo Coordenado e Timer")
	output := title + "\n\n"

	output += "== Painel de Avenidas ==\n"
	for _, light := range m.lights {
		output += fmt.Sprintf("%s (Grupo %d): %s (%ds restantes)\n", light.Avenue, light.Group, light.State, light.Countdown)
		line := make([]rune, m.width)
		for i := range line {
			line[i] = ' '
		}
		for _, v := range m.vehicles {
			if v.Avenue == light.Avenue && v.Pos < m.width {
				line[v.Pos] = '🚗'
			}
		}
		output += fmt.Sprintf("%s\n\n", string(line))
	}

	output += "\n== Cruzamento Central ==\n"
	output += m.renderIntersectionView()
	output += "\n\nPressione 'q' para sair."
	return output
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Erro ao iniciar o programa:", err)
	}
}
