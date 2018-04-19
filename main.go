package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/websocket"

	"github.com/guregu/null"
)

func ConvertFizzBuzzMoi(number int) (answer string) {
	if number%3 == 0 {
		answer += "Fizz"
	}
	if number%5 == 0 {
		answer += "Buzz"
	}
	if number%7 == 0 {
		answer += "Moi"
	}
	if len(answer) == 0 {
		answer = strconv.Itoa(number)
	}
	return
}

type FirstMessage struct {
	Signal string `json:"signal"`
}

type QuestionMessage struct {
	Number   int         `json:"number"`
	Previous null.String `json:"previous"`
}

type ResultMessage struct {
	Signal  string `json:"signal"`
	Result  string `json:"success"`
	Score   int    `json:"score"`
	Message string `json:"message"`
}

type AnswerMessage struct {
	Answer string `json:"answer"`
}

type CorrectType int

const (
	NULL CorrectType = iota + 1
	CORRECT
	INCORRECT
)

var (
	nullString    = null.NewString("null", false)
	successString = null.NewString("success", true)
	failureString = null.NewString("failure", true)
)

func HandleWSConnForAnswer(c echo.Context) error {
	logger := c.Logger()

	if !c.IsWebSocket() {
		logger.Info("it's NOT websocket connection.")
		return nil
	}

	logger.Infof("ID: %s accessed.", c.ParamValues()[0])

	websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			logger.Info("websocket connection stopped.")
			ws.Close()
		}()

		// Send first message
		var firstMessage FirstMessage
		if err := websocket.JSON.Receive(ws, &firstMessage); err != nil {
			logger.Fatal(err)
			return
		}
		logger.Info("receive first message.")

		// Process receive answer message
		answerMsgChan := make(chan AnswerMessage)
		go func() {
			defer close(answerMsgChan)

			var answerMessage AnswerMessage
			for {
				// Receive answer message
				if err := websocket.JSON.Receive(ws, &answerMessage); err != nil {
					break
				}
				logger.Info("receive answer message")

				answerMsgChan <- answerMessage
			}
		}()

		// Process send question message
		questionMessage := QuestionMessage{}
		preCorrect := NULL
		correctNum := 0
		questionNum := 0
		quit := time.After(time.Second)
		for cond := true; cond; {

			// Build question message
			questionMessage.Number = generateRand()
			switch preCorrect {
			case NULL:
				questionMessage.Previous = nullString
			case CORRECT:
				questionMessage.Previous = successString
			default:
				questionMessage.Previous = failureString
			}

			// Send question message
			if err := websocket.JSON.Send(ws, &questionMessage); err != nil {
				logger.Error(err)
				break
			}
			logger.Info("send question message")

			// Receive answer message via channel until time out
			select {
			case <-quit:
				logger.Info("time out.")
				cond = false
			case answerMessage := <-answerMsgChan:
				questionNum++
				if ConvertFizzBuzzMoi(questionMessage.Number) == answerMessage.Answer {
					preCorrect = CORRECT
					correctNum++
				} else {
					preCorrect = INCORRECT
				}
			}
		}

		// Send result message
		resultMessage := ResultMessage{}
		resultMessage.Signal = "end"
		resultMessage.Score = correctNum
		if correctNum >= 10 {
			resultMessage.Result = "success"
			resultMessage.Message = fmt.Sprintf("Challenge is successful! Your record is %d / %d 👏", correctNum, questionNum)
		} else {
			resultMessage.Result = "faild"
			resultMessage.Message = fmt.Sprintf("faild. %d / %d", correctNum, questionNum)
		}
		websocket.JSON.Send(ws, &resultMessage)
		logger.Info("send result message")
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func generateRand() int {
	return rand.Intn(100)
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Recover())
	e.GET("/websocket/:id", HandleWSConnForAnswer)

	if err := e.Start(":12345"); err != nil {
		e.Logger.Fatal(err)
	}
}
