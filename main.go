package main

import (
	"fmt"
	"github.com/muesli/termenv"
	"image/color"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)
import "bufio"

type State struct {
	Count uint64
}

func colored(text string, color termenv.Color) string {
	return fmt.Sprint(termenv.String(text).Foreground(color))
}

func red() color.Color {
	return color.RGBA{255, 130, 0, 255}
}

func yellow() color.Color {
	return color.RGBA{255, 240, 0, 255}
}

func veryDarkGray() color.Color {
	return color.RGBA{90, 90, 90, 255}
}

func green() color.Color {
	return color.RGBA{190, 255, 0, 255}
}

func lightGray() color.Color {
	return color.RGBA{200, 200, 200, 255}
}

func lessLightGray() color.Color {
	return color.RGBA{150, 150, 150, 255}
}

func darkGray() color.Color {
	return color.RGBA{100, 100, 100, 255}
}

func blue() color.Color {
	return color.RGBA{140, 200, 255, 255}
}

type ParenLikePair struct {
	start string
	end   string
}

// TODO: broken, buggy
func colorParenLike(line []ChunkWithColors, state *State) []ChunkWithColors {
	supported := []ParenLikePair{{"(", ")"}, {"[", "]"}, {"{", "}"}, {"<", ">"}, {"|", "|"}, {"'", "'"}, {`"`, `"`}, {"`", "`"}}
	charToParenLikePair := map[string]ParenLikePair{}
	for _, chars := range supported {
		charToParenLikePair[chars.start] = chars
		charToParenLikePair[chars.end] = chars
	}
	positions := map[string][]int{}

	var newLine []ChunkWithColors

	position := -1
	for _, c := range line {
		for _, char := range c.string {
			position++
			str := string(char)
			matchedPair, exists := charToParenLikePair[str]
			if !exists {
				newLine = append(newLine,
					ChunkWithColors{
						string:      str,
						colorLayers: c.colorLayers,
						marker:      c.marker,
					},
				)
				continue
			}
			if str == matchedPair.start {
				positions[str] = append(positions[str], position)
			}
			if str == matchedPair.end {
				matchedPositions, exists := positions[matchedPair.start]
				if exists && len(matchedPositions) > 0 {
					startPosition := matchedPositions[len(matchedPositions)-1]
					matchedPositions = matchedPositions[:len(matchedPositions)-1]

					var newColors []color.Color
					copy(newColors, c.colorLayers)

					var newMarker []Marker
					copy(newMarker, c.marker)

					position2 := -1
					for _, chunk := range line {
						done := false
						for positionChunk, _ := range chunk.string {
							position2++
							if done {
								continue
							}
							if position2 < startPosition {
								continue
							}
							if position2+len(chunk.string) >= position {
								newLine = append(newLine, ChunkWithColors{
									string:      chunk.string[positionChunk : positionChunk+(position+1)-position2],
									colorLayers: append(newColors, red()),
									marker:      append(newMarker, ParenLike),
								})
								done = true
							} else {
								//newLine = append(newLine, ChunkWithColors{
								//	string:      chunk.string,
								//	colorLayers: append(newColors, red()),
								//	marker:      append(newMarker, ParenLike),
								//})
								//break
							}
						}
					}
					//newLine = append(newLine, ChunkWithColors{
					//	string:      c.string[startPosition:(position + 1)],
					//	colorLayers: append(newColors, red()),
					//	marker:      append(newMarker, ParenLike),
					//})
				}
			}
		}
	}

	return newLine
}

// https://stackoverflow.com/a/38191078/2709248
func colorUUID(line []ChunkWithColors, state *State) []ChunkWithColors {
	r, err := regexp.Compile("(?i)((?:^|\\b)([0-9A-F]{8}-[0-9A-F]{4}-[1][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})|([0-9A-F]{8}-[0-9A-F]{4}-[2][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})|([0-9A-F]{8}-[0-9A-F]{4}-[3][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})|([0-9A-F]{8}-[0-9A-F]{4}-[4][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})|([0-9A-F]{8}-[0-9A-F]{4}-[5][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})(?:$|\\b))")

	if err != nil {
		panic(err)
	}

	var newLine []ChunkWithColors
	for _, c := range line {
		indexes := r.FindAllStringIndex(c.string, -1)
		if len(indexes) == 0 {
			newLine = append(newLine, c)
			continue
		}
		for _, index := range indexes {
			if index[0] != 0 {
				newLine = append(newLine, ChunkWithColors{
					string:      c.string[:index[0]],
					colorLayers: c.colorLayers,
					marker:      c.marker,
				})
			}

			structuredStuff := c.string[index[0]:index[1]]

			var newColors []color.Color
			copy(newColors, c.colorLayers)

			var newMarker []Marker
			copy(newMarker, c.marker)

			newLine = append(newLine, ChunkWithColors{
				string:      structuredStuff,
				colorLayers: append(newColors, green()),
				marker:      append(newMarker, UUID),
			})

			if index[1] != len(c.string) {
				newLine = append(newLine, ChunkWithColors{
					string:      c.string[index[1]:],
					colorLayers: c.colorLayers,
					marker:      c.marker,
				})
			}
		}
	}

	return newLine
}

