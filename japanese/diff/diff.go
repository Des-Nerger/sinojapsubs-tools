package diff
import (
	"fmt"
	"reflect"
	. "unsafe"
)

type Component struct {
	Value []rune
	Added, Removed bool
}
func (c Component) String() string {
//*
	return string(c.Value)
/*/
	return fmt.Sprintf("\033[%vm%v\033[m",
		func() string {
			if c.Added {return "32"}
			if c.Removed {return "31"}
			return ""
		} (),
		string(c.Value))
//*/
}

type path struct {
	newPos int
	components []Component
}
func (p *path) clone() *path {
	return &path{
		newPos: p.newPos,
		components: append(make([]Component, 0, cap(p.components)+2), p.components...),
	}
}
//func (p *path) String() string {if p==nil {return "<nil>"}; return fmt.Sprintf("%v %v", p.newPos, p.components)}

var panicString = *(*string)(Pointer(&reflect.StringHeader{Data: uintptr(0), Len: 1}))
func main() {
	fmt.Printf("%v\n", Do("くらくくらくくらくくらくくらく", "暗く暗く暗く暗く暗く"))
}

func Do(oldString, newString string) []Component {
	// Handle the identity case (this is due to unrolling editLength == 0
	if newString == oldString { return []Component{{Value: []rune(newString)}} }
	if newString == "" { return []Component{{Value: []rune(oldString), Removed: true}} }
	if oldString == "" { return []Component{{Value: []rune(newString), Added: true}} }
	newRunes, oldRunes := []rune(newString), []rune(oldString)
	oldString, newString = panicString, panicString //assurance to not use them anymore
	newLen, oldLen := len(newRunes), len(oldRunes)
	maxEditLength := newLen + oldLen
	bestPath := map[int]*path {0: {newPos: -1}}

	// Seed editLength = 0
	oldPos := extractCommon(bestPath[0], newRunes, oldRunes, 0)
	if bestPath[0].newPos+1 >= newLen && oldPos+1 >= oldLen {return bestPath[0].components}

	for editLength:=1; editLength<=maxEditLength; editLength++ {
		for diagonalPath:=-editLength; diagonalPath<=+editLength; diagonalPath+=2 {
			addPath, removePath := bestPath[diagonalPath-1], bestPath[diagonalPath+1]
			oldPos = func() int {if removePath!=nil {return removePath.newPos}; return 0} () - diagonalPath
			if addPath!=nil {delete(bestPath, diagonalPath-1)} // No one else is going to attempt to use this value, clear it
			canAdd, canRemove := addPath!=nil && addPath.newPos+1<newLen, removePath!=nil && 0<=oldPos && oldPos<oldLen
			if !canAdd && !canRemove {delete(bestPath, diagonalPath); continue}
			/* Select the diagonal that we want to branch from. We select the prior
			   path whose position in the new string is the farthest from the origin
			   and does not pass the bounds of the diff graph */
			var basePath *path
			if !canAdd || canRemove && addPath.newPos<removePath.newPos {
				basePath = removePath.clone()
				basePath.components = appendComponent(basePath.components,
					Component{Value: oldRunes[oldPos:oldPos+1], Removed: true})
			} else {
				basePath = addPath.clone()
				basePath.newPos++
				basePath.components = appendComponent(basePath.components,
					Component{Value: newRunes[basePath.newPos:basePath.newPos+1], Added: true})
			}
			oldPos = extractCommon(basePath, newRunes, oldRunes, diagonalPath)
			if basePath.newPos+1>=newLen && oldPos+1>=oldLen {return basePath.components}
			bestPath[diagonalPath] = basePath
		}
	}
	return nil
}

func appendComponent(components []Component, component Component) []Component {
	last := func() *Component {
		if len(components)==0 {return nil}
		return &components[len(components)-1]
	} ()
	if last!=nil && last.Added == component.Added && last.Removed == component.Removed {
		last.Value = join(last.Value, component.Value)
		return components
	}
	return append(components, component)
}

func extractCommon(basePath *path, newRunes, oldRunes []rune, diagonalPath int) int {
	newLen, oldLen, newPos := len(newRunes), len(oldRunes), basePath.newPos
	oldPos := newPos - diagonalPath
	if len(basePath.components)==0 || !(basePath.components[len(basePath.components)-1].Added) {
		for newPos+1<newLen && oldPos+1<oldLen && newRunes[newPos+1]==oldRunes[oldPos+1] {
			newPos++; oldPos++
			basePath.components = appendComponent(basePath.components, Component{Value: newRunes[newPos:newPos+1]})
		}
	}
	basePath.newPos = newPos
	return oldPos
}

func join(left, right []rune) []rune {
	leftSH, rightSH := (*reflect.SliceHeader)(Pointer(&left)), (*reflect.SliceHeader)(Pointer(&right))
	if leftSH.Data+uintptr(leftSH.Len)*Sizeof(rune(0)) == rightSH.Data {return left[:len(left)+len(right)]}
	panic(fmt.Sprintf("%x:%q and %x:%q are not adjacent", leftSH.Data, left, rightSH.Data, right))
}
