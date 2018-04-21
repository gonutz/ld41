package main

import (
	"fmt"
	"github.com/gonutz/prototype/draw"
	"math/rand"
	"time"
)

const (
	playerSpeed          = 4
	playerW, playerH     = 172, 207
	bulletShootOffsetY   = 103
	bulletW, bulletH     = 27, 9
	zombieW, zombieH     = 116, 218
	deadHeadW, deadHeadH = 87, 103
)

type playingState struct {
	playerX, playerY int
	playerFacingLeft bool
	generator        mathGenerator
	assignment       assignment
	bullets          []bullet
	zombies          []zombie
	numbers          []fadingNumber
	nextZombie       int // time until next zombie spawns
	shootBan         int // time until shooting is allowed after wrong number
	score            int
	zombieSpawnDelay struct {
		minFrames, maxFrames int
	}
}

func (s *playingState) enter(state) {
	s.playerX = (windowW - playerW) / 2
	s.playerY = windowH - playerH - 100
	s.playerFacingLeft = false
	s.generator = mathGenerator{
		ops: []mathOp{add, subtract, add, subtract, multiply, divide},
		max: 9,
	}
	s.assignment = s.generator.generate(rand.Int)
	s.bullets = nil
	s.zombies = nil
	s.numbers = nil
	s.zombieSpawnDelay.minFrames = frames(1200 * time.Millisecond)
	s.zombieSpawnDelay.maxFrames = frames(2400 * time.Millisecond)
	s.newZombie()
	s.shootBan = 0
	s.score = 0
}

func (*playingState) leave() {}

var fireKeys = [10][2]draw.Key{
	{draw.Key0, draw.KeyNum0},
	{draw.Key1, draw.KeyNum1},
	{draw.Key2, draw.KeyNum2},
	{draw.Key3, draw.KeyNum3},
	{draw.Key4, draw.KeyNum4},
	{draw.Key5, draw.KeyNum5},
	{draw.Key6, draw.KeyNum6},
	{draw.Key7, draw.KeyNum7},
	{draw.Key8, draw.KeyNum8},
	{draw.Key9, draw.KeyNum9},
}

