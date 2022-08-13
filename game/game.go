package game

import (
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"

	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/spf13/viper"
)

const (
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	menu   DemoMenu
	paused bool

	//--create slicer and declare slices--//
	tex *TextureHandler

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64
	fullscreen   bool
	vsync        bool
	fovDegrees   float64
	fovDepth     float64

	//--viewport width / height--//
	width  int
	height int

	player *model.Player

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode      MouseMode
	mouseX, mouseY int

	// zoom settings
	zoomFovDegrees float64
	zoomFovDepth   float64

	renderDistance  float64
	clutterDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        color.NRGBA
	maxLightRGB        color.NRGBA

	// Mission and map
	mission      *model.Mission
	collisionMap []geom.Line

	sprites *SpriteHandler
	clutter *ClutterHandler

	mapWidth, mapHeight int

	debug bool
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	// initialize Game object
	g := new(Game)

	g.initConfig()

	ebiten.SetWindowTitle("PixelMek 3D")

	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(false)
	g.setVsyncEnabled(true)

	// load mission
	var err error
	missionPath := "trial.yaml"
	g.mission, err = model.LoadMission(missionPath)
	if err != nil {
		log.Println("Error loading mission", missionPath)
		log.Println(err)
		exit(1)
	}

	// load texture handler
	g.tex = NewTextureHandler(g.mission.Map())

	g.collisionMap = g.mission.Map().GetCollisionLines(clipDistance)
	worldMap := g.mission.Map().Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// init player model
	angleDegrees := 60.0
	g.player = model.NewPlayer(8.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance

	// load map and mission content once when first run
	g.loadContent()

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseMode = MouseModeMove
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mission.Map(), g.tex)
	g.camera.SetRenderDistance(g.renderDistance)

	if len(g.mission.Map().FloorBox.Image) > 0 {
		g.camera.SetFloorTexture(getTextureFromFile(g.mission.Map().FloorBox.Image))
	}
	if len(g.mission.Map().SkyBox.Image) > 0 {
		g.camera.SetSkyTexture(getTextureFromFile(g.mission.Map().SkyBox.Image))
	}

	// init camera lighting from map settings
	g.lightFalloff = g.mission.Map().Lighting.Falloff
	g.globalIllumination = g.mission.Map().Lighting.Illumination
	g.minLightRGB, g.maxLightRGB = g.mission.Map().Lighting.LightRGB()

	g.camera.SetLightFalloff(g.lightFalloff)
	g.camera.SetGlobalIllumination(g.globalIllumination)
	g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.fovDegrees = g.camera.FovAngle() // TODO: store and load from config file
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0
	g.zoomFovDegrees = g.fovDegrees / g.zoomFovDepth

	// initialize clutter
	if g.clutter != nil {
		g.clutter.Update(g, true)
	}

	// init menu system
	g.menu = mainMenu()

	return g
}

func (g *Game) initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("pixelmek")

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.pixelmek-3d"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)
	viper.SetDefault("screen.width", 1024)
	viper.SetDefault("screen.height", 768)
	viper.SetDefault("screen.renderScale", 1.0)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.clutterDistance", 10.0)

	err := viper.ReadInConfig()
	if err != nil && g.debug {
		log.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.clutterDistance = viper.GetFloat64("screen.clutterDistance")
	g.debug = viper.GetBool("debug")
}

func (g *Game) SaveConfig() error {
	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.pixelmek-3d"

	userConfig := userConfigPath + "/config.json"
	log.Print("Saving config file ", userConfig)

	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	err := viper.WriteConfigAs(userConfig)
	if err != nil {
		log.Print(err)
	}

	return err
}

// Run is the Ebiten Run loop caller
func (g *Game) Run() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}

	g.paused = false

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.screenWidth)*g.renderScale), int(float64(g.screenHeight)*g.renderScale)
	g.menu.layout(w, h)
	return int(w), int(h)
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// handle input (when paused making sure only to allow input for closing menu so it can be unpaused)
	g.handleInput()

	if !g.paused {
		// Perform logical updates
		g.updateProjectiles()
		g.updateSprites()

		if g.clutter != nil {
			g.clutter.Update(g, false)
		}

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.update(g)

	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	raycastSprites := g.getRaycastSprites()

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast to screen
	g.camera.Draw(screen)

	// draw menu (if active)
	g.menu.draw(screen)
}

func (g *Game) setFullscreen(fullscreen bool) {
	g.fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

func (g *Game) setResolution(screenWidth, screenHeight int) {
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.renderScale)
}

func (g *Game) setRenderScale(renderScale float64) {
	g.renderScale = renderScale
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
}

func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	if enableVsync {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	} else {
		ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	}
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	pi2 := geom.Pi2
	if g.player.Angle >= pi2 {
		g.player.Angle = pi2 - g.player.Angle
	} else if g.player.Angle <= -pi2 {
		g.player.Angle = g.player.Angle + pi2
	}

	g.player.Moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	g.player.Pitch += pSpeed

	// current raycasting method can only allow up to 45 degree pitch in either direction
	g.player.Pitch = geom.Clamp(g.player.Pitch, -math.Pi/4, math.Pi/4)

	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.PositionZ = 0.5
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.PosZ() == 0.5
}

