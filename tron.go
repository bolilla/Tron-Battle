// Codingame - Tron Battle - http://www.codingame.com/cg/#!challenge:20
package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

const (
	MAX_X  = 30
	MAX_Y  = 20
	MAX_XY = 30
	DEBUG  = true
	UP     = "UP"
	DOWN   = "DOWN"
	LEFT   = "LEFT"
	RIGHT  = "RIGHT"
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
	for i := 0; i < MAX_X*MAX_Y; i += 1 { //Just to avoid infinite loop
		readInput()
		fmt.Println(printMovement())
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
			fmt.Println("BEFORE Board:")
			fmt.Println(b)
		}
		computeSituation()
		if DEBUG {
			fmt.Println("Voronoi:")
			fmt.Println(v)
			fmt.Println("Board:")
			fmt.Println(b)
			fmt.Println("Am I alone?", alone)
		}
		if alone {
			result = delayTheInevitable()
		} else {
			result = attack()
		}
	} else {
		fmt.Println("Skipping step. I am alone. Nothing to think about...")
	}
	return
}

//Analizes current situation
func computeSituation() {
	computeVoronoi()
}

//Computes voronoi board
func computeVoronoi() {
	v = new(board)
	var steppers []chan bool = make([]chan bool, len(playersHeads))
	var notifiers []chan bool = make([]chan bool, len(playersHeads))
	for i := 0; i < len(v); i += 1 {
		for j := 0; j < len(v[i]); j += 1 {
			v[i][j] = -1
		}
	}
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

//Returns true iff all values are false //TESTED
func allFalse(input []bool) bool {
	for _, val := range input {
		if val {
			return false
		}
	}
	return true
}

//Computes the voronoi diagram from given point
func computeVoronoiPerPlayer(x, y, i int, stepper, notifier chan bool) {
	fmt.Println("Computing voronoi from", x, y, "for", i)
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
			} else if v[pos.x][pos.y] != whoami {
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

//Makes best move based on the fact that the player is inside a dead end
func delayTheInevitable() string {
	result, _ := b.getLongestMove(playersHeads[whoami])
	return result
}

//Returns the move that leads to a bigger set of movements (longest dead)
//method is associated to board instead of *board to force the copy of the board when the function is called
func (b board) getLongestMove(c coordinate) (string, int) {
	b[c.x][c.y] = 9
	possibilities := b.possiblePositions(c.x, c.y)
	if len(possibilities) == 0 {
		return "", 0
	} else if len(possibilities) == 1 {
		return movement(possibilities[0], c), 1
	}
	lengths := make([]int, len(possibilities))
	for i, pos := range possibilities {
		_, lengths[i] = b.getLongestMove(pos)
	}
	longest := positionMaxValue(lengths)
	return movement(possibilities[longest], c), longest
}

//Moves towards tha position that maximizes the space to maneouver when alone and approaches strongest rival
func attack() string {
	possibilities := b.possiblePositions(playersHeads[whoami].x, playersHeads[whoami].y)
	if DEBUG {
		fmt.Println("voronoiSizes", voronoiSizes)
		fmt.Println("My possibilities", possibilities)
		fmt.Println(b)
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

//Returns the "space" that would be left in each of the options
func computeOptionsValues(possibilities []coordinate) []int {
	result := make([]int, len(possibilities))
	for i, pos := range possibilities {
		oldVal := b[pos.x][pos.y]
		b[pos.x][pos.y] = whoami
		computeSituation()
		if DEBUG {
			fmt.Println("Board in this case")
			fmt.Println(b)
			fmt.Println("Optional v", i, ":", pos)
			fmt.Println(v)
		}
		result[i] = voronoiSizes[whoami]
		b[pos.x][pos.y] = oldVal
		fmt.Println("Board after restore")
		fmt.Println(b)
	}
	return result
}

//Returns the position of the highest positive value //TESTED
func positionMaxValue(vals []int) int {
	var maxVal, result int
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
	fmt.Println("Possible positions from", x, y, ":", result)
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
	//	if len(playersHeads) == 0 {
	playersHeads = make([]coordinate, numPlayers)
	voronoiSizes = make([]int, numPlayers)
	//	}
	for i := 0; i < numPlayers; i += 1 {
		if num, err := fmt.Scanf("%d %d %d %d\n", &oldX, &oldY, &newX, &newY); err != nil {
			fmt.Println("Error reading positions of player", i, ". Read", num, "characters. Error:", err)
			os.Exit(1)
		}
		if newX == -1 || newY == -1 {
			b.removeBeam(i)
			voronoiSizes[i] = 0
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
				b[x][y] = 0
			}
		}
	}
}
