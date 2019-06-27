package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const (
	leftArrowKey  = string('\ue012')
	upArrowKey    = string('\ue013')
	rightArrowKey = string('\ue014')
	downArrowKey  = string('\ue015')

	seleniumPath    = "vendor/selenium-server-standalone-3.14.0.jar"
	geckoDriverPath = "vendor/geckodriver-v0.23.0-linux64"
	port            = 8080
)

/*func main() {
	service, wd := getWD()
	defer wd.Quit()
	defer service.Stop()

	// Navigate to the game.
	if err := wd.Get("https://4ark.me/2048/"); err != nil {
		panic(err)
	}

	// Wait to load
	time.Sleep(2000 * time.Millisecond)

	// Send keys
	sendKey(wd, leftArrowKey)
	getScore(wd)
	sendKey(wd, rightArrowKey)
	getScore(wd)
	sendKey(wd, upArrowKey)
	getScore(wd)
	sendKey(wd, downArrowKey)
	getScore(wd)
	printGrid(wd)

	// Wait for exit
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	wd.Quit()
	service.Stop()
}*/

func printGrid(wd selenium.WebDriver) {
	grid := getGrid(wd)

	fmt.Printf("Current grid\n")
	fmt.Printf("%d %d %d %d\n", grid[0], grid[1], grid[2], grid[3])
	fmt.Printf("%d %d %d %d\n", grid[4], grid[5], grid[6], grid[7])
	fmt.Printf("%d %d %d %d\n", grid[8], grid[9], grid[10], grid[11])
	fmt.Printf("%d %d %d %d\n", grid[12], grid[13], grid[14], grid[15])
}

func getGrid(wd selenium.WebDriver) [16]int {
	weArray, err := wd.FindElements(selenium.ByCSSSelector, ".tile")
	if err != nil {
		panic(err)
	}

	var cellValues [16]int

	for _, element := range weArray {
		dataValue, err := element.Text()
		if err != nil {
			panic(err)
		}

		dataIndex, err := element.GetAttribute("data-index")
		if err != nil {
			panic(err)
		}

		cellValues[toInt(dataIndex)] = toInt(dataValue)
	}

	return cellValues
}

func getScore(wd selenium.WebDriver) int {
	scoreDiv, err := wd.FindElement(selenium.ByCSSSelector, ".score-container > p:nth-child(2)")
	if err != nil {
		panic(err)
	}

	var output string
	for {
		output, err = scoreDiv.Text()
		if err != nil {
			panic(err)
		}
		if output != "Waiting for remote server..." {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Printf("Current score: %s\n", output)

	score, err := strconv.Atoi(output)
	if err != nil {
		panic(err)
	}

	return score
}

func didLose(wd selenium.WebDriver) bool {
	element, err := wd.FindElement(selenium.ByCSSSelector, ".failure-container")
	if err != nil {
		return false
	}
	classes, err := element.GetAttribute("class")
	if err != nil {
		return false
	}
	return strings.Contains(classes, "action")
}

func toInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return num
}

func restartGame(wd selenium.WebDriver) {
	restartButton, err := wd.FindElement(selenium.ByCSSSelector, ".restart-btn")
	if err != nil {
		panic(err)
	}
	restartButton.Click()
}

func sendKey(wd selenium.WebDriver, key string) {
	wd.KeyDown(key)
	wd.KeyUp(key)
	//time.Sleep(50 * time.Millisecond)
}

func getWD() (*selenium.Service, selenium.WebDriver) {
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	//selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}

	return service, wd
}
