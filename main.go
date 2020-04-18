package main

import (
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

func ansiColor(str string, code uint8) string {
	return "\u001b[38;5;" + strconv.Itoa(int(code)) + "m" + str + "\u001b[0m"
}

func red() uint8 {
	return 196
}

func yellow() uint8 {
	return 226
}

func veryDarkGray() uint8 {
	return 240
}

func green() uint8 {
	return 46
}

func lightGray() uint8 {
	return 253
}

func darkGray() uint8 {
	return 245
}

func blue() uint8 {
	return 80
}

func colorAndConvertTimestamp(lines []ChunkWithColors, state *State) []ChunkWithColors {
	r, err := regexp.Compile("((?:^|\\b)1[0-9]{9})|(1[0-9]{9}(?:$|\\b))") // FIXME: this is ugly and supports only soome timestamps
	if err != nil {
		panic(err)
	}
	var newLines []ChunkWithColors
	for _, c := range lines {
		index := r.FindStringIndex(c.string)
		if index == nil {
			newLines = append(newLines, c)
			continue
		}
		i, err := strconv.ParseInt(c.string[index[0]:index[1]], 10, 64)
		if err != nil {
			panic(err)
		}
		tm := time.Unix(i, 0)

		if index[0] != 0 {
			newLines = append(newLines, ChunkWithColors{
				string:      c.string[:index[0]],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}

		var newColors []uint8
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLines = append(newLines, ChunkWithColors{
			string:      tm.String(),
			colorLayers: append(newColors, blue()),
			marker:      append(newMarker, Timestamp),
		})

		if index[1] != len(c.string) {
			newLines = append(newLines, ChunkWithColors{
				string:      c.string[index[1]:],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}
	}

	return newLines
}

func colorLogLevel(lines []ChunkWithColors, state *State) []ChunkWithColors {
	var newLines []ChunkWithColors
	r, err := regexp.Compile("(?i)(?:^|\\b)(info|error|warn|debug)(?:$|\\b)")
	if err != nil {
		panic(err)
	}
	for _, c := range lines {
		index := r.FindStringIndex(c.string)
		if index == nil {
			newLines = append(newLines, c)
			continue
		}
		logLevel := c.string[index[0]:index[1]]
		var colors = map[string]uint8{"info": green(), "error": red(), "warn": yellow(), "debug": veryDarkGray()}
		var newColor = colors[strings.ToLower(logLevel)]

		if index[0] != 0 {
			newLines = append(newLines, ChunkWithColors{
				string:      c.string[:index[0]],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}

		var newColors []uint8
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLines = append(newLines, ChunkWithColors{
			string:      logLevel,
			colorLayers: append(newColors, newColor),
			marker:      append(newMarker, LogLevel),
		})

		if index[1] != len(c.string) {
			newLines = append(newLines, ChunkWithColors{
				string:      c.string[index[1]:],
				colorLayers: c.colorLayers,
				marker:      c.marker,
			})
		}
	}
	return newLines
}

func main() {

	readIt()
}

func colorLine(lines []ChunkWithColors, state *State) []ChunkWithColors {
	var newLines []ChunkWithColors
	for _, c := range lines {
		var color uint8
		var marker Marker
		if state.Count%2 == 0 {
			color = lightGray()
			marker = Even
		} else {
			color = darkGray()
			marker = Odd
		}
		var newColors []uint8
		copy(newColors, c.colorLayers)

		var newMarker []Marker
		copy(newMarker, c.marker)

		newLines = append(newLines, ChunkWithColors{string: c.string, colorLayers: append(newColors, color), marker: append(newMarker, marker)})
	}
	return newLines
}

type Marker string

const (
	Even      Marker = "Even"
	Odd              = "Odd"
	Timestamp        = "Timestamp"
	LogLevel         = "LogLevel"
)

type ChunkWithColors struct {
	string      string
	colorLayers []uint8
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

		lines := []ChunkWithColors{{string: line, colorLayers: []uint8{}}}

		lines = colorLine(lines, &state)
		lines = colorLogLevel(lines, &state)
		lines = colorAndConvertTimestamp(lines, &state)

		finalString := ""
		for _, chunkWithColors := range lines {
			colorLayers := chunkWithColors.colorLayers
			finalString += ansiColor(chunkWithColors.string, colorLayers[len(colorLayers)-1])
		}

		state.Count += 1

		_, err = os.Stdout.Write([]byte(finalString))
		if err != nil {
			panic(err)
		}

	}
}
