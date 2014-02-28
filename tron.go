// Codingame - Tron Battle - http://www.codingame.com/cg/#!challenge:20
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	MAX_X              = 30
	MAX_Y              = 20
	MAX_XY             = 30
	DEBUG              = false
	TIME               = false
	VORONOI_OVER_SPACE = 1.5
	UP                 = "UP"
	DOWN               = "DOWN"
	LEFT               = "LEFT"
	RIGHT              = "RIGHT"
	DIE                = "DIE" //when there is nowhere else to go
)

var (
	b            *board       //Board with the information about the player positions
	v            *board       //Association of the squares with the nearest player (Voronoi diagram)
	voronoiSizes []int        //Number of squares each player has
	playersHeads []coordinate //Coordinates of the heads of each of the oponents
	whoami       int          //ID of the player I play
	alone        bool         //True iff this player is alone in his zone (no possible interactions unless a beam is removed)
	playId       int          //Number of steps so far
)

//Tries to win Tron games
func main() {
	b = newBoard()
	if len(os.Args) == 2 {
		readBoard(os.Args[1])
	}
	for i := 0; i < MAX_X*MAX_Y; i += 1 { //Just to avoid infinite loop
		t0 := time.Now()
		readInput()
		fmt.Println(printMovement())
		t1 := time.Now()
		if DEBUG || TIME {
			fmt.Printf("The call took %v to run.\n", t1.Sub(t0))
		}
	}
}

//Creates and initializes a board
func newBoard() *board {
	result := new(board)
	for i, _ := range result {
		for j, _ := range result[i] {
			result[i][j] = -1
		}
	}
	return result
}

//Returns the best calculated move
func printMovement() (result string) {
	if len(playersHeads) > 1 {
		if DEBUG {
			fmt.Println("CURRENT Board:")
			fmt.Println(b)
		}
		computeSituation()
		//if DEBUG {
		//	fmt.Println("Voronoi:")
		//	fmt.Println(v)
		//	fmt.Println("Board:")
		//	fmt.Println(b)
		//	fmt.Println("Am I alone?", alone)
		//}
		result = doMove()
	} else {
		fmt.Println("Skipping step. I am alone. Nothing to think about...")
		fmt.Println(b)
	}
	return
}

//Analizes current situation
func computeSituation() {
	computeVoronoi()
}

//Computes the space around me. It is calculated in the same way as the voronoi board, but only me is moving
func computeSpace() {
	v = newBoard()
	var stepper chan bool = make(chan bool)
	var notifier chan bool = make(chan bool)
	go computeVoronoiPerPlayer(playersHeads[whoami].x, playersHeads[whoami].y, whoami, stepper, notifier)
	var moreToReceive bool = true
	for moreToReceive {
		stepper <- true
		_, moreToReceive = <-notifier
	}
}

//Computes the distance to the nearest oponent. It is calculated in the same way as the voronoi board, only me is moving and when I find the head of another player, I stop.
func computeDistance() int {
	v = newBoard()
	return computeDistanceFromPlayer(playersHeads[whoami].x, playersHeads[whoami].y, whoami)
}

//Computes voronoi board
func computeVoronoi() {
	v = newBoard()
	var steppers []chan bool = make([]chan bool, len(playersHeads))
	var notifiers []chan bool = make([]chan bool, len(playersHeads))
	for i := 0; i < len(playersHeads); i += 1 {
		steppers[i] = make(chan bool)
		notifiers[i] = make(chan bool)
		go computeVoronoiPerPlayer(playersHeads[i].x, playersHeads[i].y, i, steppers[i], notifiers[i])
	}
	notifications := make([]bool, len(playersHeads))
	for i := 0; i < len(notifications); i += 1 { //for the first loop to begin
		notifications[i] = true
	}
	for !allFalse(notifications) {
		for i := 0; i < len(steppers); i += 1 {
			if notifications[i] {
				steppers[i] <- true
				_, notifications[i] = <-notifiers[i]
			}
		}
	}
}

//Computes the voronoi diagram from given point
func computeDistanceFromPlayer(x, y, i int) int {
	voronoiSizes[i] = 0
	v[x][y] = i
	var printed bool = true
	possibilities := make(map[coordinate]bool)
	possibilities[coordinate{x, y}] = true
	voronoiSizes[i] += len(possibilities)
	result := 1
	for printed {
		result += 1
		printed = false
		possibilities = b.possiblePositionsFromArray(possibilities)
		for pos, _ := range possibilities {
			for n, head := range playersHeads {
				if n != whoami && (abs(pos.x-head.x)+abs(pos.y-head.y)) == 1 {
					return result
				}
			}
			if v[pos.x][pos.y] == -1 {
				v[pos.x][pos.y] = i
				printed = true
			}
		}
	}
	return 1
}

//Returns the absolute value of an int
func abs(i int) int {
	if i < 0 {
		return (-1) * i
	}
	return i
}

