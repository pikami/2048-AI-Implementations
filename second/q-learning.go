package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tebeka/selenium"
)

const (
	Up int = iota
	Down
	Left
	Right

	grid_size = 16
)

type QLearningTD struct {
	// Q Table
	Q  [][]float64
	nC int // Node count

	// Action count
	Qn int

	// Goal
	goal int

	// Current state
	state [16]int

	α float64
	ε float64
	γ float64
}

func (q *QLearningTD) Initialize() {

	q.α = 0.5
	q.ε = 0.1
	q.γ = 1

	q.state = [grid_size]int{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0}

	Actions := 4
	q.Qn = Actions
	q.nC = int(math.Pow(2, 15))
	q.goal = 128

	q.Q = make([][]float64, Actions)

	for i := 0; i < Actions; i++ {
		q.Q[i] = make([]float64, q.nC)
	}

	for i := 0; i < Actions; i++ {
		for j := 0; j < q.nC; j++ {
			q.Q[i][j] = rand.Float64()
		}
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	Q := QLearningTD{}
	Q.Initialize()
	Q.Start()
}

type Agent struct {
	State   [grid_size]int
	Wd      selenium.WebDriver
	Service *selenium.Service
}

func (q *QLearningTD) Start() {
	// WD
	var agents [4]Agent
	for index, _ := range agents {
		agents[index].Service, agents[index].Wd = getWD()
		defer agents[index].Wd.Quit()
		defer agents[index].Service.Stop()
		if err := agents[index].Wd.Get("https://4ark.me/2048/"); err != nil {
			panic(err)
		}
	}
	time.Sleep(2000 * time.Millisecond)
	for index, _ := range agents {
		agents[index].State = getGrid(agents[index].Wd)
	}
	// AI
	for i := 0; i < 1000; i++ {
		for _, agent := range agents {
			agent.step(q)
		}
	}
	/*episodes := 1000
	for i := 0; i < episodes; i++ {
		restartGame(wd)
		state := getGrid(wd)
		ep := 0

		for !contains(state, q.goal) {
			ep++
			Action := q.ε_greedy(state)
			r, _state := q.TakeAction(Action, state, wd)
			QSA := q.GetQ(state, Action)
			MaxAction := q.GetAction(_state)
			_QSA := q.GetQ(_state, MaxAction)

			Q := QSA + q.α*(r+q.γ*_QSA-QSA)
			q.SetQ(state, Action, Q)

			state = _state

			if didLose(wd) {
				break
			}
		}
	}*/
}

func (a *Agent) step(q *QLearningTD) {
	if !contains(a.State, q.goal) && !didLose(a.Wd) {
		Action := q.ε_greedy(a.State)
		r, _state := q.TakeAction(Action, a.State, a.Wd)
		QSA := q.GetQ(a.State, Action)
		MaxAction := q.GetAction(_state)
		_QSA := q.GetQ(_state, MaxAction)

		Q := QSA + q.α*(r+q.γ*_QSA-QSA)
		q.SetQ(a.State, Action, Q)

		a.State = _state
	} else {
		restartGame(a.Wd)
		a.State = getGrid(a.Wd)
	}
}

func printDebug(wd selenium.WebDriver) {
	fmt.Printf("didLose: %t\n", didLose(wd))
}

func (q *QLearningTD) GetAction(state [grid_size]int) int {

	Idx := q.GetIdx(state)
	max := q.Q[0][Idx]
	Action := 0
	for i := 1; i < q.Qn; i++ {
		if max < q.Q[i][Idx] {
			max = q.Q[i][Idx]
			Action = i
		}
	}

	return Action
}

func (q *QLearningTD) ε_greedy(state [grid_size]int) int {

	Action := q.GetAction(state)

	if rand.Float64() < 1-q.ε {
		return Action
	}

	return rand.Intn(q.Qn)

}

func (q *QLearningTD) TakeAction(a int, state [grid_size]int, wd selenium.WebDriver) (float64, [grid_size]int) {

	var _state [grid_size]int
	for index, value := range state {
		_state[index] = value
	}

	switch a {
	case Up:
		sendKey(wd, upArrowKey)
	case Down:
		sendKey(wd, downArrowKey)
	case Left:
		sendKey(wd, leftArrowKey)
	case Right:
		sendKey(wd, rightArrowKey)
	}

	_state = getGrid(wd)

	r := mergedCellsValue(state, _state)
	if getLargest(_state) == q.goal {
		r = 1
	} else if !didMove(state, _state) {
		r = -0.3
	} else if didLose(wd) {
		r = -1
	}
	return r, _state
}

func mergedCellsValue(prev [grid_size]int, cur [grid_size]int) float64 {
	return float64(sliceSum(cur)-sliceSum(prev)) / 8.0
}

func didMove(prev [grid_size]int, cur [grid_size]int) bool {
	for index, _ := range prev {
		if prev[index] == cur[index] {
			return true
		}
	}
	return false
}

func getLargest(grid [grid_size]int) int {
	m := 0
	for i, e := range grid {
		if i == 0 || e > m {
			m = e
		}
	}
	return m
}

func sliceSum(nums [grid_size]int) int {
	sum := 0
	for _, num := range nums {
		sum += num
	}
	return sum
}

func (q *QLearningTD) GetQ(state [grid_size]int, a int) float64 {

	return q.Q[a][q.GetIdx(state)]
}
func (q *QLearningTD) SetQ(state [grid_size]int, a int, f float64) {
	q.Q[a][q.GetIdx(state)] = f
}

func (q *QLearningTD) GetIdx(state [grid_size]int) int {
	sum := 0
	for index, cell := range state {
		sum += int(math.Pow10(index)) * cell
	}
	return sum % q.nC
}

func contains(s [grid_size]int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
