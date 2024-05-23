package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/xlab/portmidi"
)

func main() {
	// music
	rstrt := make(chan bool, 3)
	portmidi.Initialize()
	defer portmidi.Terminate()
	go func(restart chan bool) {
		waitTime := 500
		id, _ := portmidi.DefaultOutputDeviceID()
		out, err := portmidi.NewOutputStream(id, 1024, 0, 0)
		if err != nil {
			panic(err)
		}
		sink := out.Sink()
		offset := 12 * 4
		for {
			// switch {
			// case <-restart:
			// 	waitTime = 500
			// default:
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+24), 20)} //C
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+26), 20)} //D
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+27), 20)} //Eb
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+29), 20)} //F
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+31), 20)} //G
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+32), 20)} //Ab
			<-time.After(time.Duration(waitTime) * time.Millisecond)
			sink <- portmidi.Event{Timestamp: int32(time.Now().Unix()), Message: portmidi.NewMessage(0x90, byte(offset+34), 20)} //Bb
			waitTime = int(float32(waitTime) * float32(0.95))
			if len(restart) > 0 {
				waitTime = 500
				<-restart
			}
			// }
		}
	}(rstrt)

	// score
	data, err := ioutil.ReadFile("score.txt")
	highscore := 0
	resetScore := func() {
		os.Remove("score.txt")
		f, err := os.Create("score.txt")
		if err != nil {
			fmt.Println(err)
		} else {
			f.Write([]byte("420")) // epic lmao
			f.Close()
		}
	}
	if err != nil {
		fmt.Println(err)
		resetScore()
	} else {
		hs, err := strconv.ParseInt(string(data), 10, 64)
		highscore = int(hs)
		if err != nil {
			highscore = 0
			resetScore()
		}
	}
	// window cfg
	rl.SetConfigFlags(rl.FlagWindowUndecorated | rl.FlagVsyncHint)

	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "raylib [core] example - basic window")
	rl.SetTargetFPS(60)
	shader := rl.LoadShader("", "bloom.fs")
	defer rl.UnloadShader(shader)

	// global varaibles
	speed := float32(10)
	npipes := rl.GetScreenWidth() / 250
	a := float32(0)
	v := float32(0)
	player := rl.Rectangle{}
	pipes := []rl.Rectangle{}
	score := 0
	for i := 0; i < npipes; i++ {
		pipes = append(pipes, rl.Rectangle{})
	}
	portmidi.Initialize()
	restart := func() {
		a = 0
		v = 0
		speed = 10
		player = rl.Rectangle{}
		player.X = float32(rl.GetScreenWidth() / 2)
		player.Y = float32(rl.GetScreenHeight() / 2)
		player.Width = 100
		player.Height = 100

		for i := 0; i < npipes; i++ {
			// pipes = append(pipes, rl.Rectangle{})
			pipes[i].Y = float32(rand.Int() % rl.GetScreenHeight())
			pipes[i].X = float32(rl.GetScreenWidth() + 300*i)
			pipes[i].Width = 100
			pipes[i].Height = float32(rand.Int() % 500)
		}
		score = 0
		ioutil.WriteFile("score.txt", []byte(fmt.Sprint(highscore)), 0777)
		// portmidi.Terminate()
		rstrt <- true
	}
	movepipes := func() {
		for i := 0; i < npipes; i++ {
			pipes[i].X -= speed
			if pipes[i].X < 0-pipes[i].Width {
				pipes[i].X += float32(rl.GetScreenWidth()) + pipes[i].Width
				pipes[i].Y = float32(rand.Int() % rl.GetScreenHeight())
			}
			speed += 0.0001
			// pipes.Y = float32(rand.Int31() % rl.GetScreenHeight())
		}
	}
	showpipes := func() {
		for i := 0; i < npipes; i++ {
			rl.DrawRectangleRec(pipes[i], rl.Blue)
		}
	}
	gotHit := func() bool {
		for i := 0; i < npipes; i++ {
			if rl.CheckCollisionRecs(player, pipes[i]) == true {
				return true
			}
		}
		return false
	}
	moveScore := func() {
		score += 1
		if score > highscore {
			highscore = score
		}
	}
	restart()
	for !rl.WindowShouldClose() {
		movepipes()
		moveScore()
		if rl.IsKeyDown(rl.KeySpace) {
			a = -2
		} else if rl.IsKeyUp(rl.KeySpace) {
			a = 1
		}
		v += a
		player.Y += v
		if int(player.Y) > rl.GetScreenHeight() || player.Y < 0 {
			player.Y += float32(rl.GetScreenHeight())
			player.Y = float32(int(player.Y) % rl.GetScreenHeight())
		}
		if gotHit() {
			restart()
		}
		rl.EndTextureMode() // End drawing to texture (now we have a texture available for next passes)
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

		// Render previously generated texture using selected postpro shader
		rl.BeginShaderMode(shader)
		rl.DrawText(fmt.Sprint(score), 0, 200, 200, rl.Red)
		rl.DrawText(fmt.Sprint(highscore), 0, 0, 200, rl.Red)
		rl.DrawFPS(0, int32(rl.GetScreenHeight()-50))
		// fmt.Println(pipes[0])
		showpipes()
		rl.DrawRectangleRec(player, rl.Red)
		rl.EndShaderMode()
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