func (g *Game) Jump() {
	g.player.PositionZ = 0.9
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.PositionZ = 0.3
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.PositionZ = 0.1
	g.player.Moved = true
}

func (g *Game) fireWeapon() {
	p := g.player.TestProjectile
	if p == nil {
		return
	}

	if g.player.TestCooldown > 0 {
		return
	}

	// spawning projectile at player position just slightly below player's center point of view
	pX, pY, pZ := g.player.Position.X, g.player.Position.Y, geom.Clamp(g.player.PositionZ-0.15, 0.05, g.player.PositionZ+0.5)
	// TODO: pitch angle should be based on raycasted angle toward crosshairs, for now just simplified as player pitch angle
	pAngle, pPitch := g.player.Angle, g.player.Pitch

	pVelocity := 24.0
	projectile := p.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, pVelocity, g.player.Entity)
	if projectile != nil {
		g.sprites.addProjectile(projectile)
		g.player.TestCooldown = 10
	}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	playerPos := g.player.Position.Copy()
	playerPosZ := (g.player.PositionZ - 0.5) * float64(g.height)

	g.camera.SetPosition(playerPos)
	g.camera.SetPositionZ(playerPosZ)
	g.camera.SetHeadingAngle(g.player.Angle)
	g.camera.SetPitchAngle(g.player.Pitch)
}

func (g *Game) updateProjectiles() {
	// Testing animated projectile movement
	if g.player.TestCooldown > 0 {
		g.player.TestCooldown--
	}

	for p := range g.sprites.projectiles {
		if p.Velocity != 0 {

			realVelocity := p.Velocity
			zVelocity := 0.0
			if p.Pitch != 0 {
				// would be better to use proper 3D geometry math here, but trying to avoid matrix math library
				// for this one simple use (but if becomes desired: https://github.com/ungerik/go3d)
				realVelocity = geom.GetAdjacentHypotenuseTriangleLeg(p.Pitch, p.Velocity)
				zVelocity = geom.LineFromAngle(0, 0, p.Pitch, realVelocity).Y2
			}

			vLine := geom.LineFromAngle(p.Position.X, p.Position.Y, p.Angle, realVelocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2

			// TODO: getValidMove needs to be able to take PosZ into account for wall/sprite collisions
			newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, false)
			if isCollision || p.PositionZ <= 0 {
				// for testing purposes, projectiles instantly get deleted when collision occurs
				g.sprites.deleteProjectile(p)

				// make a sprite/wall getting hit by projectile cause some visual effect
				if p.ImpactEffect.Sprite != nil {
					if len(collisions) >= 1 {
						// use the first collision point to place effect at
						newPos = collisions[0].collision
					}

					// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
					effect := p.SpawnEffect(newPos.X, newPos.Y, p.PositionZ, p.Angle, p.Pitch)

					g.sprites.addEffect(effect)
				}

				for _, collisionEntity := range collisions {
					if collisionEntity.entity == g.player.Entity {
						println("ouch!")
					} else {
						println("hit!")
					}
				}
			} else {
				p.Position = newPos

				if zVelocity != 0 {
					p.PositionZ += zVelocity
				}
			}
		}
		p.Update(g.player.Position)
	}

	// Testing animated effects (explosions)
	for e := range g.sprites.effects {
		e.Update(g.player.Position)
		if e.GetLoopCounter() >= e.LoopCount {
			g.sprites.deleteEffect(e)
		}
	}
}

func (g *Game) AllEntities() map[*model.Entity]struct{} {
	numEntities := len(g.sprites.mapSprites) + len(g.sprites.mechSprites)
	entities := make(map[*model.Entity]struct{}, numEntities)
	for s := range g.sprites.mapSprites {
		entities[s.Entity] = struct{}{}
	}
	for s := range g.sprites.mechSprites {
		entities[s.Entity] = struct{}{}
	}
	return entities
}

func (g *Game) updateSprites() {
	// Testing animated sprite movement
	for s := range g.sprites.mapSprites {
		g.updateSpritePosition(s)
		s.Update(g.player.Position)
	}

	// Testing animated mech sprite movement
	for s := range g.sprites.mechSprites {
		g.updateMechPosition(s)
		s.Update(g.player.Position)
	}
}

func (g *Game) updateMechPosition(s *model.MechSprite) {
	// TODO: give mechs a bit more of a brain than this
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: s.Position.X, Y1: s.Position.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.Velocity = 0.02

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateMechPosition(s)
		} else {
			// keep movements towards current patrol point
			s.Angle = angle
		}
	}

	if s.Velocity != 0 {
		vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.Angle = randFloat(-math.Pi, math.Pi)
			s.Velocity = randFloat(0.005, 0.009)
		} else {
			s.Position = newPos
		}
	}
}

func (g *Game) updateSpritePosition(s *model.Sprite) {
	if s.Velocity != 0 {
		vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.Angle = randFloat(-math.Pi, math.Pi)
			s.Velocity = randFloat(0.005, 0.009)
		} else {
			s.Position = newPos
		}
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	os.Exit(rc)
}