//Computes the voronoi diagram from given point
func computeVoronoiPerPlayer(x, y, i int, stepper, notifier chan bool) {
	if i == whoami {
		alone = true
	}
	voronoiSizes[i] = 0
	v[x][y] = i
	var printed bool = true
	possibilities := make(map[coordinate]bool)
	possibilities[coordinate{x, y}] = true
	voronoiSizes[i] += len(possibilities)
	for printed {
		<-stepper
		printed = false
		possibilities = b.possiblePositionsFromArray(possibilities)
		for pos, _ := range possibilities {
			if v[pos.x][pos.y] == -1 {
				v[pos.x][pos.y] = i
				voronoiSizes[i] += 1
				printed = true
			} else if i == whoami && v[pos.x][pos.y] != whoami {
				alone = false
			}
		}
		if printed {
			notifier <- true
		} else {
			close(notifier)
		}
	}
}

//Returns true iff all values are false //TESTED
func allFalse(input []bool) bool {
	for _, val := range input {
		if val {
			return false
		}
	}
	return true
}

////Makes best move based on the fact that the player is inside a dead end
//func delayTheInevitable() string {
//	result, _ := b.getLongestMove(playersHeads[whoami])
//	return result
//}

////Returns the move that leads to a bigger set of movements (longest dead)
////method is associated to board instead of *board to force the copy of the board when the function is called
//func (b board) getLongestMove(c coordinate) (string, int) {
//	b[c.x][c.y] = 9
//	possibilities := b.possiblePositions(c.x, c.y)
//	if len(possibilities) == 0 {
//		return "", 0
//	} else if len(possibilities) == 1 {
//		_, length := b.getLongestMove(possibilities[0])
//		return movement(possibilities[0], c), length + 1
//	}
//	lengths := make([]int, len(possibilities))
//	for i, pos := range possibilities {
//		_, lengths[i] = b.getLongestMove(pos)
//	}
//	longest := positionMaxValue(lengths)
//	return movement(possibilities[longest], c), longest
//}

//Moves towards tha position that maximizes the space to maneouver when alone and approaches strongest rival
func doMove() string {
	possibilities := b.possiblePositions(playersHeads[whoami].x, playersHeads[whoami].y)
	if DEBUG {
		fmt.Println("My possibilities", possibilities)
	}

	if len(possibilities) > 0 {
		optionsValues := computeOptionsValues(possibilities)
		if DEBUG {
			fmt.Println("optionsValues", optionsValues)
			fmt.Println("positionMaxValue(optionsValues)", positionMaxValue(optionsValues))
			fmt.Println("possibilities[positionMaxValue(optionsValues)]", possibilities[positionMaxValue(optionsValues)])
		}
		return movement(possibilities[positionMaxValue(optionsValues)], playersHeads[whoami])
	} else {
		return UP //Nowhere to go. Commiting suicide
	}
}

//Returns the "space" that would be left in each of the options.
func computeOptionsValues(possibilities []coordinate) []int {
	result := make([]int, len(possibilities))
	pointsPerVoronoi := calculatePointsPerVoronoi(possibilities)
	pointsPerSpace := calculatePointsPerSpace(possibilities)
	pointsPerDistance := calculatePointsPerDistance(possibilities)
	if DEBUG {
		fmt.Println("Possibilities", possibilities)
		fmt.Println("pointsPerVoronoi", pointsPerVoronoi)
		fmt.Println("pointsPerSpace", pointsPerSpace)
	}
	for i, _ := range result {
		result[i] = int((float32(pointsPerVoronoi[i])*VORONOI_OVER_SPACE)+float32(pointsPerSpace[i])) + (30 / pointsPerDistance[i])
	}
	return result
}

//Calculates the score of each possibility based on the distance to the nearest opponent
func calculatePointsPerDistance(possibilities []coordinate) []int {
	result := make([]int, len(possibilities))
	for i, pos := range possibilities {
		oldHead := playersHeads[whoami]
		playersHeads[whoami] = coordinate{pos.x, pos.y}
		b[pos.x][pos.y] = whoami
		result[i] = computeDistance()
		if DEBUG {
			fmt.Println("Computing option", pos)
			fmt.Println("Board in this case")
			fmt.Println(b)
			fmt.Println("Voronoi calculating distance")
			fmt.Println(v)
		}
		b[pos.x][pos.y] = -1
		playersHeads[whoami] = oldHead
	}
	return result
}

//Calculates the scrore of each possibility based on the space left after the move
func calculatePointsPerSpace(possibilities []coordinate) []int {
	result := make([]int, len(possibilities))
	for i, pos := range possibilities {
		oldHead := playersHeads[whoami]
		playersHeads[whoami] = coordinate{pos.x, pos.y}
		b[pos.x][pos.y] = whoami
		computeSpace()
		if DEBUG {
			fmt.Println("Computing option", pos)
			fmt.Println("Board in this case")
			fmt.Println(b)
			fmt.Println("Voronoi in this case")
			fmt.Println(v)
		}
		result[i] = voronoiSizes[whoami]
		b[pos.x][pos.y] = -1
		playersHeads[whoami] = oldHead
	}
	return result
}