// TODO: broken, buggy
func colorStructuredStuff(line []ChunkWithColors, state *State) []ChunkWithColors {
	r, err := regexp.Compile("((?:^|\\b)[a-zA-Z0-9-_]+ {0,2}[:=] {0,2}[a-zA-Z0-9-_]+(?:$|\\b))")

	if err != nil {
		panic(err)
	}
	var newLine []ChunkWithColors
	for _, c := range line {
		indexes := r.FindAllStringIndex(c.string, -1)
		if len(indexes) == 0 {
			newLine = append(newLine, c)
			continue
		}
		for _, index := range indexes {
			if index[0] != 0 {
				newLine = append(newLine, ChunkWithColors{
					string:      c.string[:index[0]],
					colorLayers: c.colorLayers,
					marker:      c.marker,
				})
			}

			structuredStuff := c.string[index[0]:index[1]]

			var newColors []color.Color
			copy(newColors, c.colorLayers)

			var newMarker []Marker
			copy(newMarker, c.marker)

			newLine = append(newLine, ChunkWithColors{
				string:      structuredStuff,
				colorLayers: append(newColors, yellow()),
				marker:      append(newMarker, Structure),
			})

			if index[1] != len(c.string) {
				newLine = append(newLine, ChunkWithColors{
					string:      c.string[index[1]:],
					colorLayers: c.colorLayers,
					marker:      c.marker,
				})
			}
		}
	}

	return newLine
}

func colorAndConvertTimestamp(line []ChunkWithColors, state *State) []ChunkWithColors {
	r, err := regexp.Compile("((?:^|\\b)1[0-9]{9})|(1[0-9]{9}(?:$|\\b))") // FIXME: this is ugly and supports only soome timestamps
	if err != nil {
		panic(err)
	}
	var newLine []ChunkWithColors
	for _, c := range line {
		index := r.FindStringIndex(c.string)
		if index == nil {
			newLine = append(newLine, c)
			continue
		}
		i, err := strconv.ParseInt(c.string[index[0]:index[1]], 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)

		if index[0] != 0 {
			newLine = append(newLine, ChunkWithColors{
				string:      c.string[:index[0]],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}

		var newColors []color.Color
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLine = append(newLine, ChunkWithColors{
			string:      tm.String(),
			colorLayers: append(newColors, blue()),
			marker:      append(newMarker, Timestamp),
		})

		if index[1] != len(c.string) {
			newLine = append(newLine, ChunkWithColors{
				string:      c.string[index[1]:],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}
	}

	return newLine
}

func colorLogLevel(line []ChunkWithColors, state *State) []ChunkWithColors {
	var newLine []ChunkWithColors
	r, err := regexp.Compile("(?i)(?:^|\\b)(info|error|warn|debug)(?:$|\\b)")
	if err != nil {
		panic(err)
	}
	for _, c := range line {
		index := r.FindStringIndex(c.string)
		if index == nil {
			newLine = append(newLine, c)
			continue
		}
		logLevel := c.string[index[0]:index[1]]
		var colors = map[string]color.Color{"info": green(), "error": red(), "warn": yellow(), "debug": veryDarkGray()}
		var newColor = colors[strings.ToLower(logLevel)]

		if index[0] != 0 {
			newLine = append(newLine, ChunkWithColors{
				string:      c.string[:index[0]],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}

		var newColors []color.Color
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLine = append(newLine, ChunkWithColors{
			string:      logLevel,
			colorLayers: append(newColors, newColor),
			marker:      append(newMarker, LogLevel),
		})

		if index[1] != len(c.string) {
			newLine = append(newLine, ChunkWithColors{
				string:      c.string[index[1]:],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}
	}
	return newLine
}

func main() {

	readIt()
}

func colorLine(line []ChunkWithColors, state *State) []ChunkWithColors {
	var newLine []ChunkWithColors
	for _, c := range line {
		var color1 color.Color
		var marker Marker
		if state.Count%2 == 0 {
			color1 = lightGray()
			marker = Even
		} else {
			color1 = lessLightGray()
			marker = Odd
		}
		var newColors []color.Color
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLine = append(newLine, ChunkWithColors{string: c.string, colorLayers: append(newColors, color1), marker: append(newMarker, marker)})
	}
	return newLine
}

type Marker string

const (
	Even      Marker = "Even"
	Odd              = "Odd"
	Timestamp        = "Timestamp"
	Structure        = "Timestamp"
	UUID             = "UUID"
	Hex              = "Hex"
	ParenLike        = "ParenLike"
	LogLevel         = "LogLevel"
)

type ChunkWithColors struct {
	string      string
	colorLayers []color.Color
	marker      []Marker
}

func readIt() {
	state := State{Count: 0}

	stdin := bufio.NewReader(os.Stdin)

	for true {
		line, err := stdin.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		coloredLine := []ChunkWithColors{{string: line, colorLayers: []color.Color{}}}

		coloredLine = colorLine(coloredLine, &state)
		//coloredLine = colorStructuredStuff(coloredLine, &state)
		coloredLine = colorLogLevel(coloredLine, &state)
		coloredLine = colorAndConvertTimestamp(coloredLine, &state)
		coloredLine = colorUUID(coloredLine, &state)
		//coloredLine = colorParenLike(coloredLine, &state)

		finalString := ""
		for _, chunkWithColors := range coloredLine {
			//blendedColors := chunkWithColors.colorLayers[len(chunkWithColors.colorLayers)-1]
			blendedColors := blendAllColors(chunkWithColors.colorLayers)
			finalString += colored(chunkWithColors.string, colorfulToAnsi(blendedColors))
		}

		state.Count += 1

		_, err = os.Stdout.Write([]byte(finalString))
		if err != nil {
			panic(err)
		}

	}
}
