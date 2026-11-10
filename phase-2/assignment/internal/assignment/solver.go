package assignment

type Result struct {
	Score  int
	Optima [][]int // Optima[k][j] - индекс исполнителя задачи j в k-м оптимальном распределении
	Nodes  int
}

type search struct {
	p         *Problem
	lo, hi    float64
	sameAs    []int     // более ранняя задача с тем же исполнителем или -1
	restScore []int     // верхняя оценка суммы по задачам j и дальше
	restHours []float64 // трудозатраты задач j и дальше
	assign    []int
	load      []float64
	res       Result
}

// Solve ищет все оптимальные распределения методом ветвей и границ.
// Если допустимых распределений нет, Result.Optima пуст.
func Solve(p *Problem) Result {
	m := len(p.Tasks)
	s := &search{
		p:         p,
		sameAs:    make([]int, m),
		restScore: make([]int, m+1),
		restHours: make([]float64, m+1),
		assign:    make([]int, m),
		load:      make([]float64, len(p.Experts)),
		res:       Result{Score: -1},
	}
	s.lo, s.hi = p.LoadBounds()

	for j := range s.sameAs {
		s.sameAs[j] = -1
	}
	for _, l := range p.Links {
		s.sameAs[max(l[0], l[1])] = min(l[0], l[1])
	}
	for j := m - 1; j >= 0; j-- {
		best := 0
		for _, pref := range p.Tasks[j].Prefs {
			if pref > p.Veto {
				best = max(best, pref)
			}
		}
		s.restScore[j] = s.restScore[j+1] + best
		s.restHours[j] = s.restHours[j+1] + p.Tasks[j].Hours
	}

	s.branch(0, 0)
	return s.res
}

func (s *search) branch(j, score int) {
	s.res.Nodes++
	if score+s.restScore[j] < s.res.Score || s.shortage() > s.restHours[j] {
		return
	}
	if j == len(s.p.Tasks) {
		if score > s.res.Score {
			s.res.Score, s.res.Optima = score, nil
		}
		s.res.Optima = append(s.res.Optima, append([]int(nil), s.assign...))
		return
	}

	t := s.p.Tasks[j]
	for i := range s.p.Experts {
		if t.Prefs[i] <= s.p.Veto {
			continue
		}
		if k := s.sameAs[j]; k >= 0 && s.assign[k] != i {
			continue
		}
		if s.load[i]+t.Hours > s.hi {
			continue
		}
		s.assign[j] = i
		s.load[i] += t.Hours
		s.branch(j+1, score+t.Prefs[i])
		s.load[i] -= t.Hours
	}
}

// shortage - сколько часов не хватает участникам до нижней границы нагрузки.
func (s *search) shortage() float64 {
	total := 0.0
	for _, h := range s.load {
		total += max(0, s.lo-h)
	}
	return total
}
