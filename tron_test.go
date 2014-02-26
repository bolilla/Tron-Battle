// Codingame - Tron Battle - http://www.codingame.com/cg/#!challenge:20
package main

import (
	"testing"
)

//Tests method anyFalse
func TestAllFalse(t *testing.T) {
	var testCases = []struct {
		in  []bool
		out bool
	}{
		{[]bool{false}, true},
		{[]bool{false, false}, true},
		{[]bool{true}, false},
		{[]bool{true, true}, false},
		{[]bool{true, false}, false},
		{[]bool{false, true}, false},
	}
	for i, testCase := range testCases {
		if testCase.out != allFalse(testCase.in) {
			t.Error("Error in item", i, "Got", allFalse(testCase.in), "Expected", testCase.out)
		}
	}
}

//Tests method positionMaxValue
func TestPositionMaxValue(t *testing.T) {
	var testCases = []struct {
		in  []int
		out int
	}{
		{[]int{0}, 0},
		{[]int{0, 1}, 1},
		{[]int{1, 0}, 0},
		{[]int{0, 2, 1}, 1},
		{[]int{1, 0, 3}, 2},
	}
	for i, testCase := range testCases {
		if testCase.out != positionMaxValue(testCase.in) {
			t.Error("Error in item", i, "Got", positionMaxValue(testCase.in), "Expected", testCase.out)
		}
	}
}

//Tests method possiblePositions when no walls are involved
func TestPossiblePositionsNoWall(t *testing.T) {
	b = newBoard()
	var testCases = []struct {
		inX int
		inY int
		out []coordinate
	}{
		{0, 0, []coordinate{{1, 0}, {0, 1}}},
		{MAX_X - 1, 0, []coordinate{{MAX_X - 2, 0}, {MAX_X - 1, 1}}},
		{0, MAX_Y - 1, []coordinate{{1, MAX_Y - 1}, {0, MAX_Y - 2}}},
		{MAX_X - 1, MAX_Y - 1, []coordinate{{MAX_X - 2, MAX_Y - 1}, {MAX_X - 1, MAX_Y - 2}}},
		{5, 5, []coordinate{{4, 5}, {6, 5}, {5, 4}, {5, 6}}},
	}
	for i, testCase := range testCases {
		if !compareCoordinateSlices(testCase.out, b.possiblePositions(testCase.inX, testCase.inY)) {
			t.Error("Error in item", i, "Got", b.possiblePositions(testCase.inX, testCase.inY), "Expected", testCase.out)
		}
	}
}

//Tests method possiblePositions when some walls are involved
func TestPossiblePositionsWithWalls(t *testing.T) {
	b = newBoard()
	var testCases = []struct {
		inX int
		inY int
		out []coordinate
	}{
		{5, 5, []coordinate{{4, 5}, {6, 5}, {5, 4}, {5, 6}}},
		{5, 5, []coordinate{{4, 5}, {6, 5}, {5, 6}}},
		{5, 5, []coordinate{{4, 5}, {5, 6}}},
		{5, 5, []coordinate{{5, 6}}},
		{5, 5, []coordinate{}},
	}
	if !compareCoordinateSlices(testCases[0].out, b.possiblePositions(testCases[0].inX, testCases[0].inY)) {
		t.Error("Error in first test case")
	}
	b[5][4] = 3
	if !compareCoordinateSlices(testCases[1].out, b.possiblePositions(testCases[1].inX, testCases[1].inY)) {
		t.Error("Error in second test case")
	}
	b[6][5] = 3
	if !compareCoordinateSlices(testCases[2].out, b.possiblePositions(testCases[2].inX, testCases[2].inY)) {
		t.Error("Error in third test case")
	}
	b[4][5] = 3
	if !compareCoordinateSlices(testCases[3].out, b.possiblePositions(testCases[3].inX, testCases[3].inY)) {
		t.Error("Error in fourth test case")
	}
	b[5][6] = 3
	if !compareCoordinateSlices(testCases[4].out, b.possiblePositions(testCases[4].inX, testCases[4].inY)) {
		t.Error("Error in fifth test case")
	}

}

//Compares two slices of coordinates
func compareCoordinateSlices(s1, s2 []coordinate) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, _ := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

//Tests method movement
func TestMovement(t *testing.T) {
	b = newBoard()
	var testCases = []struct {
		to, from coordinate
		out      string
	}{
		{coordinate{5, 4}, coordinate{5, 5}, UP},
		{coordinate{5, 6}, coordinate{5, 5}, DOWN},
		{coordinate{4, 5}, coordinate{5, 5}, LEFT},
		{coordinate{6, 5}, coordinate{5, 5}, RIGHT},
		{coordinate{0, 0}, coordinate{5, 5}, RIGHT},
	}
	for i, testCase := range testCases {
		if testCase.out != movement(testCase.to, testCase.from) {
			t.Error("Error in item", i, "Got", movement(testCase.to, testCase.from), "Expected", testCase.out)
		}
	}
}
