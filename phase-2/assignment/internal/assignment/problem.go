package assignment

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	ID    string
	Name  string
	Hours float64
	Prefs []int
}

type Problem struct {
	Experts []string
	Tasks   []Task
	Delta   float64  // допустимое отклонение нагрузки от равной доли, ч
	Veto    int      // назначения с оценкой не выше Veto запрещены
	Links   [][2]int // пары задач с общим исполнителем
}

func (p *Problem) TotalHours() float64 {
	total := 0.0
	for _, t := range p.Tasks {
		total += t.Hours
	}
	return total
}

func (p *Problem) LoadBounds() (lo, hi float64) {
	share := p.TotalHours() / float64(len(p.Experts))
	return share - p.Delta, share + p.Delta
}

// Load читает CSV со столбцами task, name, hours и оценками участников.
func Load(path string) (*Problem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 || len(rows[0]) < 4 {
		return nil, fmt.Errorf("%s: ожидаются столбцы task,name,hours и оценки участников", path)
	}

	p := &Problem{Experts: rows[0][3:]}
	for i, row := range rows[1:] {
		t := Task{ID: row[0], Name: row[1]}
		if t.Hours, err = strconv.ParseFloat(row[2], 64); err != nil {
			return nil, fmt.Errorf("%s, строка %d: %w", path, i+2, err)
		}
		for _, cell := range row[3:] {
			pref, err := strconv.Atoi(cell)
			if err != nil {
				return nil, fmt.Errorf("%s, строка %d: %w", path, i+2, err)
			}
			t.Prefs = append(t.Prefs, pref)
		}
		p.Tasks = append(p.Tasks, t)
	}
	return p, nil
}

// SetLinks принимает пары задач в виде "3.7=3.8,3.10=3.11".
func (p *Problem) SetLinks(spec string) error {
	index := make(map[string]int, len(p.Tasks))
	for j, t := range p.Tasks {
		index[t.ID] = j
	}

	p.Links = nil
	for _, pair := range strings.FieldsFunc(spec, func(r rune) bool { return r == ',' }) {
		a, b, _ := strings.Cut(pair, "=")
		ja, okA := index[a]
		jb, okB := index[b]
		if !okA || !okB {
			return fmt.Errorf("неизвестная задача в связи %q", pair)
		}
		p.Links = append(p.Links, [2]int{ja, jb})
	}
	return nil
}
