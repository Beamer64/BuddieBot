package prefix

import (
	"fmt"
	"github.com/StephaneBunel/bresenham"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"testing"
)

func TestCistercianNumber(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

	input := "1685"

	var imgRect = image.Rect(0, 0, 100, 100)
	var img = image.NewRGBA(imgRect)
	var col = color.RGBA{255, 255, 255, 255}

	// draw vertical line
	bresenham.DrawLine(img, 50, 10, 50, 90, col)

	bresenham.DrawLine(img, 30, 50, 70, 50, col)

	var x1 int
	var x2 int
	var y1 int
	var y2 int
	for pos, char := range input {
		// fmt.Printf("character %c starts at byte position %d\n", char, pos)
		switch pos {
		case 0: // thous
			switch char {
			case '5':
				bresenham.DrawLine(img, 30, 90, 50, 70, col)
			case '7':
				bresenham.DrawLine(img, 30, 90, 30, 70, col)
			case '8':
				bresenham.DrawLine(img, 30, 70, 30, 90, col)
			case '9':
				bresenham.DrawLine(img, 30, 90, 30, 70, col)
				bresenham.DrawLine(img, 30, 70, 50, 70, col)
			}

			x1 = thous[string(char)].x1
			y1 = thous[string(char)].y1
			x2 = thous[string(char)].x2
			y2 = thous[string(char)].y2
		case 1: // hunds
			switch char {
			case '5':
				bresenham.DrawLine(img, 70, 90, 50, 70, col)
			case '7':
				bresenham.DrawLine(img, 70, 90, 70, 70, col)
			case '8':
				bresenham.DrawLine(img, 70, 70, 70, 90, col)
			case '9':
				bresenham.DrawLine(img, 70, 90, 70, 70, col)
				bresenham.DrawLine(img, 70, 70, 50, 70, col)
			}

			x1 = hunds[string(char)].x1
			y1 = hunds[string(char)].y1
			x2 = hunds[string(char)].x2
			y2 = hunds[string(char)].y2
		case 2: // tens
			switch char {
			case '5':
				bresenham.DrawLine(img, 30, 10, 50, 30, col)
			case '7':
				bresenham.DrawLine(img, 30, 10, 30, 30, col)
			case '8':
				bresenham.DrawLine(img, 30, 30, 30, 10, col)
			case '9':
				bresenham.DrawLine(img, 30, 10, 30, 30, col)
				bresenham.DrawLine(img, 30, 30, 50, 30, col)
			}

			x1 = tens[string(char)].x1
			y1 = tens[string(char)].y1
			x2 = tens[string(char)].x2
			y2 = tens[string(char)].y2
		case 3: // ones
			switch char {
			case '5':
				bresenham.DrawLine(img, 50, 30, 70, 10, col)
			case '7':
				bresenham.DrawLine(img, 70, 10, 70, 30, col)
			case '8':
				bresenham.DrawLine(img, 70, 30, 70, 10, col)
			case '9':
				bresenham.DrawLine(img, 70, 10, 70, 30, col)
				bresenham.DrawLine(img, 70, 30, 50, 30, col)
			}

			x1 = ones[string(char)].x1
			y1 = ones[string(char)].y1
			x2 = ones[string(char)].x2
			y2 = ones[string(char)].y2
		}

		bresenham.DrawLine(img, x1, y1, x2, y2, col)
	}

	// save imageInfo
	toimg, err := os.Create("../../res/genFiles/symbol.png")
	if err != nil {
		t.Fatal(err)
	}
	defer toimg.Close()

	err = png.Encode(toimg, img)
	if err != nil {
		t.Fatal(err)
	}

}

func TestCistercianRange(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	pngName := "Symbol.png"
	// i := 0; i <= 10; i++
	// i := 10; i <= 90; i += 10
	// i := 100; i <= 900; i += 100
	// i := 1000; i <= 9000; i += 1000
	for i := 1000; i <= 9000; i += 1000 {
		strI := strconv.Itoa(i)
		switch len(strI) {
		case 1:
			strI = "000" + strI
		case 2:
			strI = "00" + strI
		case 3:
			strI = "0" + strI
		}
		pngName = fmt.Sprintf("Symbol_%d.png", i)

		var imgRect = image.Rect(0, 0, 100, 100)
		var img = image.NewRGBA(imgRect)
		var col = color.RGBA{255, 255, 255, 255}

		// draw vertical line
		bresenham.DrawLine(img, 50, 10, 50, 90, col)

		var x1 int
		var x2 int
		var y1 int
		var y2 int
		for pos, char := range strI {
			// fmt.Printf("character %c starts at byte position %d\n", char, pos)
			switch pos {
			case 0: // thous
				switch char {
				case '5':
					bresenham.DrawLine(img, 30, 90, 50, 70, col)
				case '7':
					bresenham.DrawLine(img, 30, 90, 30, 70, col)
				case '8':
					bresenham.DrawLine(img, 30, 70, 30, 90, col)
				case '9':
					bresenham.DrawLine(img, 30, 90, 30, 70, col)
					bresenham.DrawLine(img, 30, 70, 50, 70, col)
				}

				x1 = thous[string(char)].x1
				y1 = thous[string(char)].y1
				x2 = thous[string(char)].x2
				y2 = thous[string(char)].y2
			case 1: // hunds
				switch char {
				case '5':
					bresenham.DrawLine(img, 70, 90, 50, 70, col)
				case '7':
					bresenham.DrawLine(img, 70, 90, 70, 70, col)
				case '8':
					bresenham.DrawLine(img, 70, 70, 70, 90, col)
				case '9':
					bresenham.DrawLine(img, 70, 90, 70, 70, col)
					bresenham.DrawLine(img, 70, 70, 50, 70, col)
				}

				x1 = hunds[string(char)].x1
				y1 = hunds[string(char)].y1
				x2 = hunds[string(char)].x2
				y2 = hunds[string(char)].y2
			case 2: // tens
				switch char {
				case '5':
					bresenham.DrawLine(img, 30, 10, 50, 30, col)
				case '7':
					bresenham.DrawLine(img, 30, 10, 30, 30, col)
				case '8':
					bresenham.DrawLine(img, 30, 30, 30, 10, col)
				case '9':
					bresenham.DrawLine(img, 30, 10, 30, 30, col)
					bresenham.DrawLine(img, 30, 30, 50, 30, col)
				}

				x1 = tens[string(char)].x1
				y1 = tens[string(char)].y1
				x2 = tens[string(char)].x2
				y2 = tens[string(char)].y2
			case 3: // ones
				switch char {
				case '5':
					bresenham.DrawLine(img, 50, 30, 70, 10, col)
				case '7':
					bresenham.DrawLine(img, 70, 10, 70, 30, col)
				case '8':
					bresenham.DrawLine(img, 70, 30, 70, 10, col)
				case '9':
					bresenham.DrawLine(img, 70, 10, 70, 30, col)
					bresenham.DrawLine(img, 70, 30, 50, 30, col)
				}

				x1 = ones[string(char)].x1
				y1 = ones[string(char)].y1
				x2 = ones[string(char)].x2
				y2 = ones[string(char)].y2
			}

			bresenham.DrawLine(img, x1, y1, x2, y2, col)
		}

		// save imageInfo
		toimg, _ := os.Create(pngName)
		defer toimg.Close()

		err := png.Encode(toimg, img)
		if err != nil {
			t.Fatal(err)
		}
	}
}
