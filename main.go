package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/YiYeYu/LazySin/camera"
	"github.com/YiYeYu/LazySin/draw"
)

const (
	fps        = 2
	width      = 640
	height     = 480
	row        = 10
	col        = 10
	halfWidth  = width / 2
	halfHeight = height / 2
	tw         = width / col
	th         = height / row

	aliveRate = 0.3
)

var (
	v0_0 = draw.Vertex{X: 0, Y: 0}
	v0_1 = draw.Vertex{X: 0, Y: 1}
	v1_1 = draw.Vertex{X: 1, Y: 1}
	v1_0 = draw.Vertex{X: 1, Y: 0}
)

type Tile struct {
	X       int32
	Y       int32
	Vertexs []draw.Vertex
	draw.Color
	alive     bool
	aliveNext bool
}

func (t *Tile) Draw(r *draw.Render) {
	if !t.alive {
		return
	}
	r.SetColor(t.Color)
	r.PutVertexs(t.Vertexs)
}

//var tile = Tile{Vertexs: []draw.Vertex{v0_0, v0_1, v1_1, v1_0}}

func main() {
	runtime.LockOSThread() //very important, gl need it

	//date fmt string 2006_01_02_15_04_05
	logPath := fmt.Sprintf("log/log_%v.log", time.Now().Format("2006_01_02_15"))
	f, e := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC,
		0755)
	defer f.Close()
	if e != nil {
		fmt.Printf("log error: %v\n", e)
	} else {
		fmt.Printf("log sucess: file[%v]\n", logPath)
		//os.Stdout = f
		os.Stderr = f
		log.SetOutput(f)
		//fmt.Printf("log start: file[%v]\n", logPath)
	}

	if e := glfw.Init(); e != nil {
		log.Fatalf("init glfw failed\n")
		return
	}

	window, e := glfw.CreateWindow(width, height, "Title", nil, nil)
	defer glfw.Terminate()
	if e != nil || window == nil {
		log.Fatalf("init window failed\n")
		return
	}
	window.MakeContextCurrent()

	prog, e := initOpenGL()
	if e != nil {
		log.Fatalf("init gl failed\n")
		return
	}

	r := &draw.Render{}
	camera := &camera.Camera{}

	camera.PrepareViewPort(0, 0, width, height)
	camera.SetCamera2D(-halfWidth, halfWidth, -halfHeight, halfHeight)

	i := 0
	t := time.Now()
	log.Printf("seed: %v\n", t.Nanosecond())
	rand.Seed(int64(t.Nanosecond()))

	tiles := makeTiles()

	for !window.ShouldClose() {
		i++
		//log.Printf("loop: %v\n", i)

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.UseProgram(prog)

		//
		gl.Begin(gl.QUADS)

		for i, tile := range tiles {
			_ = i
			// if i != 0*row+0 {
			// 	continue
			// }
			tile.Draw(r)
		}
		//r.PutVertexs(tile.Vertexs)

		gl.End()

		//
		r.Flush()

		window.SwapBuffers()
		glfw.PollEvents() //debug mode may crash, why???

		checkAlive(tiles)
		nextLoop(tiles)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
		t = time.Now()
	}
	log.Println("filnish")
	fmt.Println("filnish")
}

// initOpenGL 初始化 OpenGL 并且返回一个初始化了的程序。
func initOpenGL() (uint32, error) {
	if err := gl.Init(); err != nil {
		log.Fatalf("init gl failed\n")
		return 0, fmt.Errorf("init gl failed")
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)
	prog := gl.CreateProgram()
	gl.LinkProgram(prog)
	return prog, nil
}

func makeTiles() []*Tile {
	tiles := make([]*Tile, row*col)
	for i := 0; i < row; i++ {
		for j := 0; j < col; j++ {
			tiles[i*row+j] = makeTile(int32(j), int32(i))
		}
	}
	//log.Printf("tiles: %v", tiles)
	return tiles
}

func makeTile(x, y int32) *Tile {
	v0 := draw.Vertex{X: tw*x - halfWidth, Y: th*y - halfHeight}
	v1 := draw.Vertex{X: tw*x - halfWidth, Y: th*(y+1) - halfHeight}
	v2 := draw.Vertex{X: tw*(x+1) - halfWidth, Y: th*(y+1) - halfHeight}
	v3 := draw.Vertex{X: tw*(x+1) - halfWidth, Y: th*y - halfHeight}
	color := draw.Color{
		R: 128 + uint8(rand.Intn(128)),
		G: 128 + uint8(rand.Intn(128)),
		B: 128 + uint8(rand.Intn(128)),
		A: 255}
	alive := rand.Float32() < aliveRate
	return &Tile{X: x, Y: y, Vertexs: []draw.Vertex{v0, v1, v2, v3}, Color: color, alive: alive}
}

func checkAlive(tiles []*Tile) {
	for _, tile := range tiles {
		cnt := countAliveNeighbor(tile, tiles)
		if tile.alive {
			// 1. 当任何一个存活的 cell 的附近少于 2 个存活的 cell 时，该 cell 将会消亡，就像人口过少所导致的结果一样
			// 2. 当任何一个存活的 cell 的附近有 2 至 3 个存活的 cell 时，该 cell 在下一代中仍然存活
			// 3. 当任何一个存活的 cell 的附近多于 3 个存活的 cell 时，该 cell 将会消亡，就像人口过多所导致的结果一样
			// if cnt < 2 {
			// 	tile.aliveNext = false
			// } else if cnt >= 2 && cnt <= 3 {
			// 	tile.aliveNext = true
			// } else if cnt > 3 {
			// 	tile.aliveNext = false
			// }
			tile.aliveNext = cnt >= 2 && cnt <= 3
		} else {
			// 4. 任何一个消亡的 cell 附近刚好有 3 个存活的 cell，该 cell 会变为存活的状态，就像重生一样。
			if cnt == 3 {
				tile.aliveNext = true
			}
		}
	}
}

func countAliveNeighbor(tile *Tile, tiles []*Tile) (cnt int32) {
	cnt = 0
	tryadd := func(x, y int32) {
		if x < 0 {
			x = row - 1
		}
		if x >= row {
			x = 0
		}
		if y < 0 {
			y = col - 1
		}
		if y >= col {
			y = 0
		}
		if tiles[x+y*row].alive {
			cnt++
		}
	}
	tryadd(tile.X-1, tile.Y-1) //bottomleft
	tryadd(tile.X-1, tile.Y)   //left
	tryadd(tile.X-1, tile.Y+1) //topleft
	tryadd(tile.X, tile.Y+1)   //top
	tryadd(tile.X+1, tile.Y+1) //topright
	tryadd(tile.X+1, tile.Y)   //right
	tryadd(tile.X+1, tile.Y-1) //bottomright
	tryadd(tile.X, tile.Y-1)   //bottom
	return
}

func nextLoop(tiles []*Tile) {
	i := 0
	for _, tile := range tiles {
		tile.alive = tile.aliveNext
		if tile.alive {
			i++
		}
	}
	fmt.Printf("alives: %v\n", i)
}