//Calculates the score of each possibility based on the voronoi size of the player in each move
func calculatePointsPerVoronoi(possibilities []coordinate) []int {
	result := make([]int, len(possibilities))
	for i, pos := range possibilities {
		oldHead := playersHeads[whoami]
		playersHeads[whoami] = coordinate{pos.x, pos.y}
		b[pos.x][pos.y] = whoami
		computeSituation()
		if DEBUG {
			fmt.Println("Computing option", pos)
			fmt.Println("Board in this case")
			fmt.Println(b)
			fmt.Println("Voronoi in this case")
			fmt.Println(v)
		}
		result[i] = voronoiSizes[whoami]
		b[pos.x][pos.y] = -1
		playersHeads[whoami] = oldHead
	}
	return result
}

//Returns the position of the highest positive value //TESTED
func positionMaxValue(vals []int) int {
	var maxVal, result int = -10000, -10000
	for i, val := range vals {
		if maxVal < val {
			maxVal = val
			result = i
		}
	}
	return result
}

//Returns an array with all the possible positions from given array
func (b *board) possiblePositionsFromArray(positions map[coordinate]bool) map[coordinate]bool {
	result := make(map[coordinate]bool, 2*len(positions))
	for inPos, _ := range positions {
		for _, outPos := range b.possiblePositions(inPos.x, inPos.y) {
			result[outPos] = true
		}
	}
	return result
}

//Returns the set of possible positions (i.e. no wall and not outside boundaries) that can be reached from given position //TESTED
func (b *board) possiblePositions(x, y int) []coordinate {
	result := make([]coordinate, 0, 4)
	if x > 0 && b[x-1][y] == -1 {
		result = append(result, coordinate{x - 1, y})
	}
	if x < MAX_X-1 && b[x+1][y] == -1 {
		result = append(result, coordinate{x + 1, y})
	}
	if y > 0 && b[x][y-1] == -1 {
		result = append(result, coordinate{x, y - 1})
	}
	if y < MAX_Y-1 && b[x][y+1] == -1 {
		result = append(result, coordinate{x, y + 1})
	}
	return result
}

//Returns the movement name to given position (RIGHT, LEFT, UP OR DOWN) //TESTED
func movement(to, from coordinate) string {
	if to.y == from.y-1 {
		return UP
	}
	if to.y == from.y+1 {
		return DOWN
	}
	if to.x == from.x-1 {
		return LEFT
	}
	//if to.x == from.x+1 {
	return RIGHT
	//}
}

//Reads the movements from the rest of the players and update board
func readInput() {
	var numPlayers, oldX, oldY, newX, newY int
	if num, err := fmt.Scanf("%d %d\n", &numPlayers, &whoami); err != nil {
		fmt.Println("Error reading number of players. Read", num, "characters. Error:", err)
		os.Exit(1)
	}
	if DEBUG {
		fmt.Println(numPlayers, "players. I am ", whoami)
	}
	//	if len(playersHeads) == 0 {
	playersHeads = make([]coordinate, numPlayers)
	voronoiSizes = make([]int, numPlayers)
	//	}
	for i := 0; i < numPlayers; i += 1 {
		if num, err := fmt.Scanf("%d %d %d %d\n", &oldX, &oldY, &newX, &newY); err != nil {
			fmt.Println("Error reading positions of player", i, ". Read", num, "characters. Error:", err)
			os.Exit(1)
		}
		if DEBUG {
			fmt.Println("Player", i, "to", newX, newY)
		}
		if newX == -1 || newY == -1 {
			b.removeBeam(i)
		} else {
			b[newX][newY] = i
			playersHeads[i].x = newX
			playersHeads[i].y = newY
		}
	}
	if DEBUG {
		fmt.Println("Playing game", playId)
		fmt.Println("I am", whoami)
		playId += 1
	}
}

//Contains the information about the board, the position of the rest of the players and the traces they have left
type board [MAX_X][MAX_Y]int //-1 means the square is empty. Else, the id of the player

//Contains a simple x,y information of a position
type coordinate struct {
	x int
	y int
}

//Prints a visual representation of the board
func (b *board) String() string {
	var result bytes.Buffer
	result.WriteString("+------------------------------+\n")
	for y := 0; y < MAX_Y; y += 1 {
		result.WriteString("|")
		for x := 0; x < MAX_X; x += 1 {
			if b[x][y] == -1 {
				result.WriteString(" ")
			} else {
				result.WriteString(strconv.Itoa(b[x][y]))
			}
		}
		result.WriteString("|\n")
	}
	result.WriteString("+------------------------------+\n")
	return result.String()
}

//Removes the beam of a certain player
func (b *board) removeBeam(p int) {
	for y := 0; y < MAX_Y; y += 1 {
		for x := 0; x < MAX_X; x += 1 {
			if b[x][y] == p {
				b[x][y] = -1
			}
		}
	}
}

//Reads a board from a file
func readBoard(filePath string) {
	fmt.Println("Reading board")
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	for y := 0; y < MAX_Y; y += 1 {
		scanner.Scan()
		line := scanner.Text()
		for x := 0; x < MAX_X && x < len(line); x += 1 {
			if line[x] != 32 { //white space
				b[x][y] = int(line[x] - 48) //Where numbers begin
			}
		}
	}
	fmt.Println("Read board:")
	fmt.Println(b)
}
