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
	playerHeadH          = 60
	bulletShootOffsetY   = 103
	bulletW, bulletH     = 27, 9
	zombieW, zombieH     = 116, 218
	deadHeadW, deadHeadH = 87, 103
	zombieSpawnReduction = 0.97
	zombieSpawnMin       = 1000 * time.Millisecond
	zombieSpawnMax       = 2000 * time.Millisecond
	playerWalkFrames     = 4
	bloodW, bloodH       = 24, 20
)

type torsoState int

const (
	idle torsoState = iota
	reloading
	waitingToReload
	shooting
	aimingAtHead
	bleeding
)

func dying(s torsoState) bool {
	return s >= aimingAtHead
}

type playingState struct {
	playerX, playerY int
	playerFacingLeft bool
	playerWalkFrame  int
	playerWalkTime   int
	generator        mathGenerator
	assignment       assignment
	bullets          []bullet
	zombies          []zombie
	numbers          []fadingNumber
	nextZombie       int // time until next zombie spawns
	shootBan         int // time until shooting is allowed after wrong number
	score            int
	zombieSpawnDelay struct {
		minFrames, maxFrames float32
	}
	torso          torsoState
	torsoTime      int
	blood          []bloodParticle
	leaveStateTime int
}

func (s *playingState) enter(state) {
	s.playerX = (windowW - playerW) / 2
	s.playerY = windowH - playerH - 100
	s.playerFacingLeft = false
	s.playerWalkFrame = 0
	s.playerWalkTime = 0
	s.generator = mathGenerator{
		ops: []mathOp{add, subtract, add, subtract, multiply, divide},
		max: 9,
	}
	s.assignment = s.generator.generate(rand.Int)
	s.bullets = nil
	s.zombies = nil
	s.numbers = nil
	s.nextZombie = 0
	s.shootBan = 0
	s.score = 0
	s.zombieSpawnDelay.minFrames = float32(frames(zombieSpawnMin))
	s.zombieSpawnDelay.maxFrames = float32(frames(zombieSpawnMax))
	s.newZombie()
	s.torso = idle
	s.torsoTime = 0
	s.blood = nil
	s.leaveStateTime = -1
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
		if dying(s.torso) {
			return dead
		} else {
			return menu
		}
	}
	// shoot or miss
	s.shootBan--
	if s.shootBan < 0 {
		s.shootBan = 0
	}
	if !dying(s.torso) && s.shootBan <= 0 {
		wrongNumber := false
		for n, keys := range fireKeys {
			if window.WasKeyPressed(keys[0]) || window.WasKeyPressed(keys[1]) {
				if n != s.assignment.answer {
					wrongNumber = true
					window.PlaySoundFile(file("miss shot.wav"))
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
	walking := false
	if !dying(s.torso) {
		const margin = -50
		if window.IsKeyDown(draw.KeyLeft) || window.IsKeyDown(draw.KeyA) {
			walking = true
			s.playerX -= playerSpeed
			if s.playerX < margin {
				s.playerX = margin
			}
			s.playerFacingLeft = true
		} else if window.IsKeyDown(draw.KeyRight) || window.IsKeyDown(draw.KeyD) {
			walking = true
			s.playerX += playerSpeed
			if s.playerX+playerW > windowW-margin {
				s.playerX = windowW - margin - playerW
			}
			s.playerFacingLeft = false
		}
	}
	if walking {
		s.playerWalkTime--
		if s.playerWalkTime <= 0 {
			s.playerWalkFrame = (s.playerWalkFrame + 1) % playerWalkFrames
			s.playerWalkTime = frames(100 * time.Millisecond)
		}
	} else {
		s.playerWalkFrame = 0
		s.playerWalkTime = 0
	}

	// update world
	if s.leaveStateTime > 0 {
		s.leaveStateTime--
		if s.leaveStateTime <= 0 {
			return dead
		}
	}
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
			window.PlaySoundFile(file("zombie death.wav"))
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
	if !dying(s.torso) {
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
			const hitDist = 40
			if abs((s.playerX+playerW/2)-(z.x+zombieW/2)) < hitDist {
				s.torso = aimingAtHead
				s.torsoTime = frames(time.Second)
			}
			const zombieFrameCount = 4
			z.nextFrame--
			if z.nextFrame <= 0 {
				z.nextFrame = frames(250 * time.Millisecond)
				z.frame = (z.frame + 1) % zombieFrameCount
			}
		}
	}
	// update blood and gore
	{
		n := 0
		for i := range s.blood {
			b := &s.blood[i]
			b.x += b.vx
			b.y += b.vy
			b.rotation += b.dRotation
			b.vy += 0.5
			if b.y < windowH {
				s.blood[n] = *b
				n++
			}
		}
		s.blood = s.blood[:n]
	}
	// animations
	if s.torsoTime > 0 {
		s.torsoTime--
		if s.torsoTime == 0 {
			switch s.torso {
			case idle:
				// nothing to do in this case
			case shooting:
				s.torso = waitingToReload
				s.torsoTime = frames(200 * time.Millisecond)
			case reloading:
				s.torso = idle
			case waitingToReload:
				s.torso = reloading
				s.torsoTime = frames(250 * time.Millisecond)
				window.PlaySoundFile(file("reload.wav"))
			case aimingAtHead:
				s.torso = bleeding
				window.PlaySoundFile(file("shot.wav"))
				x, y := s.playerNeck()
				s.sprayBlood(x, y, 100, 200)
				s.torsoTime = frames(50 * time.Millisecond)
				s.leaveStateTime = frames(3 * time.Second)
			case bleeding:
				// nothing to do in this case
				s.torsoTime = frames(50 * time.Millisecond)
				x, y := s.playerNeck()
				s.sprayBlood(x, y, 5, 10)
			}
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
	hero := "hero "
	if s.torso == reloading {
		hero += "reload "
	}
	if s.torso == shooting {
		hero += "shoot "
	}
	if s.torso == aimingAtHead {
		hero += "aiming at head "
	}
	if s.torso == bleeding {
		hero += "bleeding head "
	}
	dir := "right"
	if s.playerFacingLeft {
		dir = "left"
	}
	hero += dir
	hero += ".png"
	window.DrawImageFile(file(hero), s.playerX, s.playerY)
	if s.shootBan > 0 {
		window.DrawImageFile(file("hero eye blink "+dir+".png"), s.playerX, s.playerY)
	}
	if walking {
		img := fmt.Sprintf("hero legs walk %s %d.png", dir, s.playerWalkFrame)
		window.DrawImageFile(file(img), s.playerX, s.playerY)
	} else {
		window.DrawImageFile(file("hero legs stand "+dir+".png"), s.playerX, s.playerY)
	}
	// zombies
	for _, z := range s.zombies {
		dir := "right"
		if z.facingLeft {
			dir = "left"
		}
		var img string
		if dying(s.torso) {
			img = fmt.Sprintf("zombie %d %s.png", z.kind, dir)
		} else {
			img = fmt.Sprintf("zombie %d %s %d.png", z.kind, dir, z.frame)
		}
		window.DrawImageFile(file(img), z.x, z.y)
	}
	// blood and gore
	for i := range s.blood {
		b := &s.blood[i]
		window.DrawImageFileRotated(
			file("blood particle.png"),
			round(b.x),
			round(b.y),
			round(b.rotation),
		)
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
	s.torso = shooting
	s.torsoTime = frames(100 * time.Millisecond)
}

func (s *playingState) killZombie(i int) {
	// spray blood
	z := s.zombies[i]
	cx, cy := z.x+zombieW/2, z.y+zombieH/2
	s.sprayBlood(cx, cy, 10, 30)

	// remove zombie from list
	copy(s.zombies[i:], s.zombies[i+1:])
	s.zombies = s.zombies[:len(s.zombies)-1]
	s.score++
	min, max := s.zombieSpawnDelay.minFrames, s.zombieSpawnDelay.maxFrames
	s.zombieSpawnDelay.minFrames = min * zombieSpawnReduction
	if s.score%2 == 1 {
		s.zombieSpawnDelay.maxFrames = max * zombieSpawnReduction
	}
}

func (s *playingState) sprayBlood(x, y, min, max int) {
	count := min + rand.Intn(max-min)
	for i := 0; i < count; i++ {
		s.blood = append(s.blood, bloodParticle{
			x:         float32(x - bloodW/2),
			y:         float32(y - bloodH/2),
			vx:        3 - 6*rand.Float32(),
			vy:        -10 - 5*rand.Float32(),
			rotation:  360 * rand.Float32(),
			dRotation: 2 - 4*rand.Float32(),
		})
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
	const zombieKindCount = 3
	z.kind = rand.Intn(zombieKindCount)
	s.zombies = append(s.zombies, z)
	min := round(s.zombieSpawnDelay.minFrames)
	max := round(s.zombieSpawnDelay.maxFrames)
	s.nextZombie = min + rand.Intn(max-min)
}

func (s *playingState) playerNeck() (x, y int) {
	dx := -6
	if s.playerFacingLeft {
		dx = -dx
	}
	return s.playerX + playerW/2 + dx, s.playerY + playerHeadH
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
	frame      int
	nextFrame  int
	kind       int
}

type bloodParticle struct {
	x, y      float32
	vx, vy    float32
	rotation  float32
	dRotation float32
}
