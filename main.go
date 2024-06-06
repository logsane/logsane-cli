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
		coloredLine = colorLogLevel(coloredLine, &state)
		coloredLine = colorAndConvertTimestamp(coloredLine, &state)

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
