package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/happydez/project-managment/phase-2/assignment/internal/assignment"
)

func main() {
	data := flag.String("data", "tasks.csv", "CSV с трудозатратами и оценками предпочтительности")
	delta := flag.Float64("delta", 1, "допустимое отклонение нагрузки от равной доли, ч")
	veto := flag.Int("veto", 1, "назначения с оценкой не выше этой запрещены")
	links := flag.String("links", "3.7=3.8,3.10=3.11", "пары задач с общим исполнителем")
	flag.Parse()
	log.SetFlags(0)

	p, err := assignment.Load(*data)
	if err != nil {
		log.Fatal(err)
	}
	p.Delta, p.Veto = *delta, *veto
	if err := p.SetLinks(*links); err != nil {
		log.Fatal(err)
	}

	lo, hi := p.LoadBounds()
	fmt.Printf("Задач: %d, участников: %d, трудозатраты: %.1f ч\n", len(p.Tasks), len(p.Experts), p.TotalHours())
	fmt.Printf("Допустимая нагрузка: от %.1f до %.1f ч\n", lo, hi)

	res := assignment.Solve(p)
	if len(res.Optima) == 0 {
		log.Fatal("допустимых распределений нет")
	}
	fmt.Printf("Оптимум: %d, оптимальных распределений: %d, просмотрено узлов: %d\n", res.Score, len(res.Optima), res.Nodes)

	for k, assign := range res.Optima {
		fmt.Printf("\nРаспределение %d\n\n", k+1)
		printTasks(p, assign)
		fmt.Println()
		printExperts(p, assign)
	}
}

func printTasks(p *assignment.Problem, assign []int) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Задача\tНазвание\tЧасы\tИсполнитель\tОценка")
	for j, t := range p.Tasks {
		i := assign[j]
		fmt.Fprintf(w, "%s\t%s\t%.1f\t%s\t%d\n", t.ID, t.Name, t.Hours, p.Experts[i], t.Prefs[i])
	}
}

func printExperts(p *assignment.Problem, assign []int) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Участник\tЗадачи\tЧасы\tСумма оценок\tСредняя оценка")
	for i, name := range p.Experts {
		var ids []string
		hours, score := 0.0, 0
		for j, t := range p.Tasks {
			if assign[j] == i {
				ids = append(ids, t.ID)
				hours += t.Hours
				score += t.Prefs[i]
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%.1f\t%d\t%.2f\n", name, strings.Join(ids, ", "), hours, score, float64(score)/float64(len(ids)))
	}
}