func (s *playingState) update(window draw.Window) state {
	// handle input
	if window.WasKeyPressed(draw.KeyEscape) {
		return menu
	}
	// shoot or miss
	s.shootBan--
	if s.shootBan < 0 {
		s.shootBan = 0
	}
	if s.shootBan <= 0 {
		wrongNumber := false
		for n, keys := range fireKeys {
			if window.WasKeyPressed(keys[0]) || window.WasKeyPressed(keys[1]) {
				if n != s.assignment.answer {
					wrongNumber = true
					s.addFadingNumber(n, draw.Red)
					s.shootBan = frames(500 * time.Millisecond)
					break
				}
			}
		}
		if !wrongNumber {
			keys := fireKeys[s.assignment.answer]
			if window.WasKeyPressed(keys[0]) || window.WasKeyPressed(keys[1]) {
				// add the number before shooting, shooting generates a new one
				s.addFadingNumber(s.assignment.answer, draw.Green)
				s.shoot(window)
			}
		}
	}
	// move left/right
	const margin = -50
	if window.IsKeyDown(draw.KeyLeft) || window.IsKeyDown(draw.KeyA) {
		s.playerX -= playerSpeed
		if s.playerX < margin {
			s.playerX = margin
		}
		s.playerFacingLeft = true
	}
	if window.IsKeyDown(draw.KeyRight) || window.IsKeyDown(draw.KeyD) {
		s.playerX += playerSpeed
		if s.playerX+playerW > windowW-margin {
			s.playerX = windowW - margin - playerW
		}
		s.playerFacingLeft = false
	}

	// update world
	// shoot bullets
	n := 0
	for i := range s.bullets {
		b := &s.bullets[i]
		bulletHitbox := rectangle{
			x: b.x,
			y: b.y,
			w: bulletW + abs(b.dx),
			h: bulletH,
		}
		if b.dx < 0 {
			bulletHitbox.x += b.dx
		}
		b.x += b.dx
		victimIndex := -1
		for i, z := range s.zombies {
			hitbox := rectangle{
				x: z.x + zombieW/4,
				y: z.y,
				w: zombieW / 2,
				h: zombieH,
			}
			if overlap(bulletHitbox, hitbox) {
				if victimIndex == -1 ||
					(b.dx > 0 && z.x < s.zombies[victimIndex].x) ||
					(b.dx < 0 && z.x > s.zombies[victimIndex].x) {
					victimIndex = i
				}
			}
		}
		if victimIndex != -1 {
			s.killZombie(victimIndex)
		}
		if victimIndex == -1 && (-100 <= b.x) && (b.x <= windowW+100) {
			s.bullets[n] = *b
			n++
		}
	}
	s.bullets = s.bullets[:n]
	// update fading numbers
	n = 0
	for i := range s.numbers {
		num := &s.numbers[i]
		num.life -= 0.02
		if num.life > 0 {
			s.numbers[n] = *num
			n++
		}
	}
	s.numbers = s.numbers[:n]
	// update zombies
	s.nextZombie--
	if s.nextZombie <= 0 {
		s.newZombie()
	}
	for i := range s.zombies {
		z := &s.zombies[i]
		if z.facingLeft {
			z.x -= 2
		} else {
			z.x += 2
		}
		const hitDist = 30
		if abs((s.playerX+playerW/2)-(z.x+zombieW/2)) < hitDist {
			return dead
		}
	}

	// render
	// background
	{
		const h = 3
		for y := 0; y < windowH; y += h {
			window.FillRect(0, y, windowW, h, draw.RGB(0, 0, float32(y+50)/windowH))
		}
		const groundH = 170
		groundCenter := draw.RGB(135/255.0, 33/255.0, 2/255.0)
		groundEdge := draw.RGB(95/255.0, 23/255.0, 1/255.0)
		for y := windowH - groundH; y < windowH; y += h {
			centerWeight := 1.0 - float32(abs(y-(windowH-groundH/2)))/80.0
			color := draw.RGB(
				groundCenter.R*centerWeight+groundEdge.R*(1-centerWeight),
				groundCenter.G*centerWeight+groundEdge.G*(1-centerWeight),
				groundCenter.B*centerWeight+groundEdge.B*(1-centerWeight),
			)
			window.FillRect(0, y, windowW, h, color)
		}
	}
	// player
	if s.playerFacingLeft {
		window.DrawImageFile(file("hero left.png"), s.playerX, s.playerY)
	} else {
		window.DrawImageFile(file("hero right.png"), s.playerX, s.playerY)
	}
	// zombies
	for _, z := range s.zombies {
		img := "zombie right.png"
		if z.facingLeft {
			img = "zombie left.png"
		}
		window.DrawImageFile(file(img), z.x, z.y)
	}
	// bullets
	for _, b := range s.bullets {
		img := "bullet left.png"
		if b.dx > 0 {
			img = "bullet right.png"
		}
		window.DrawImageFile(file(img), b.x, b.y)
	}
	// score
	{
		window.DrawImageFile(file("dead head.png"), 0, 0)
		text := romanNumeral(s.score)
		const textScale = 3
		_, h := window.GetScaledTextSize(text, textScale)
		window.DrawScaledText(text, deadHeadW, (deadHeadH-h)/2, textScale, draw.Red)
	}
	// fading numbers from the past
	for _, num := range s.numbers {
		scale := 3 + 6*(1-num.life)
		color := num.color
		color.A = num.life
		w, h := window.GetScaledTextSize(num.text, scale)
		window.DrawScaledText(num.text, (windowW-w)/2, 100-h/2, scale, color)
	}
	// assigment
	const mathScale = 2
	w, h := window.GetScaledTextSize(s.assignment.question, mathScale)
	window.DrawScaledText(
		s.assignment.question,
		s.playerX+(playerW-w)/2,
		s.playerY-2*h,
		mathScale,
		draw.White,
	)

	return playing
}

func (s *playingState) shoot(window draw.Window) {
	window.PlaySoundFile(file("shot.wav"))
	const bulletSpeed = 30
	var b bullet
	b.y = s.playerY + bulletShootOffsetY
	if s.playerFacingLeft {
		b.x = s.playerX
		b.dx = -bulletSpeed
	} else {
		b.x = s.playerX + playerW - bulletW
		b.dx = bulletSpeed
	}
	s.bullets = append(s.bullets, b)
	s.assignment = s.generator.generate(rand.Int)
}

func (s *playingState) killZombie(i int) {
	copy(s.zombies[i:], s.zombies[i+1:])
	s.zombies = s.zombies[:len(s.zombies)-1]
	s.score++
	const reduction = 0.95
	min, max := s.zombieSpawnDelay.minFrames, s.zombieSpawnDelay.maxFrames
	s.zombieSpawnDelay.minFrames = int(float32(min) * reduction)
	if s.score%2 == 1 {
		s.zombieSpawnDelay.maxFrames = int(float32(max) * reduction)
	}
}

func (s *playingState) addFadingNumber(n int, color draw.Color) {
	s.numbers = append(s.numbers, fadingNumber{
		text:  fmt.Sprintf("%d", n),
		life:  1.0,
		color: color,
	})
}

func (s *playingState) newZombie() {
	var z zombie
	z.facingLeft = rand.Intn(2) == 0
	z.y = s.playerY + playerH - zombieH - 10 + rand.Intn(30)
	if z.facingLeft {
		z.x = windowW
	} else {
		z.x = -zombieW
	}
	s.zombies = append(s.zombies, z)
	min, max := s.zombieSpawnDelay.minFrames, s.zombieSpawnDelay.maxFrames
	s.nextZombie = min + rand.Intn(max-min)
}

type fadingNumber struct {
	text  string
	life  float32
	color draw.Color
}

type bullet struct {
	x, y int
	dx   int
}

type zombie struct {
	x, y       int
	facingLeft bool
}
