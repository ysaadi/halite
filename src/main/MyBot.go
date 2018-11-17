package main

import (
	"fmt"
	"hlt"
	"hlt/gameconfig"
	"hlt/log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func gracefulExit(logger *log.FileLogger) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		fmt.Println("Wait for 2 second to finish processing")
		time.Sleep(2 * time.Second)
		logger.Close()
		os.Exit(0)
	}()
}

func aveInDirection(game *hlt.Game, direction *Direction) float64 {
	var gameMap = game.Map
	var me = game.Me
	var shipyard = me.Shipyard
	var pathPosition = shipyard.E.Pos
	var haliteSum int
	for x := 0; x < 25; x++ {
		pathPosition, _ = pathPosition.DirectionalOffset(direction)
		haliteSum += gameMap.AtPosition(pathPosition).Halite
	}
	var haliteAverage = float64(haliteSum) / 25.0
	return haliteAverage
}

func main() {
	args := os.Args
	var seed = time.Now().UnixNano() % int64(os.Getpid())
	if len(args) > 1 {
		seed, _ = strconv.ParseInt(args[0], 10, 64)
	}
	rand.Seed(seed)

	var game = hlt.NewGame()
	// At this point "game" variable is populated with initial map data.
	// This is a good place to do computationally expensive start-up pre-processing.
	// As soon as you call "ready" function below, the 2 second per turn timer will start.

	var config = gameconfig.GetInstance()
	fileLogger := log.NewFileLogger(game.Me.ID)
	var logger = fileLogger.Logger

	var northAve = aveInDirection(game, hlt.North())
	logger.Printf("Average halite in the 25 north blocks is %d", northAve)

	logger.Printf("Successfully created bot! My Player ID is %d. Bot rng seed is %d.", game.Me.ID, seed)
	gracefulExit(fileLogger)
	game.Ready("MyBot")
	for {
		game.UpdateFrame()
		var me = game.Me
		var gameMap = game.Map
		var ships = me.Ships
		var commands = []hlt.Command{}
		for i := range ships {
			var ship = ships[i]
			var maxHalite, _ = config.GetInt(gameconfig.MaxHalite)
			var currentCell = gameMap.AtEntity(ship.E)
			if currentCell.Halite < (maxHalite/10) || ship.IsFull() {
				commands = append(commands, ship.Move(hlt.AllDirections[rand.Intn(4)]))
			} else {
				commands = append(commands, ship.Move(hlt.Still()))
			}
		}
		var shipCost, _ = config.GetInt(gameconfig.ShipCost)
		if game.TurnNumber <= 200 && me.Halite >= shipCost && !gameMap.AtEntity(me.Shipyard.E).IsOccupied() {
			commands = append(commands, hlt.SpawnShip{})
		}
		game.EndTurn(commands)
	}
}
