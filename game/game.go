package game

import (
	"fmt"
	"image"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"

	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/spf13/viper"
)

const (
	title = "PixelMek 3D"
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

// Game - This is the main type for your game.
type Game struct {
	menu   GameMenu
	paused bool

	resources *model.ModelResources

	//--create slicer and declare slices--//
	tex                *TextureHandler
	initRenderFloorTex bool

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

	player        *Player
	playerStatus  *render.UnitStatus
	targetStatus  *render.UnitStatus
	navStatus     *render.NavStatus
	armament      *render.Armament
	compass       *render.Compass
	altimeter     *render.Altimeter
	heat          *render.HeatIndicator
	radar         *render.Radar
	throttle      *render.Throttle
	jets          *render.JumpJetIndicator
	crosshairs    *render.Crosshairs
	targetReticle *render.TargetReticle
	navReticle    *render.NavReticle
	fonts         *render.FontHandler

	hudEnabled        bool
	hudScale          float64
	hudRGBA           color.RGBA
	hudUseCustomColor bool

	//--define camera and renderer--//
	camera *raycaster.Camera

	mouseMode      MouseMode
	mouseX, mouseY int

	// zoom settings
	zoomFovDepth float64

	renderDistance  float64
	clutterDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        color.NRGBA
	maxLightRGB        color.NRGBA

	// Mission and map
	mission      *model.Mission
	collisionMap []*geom.Line

	sprites                *SpriteHandler
	clutter                *ClutterHandler
	collisonSpriteTypes    map[SpriteType]struct{}
	interactiveSpriteTypes map[SpriteType]struct{}
	delayedProjectiles     map[*DelayedProjectileSpawn]struct{}

	mapWidth, mapHeight int

	// control options
	throttleDecay bool

	debug           bool
	debugFpsCounter int
}

type TargetCycleType int

const (
	TARGET_NEXT TargetCycleType = iota
	TARGET_PREVIOUS
	TARGET_NEAREST
)

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	// initialize Game object
	g := new(Game)
	g.fonts = render.NewFontHandler()

	g.initConfig()

	g.initInteractiveTypes()
	g.initCollisionTypes()
	g.initCombatVariables()

	ebiten.SetWindowTitle(title)
	ebiten.SetTPS(int(model.TICKS_PER_SECOND))

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(g.fullscreen)
	g.setVsyncEnabled(g.vsync)

	var err error
	g.resources, err = model.LoadModelResources()
	if err != nil {
		log.Println("Error loading models:")
		log.Println(err)
		exit(1)
	}

	// load mission // TODO: mission select UI
	missionPath := "trial.yaml"
	//missionPath := "debug.yaml"
	g.mission, err = model.LoadMission(missionPath)
	if err != nil {
		log.Println("Error loading mission: ", missionPath)
		log.Println(err)
		exit(1)
	}

	// load texture handler
	g.tex = NewTextureHandler(g.mission.Map())
	g.tex.renderFloorTex = g.initRenderFloorTex

	g.collisionMap = g.mission.Map().GetCollisionLines(clipDistance)
	worldMap := g.mission.Map().Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// load map and mission content once when first run
	g.loadContent()

	// init player model
	pX, pY, pDegrees := 8.5, 3.5, 60.0 // TODO: get from mission
	pUnit := g.SetPlayerUnit(model.MechResourceType, "timberwolf_prime.yaml")
	//pUnit := g.SetPlayerUnit(model.MechResourceType, "jenner_iic.yaml")
	//pUnit := g.createModelInfantry("heavy_foot.yaml")
	//pUnit := g.createModelVehicle("srm_carrier.yaml")
	//pUnit:= g.createModelVTOL("donar.yaml")

	pUnit.SetPos(&geom.Vector2{X: pX, Y: pY})
	pUnit.SetHeading(geom.Radians(pDegrees))

	// init mouse movement mode
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mission.Map(), g.tex)
	g.camera.SetRenderDistance(g.renderDistance)
	g.camera.SetAlwaysSetSpriteScreenRect(true)

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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

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
	viper.SetDefault("screen.fullscreen", false)
	viper.SetDefault("screen.vsync", true)
	viper.SetDefault("screen.renderFloor", true)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.clutterDistance", 10.0)

	viper.SetDefault("hud.enabled", true)
	viper.SetDefault("hud.scale", 1.0)
	viper.SetDefault("hud.font", "pixeloid.otf")
	viper.SetDefault("hud.color.useCustom", false)
	viper.SetDefault("hud.color.red", 100)
	viper.SetDefault("hud.color.green", 255)
	viper.SetDefault("hud.color.blue", 230)
	viper.SetDefault("hud.color.alpha", 255)

	viper.SetDefault("controls.throttleDecay", false)

	err := viper.ReadInConfig()
	if err != nil && g.debug {
		log.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.fullscreen = viper.GetBool("screen.fullscreen")
	g.vsync = viper.GetBool("screen.vsync")
	g.initRenderFloorTex = viper.GetBool("screen.renderFloor")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.clutterDistance = viper.GetFloat64("screen.clutterDistance")

	g.fonts.HUDFont, err = g.fonts.LoadFont(viper.GetString("hud.font"))
	if err != nil {
		log.Fatal(err)
		exit(1)
	}
	g.hudEnabled = viper.GetBool("hud.enabled")
	g.hudScale = viper.GetFloat64("hud.scale")
	g.hudUseCustomColor = viper.GetBool("hud.color.useCustom")
	g.hudRGBA = color.RGBA{
		R: uint8(viper.GetUint("hud.color.red")),
		G: uint8(viper.GetUint("hud.color.green")),
		B: uint8(viper.GetUint("hud.color.blue")),
		A: uint8(viper.GetUint("hud.color.alpha")),
	}

	g.throttleDecay = viper.GetBool("controls.throttleDecay")

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
		g.updatePlayer()
		g.updateProjectiles()
		g.updateSprites()

		if g.clutter != nil {
			g.clutter.Update(g, false)
		}

		// handle player weapon updates
		g.updateWeaponCooldowns(g.player.Unit)

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.update(g)

	if g.debug {
		if g.debugFpsCounter == 0 {
			g.debugFpsCounter = ebiten.TPS()
			// draw FPS/TPS counter debug display
			fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
			ebiten.SetWindowTitle(title + " | " + fps)
		} else {
			g.debugFpsCounter--
		}
	}

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

	// store raycasted convergence point for next Update
	g.player.convergenceDistance = g.camera.GetConvergenceDistance()
	g.player.convergencePoint = g.camera.GetConvergencePoint()
	g.player.convergenceSprite = getSpriteFromInterface(g.camera.GetConvergenceSprite())

	// draw HUD elements
	g.drawHUD(screen)

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
	playerPosition := g.player.Pos()
	moveLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Heading(), mSpeed)
	g.updatePlayerPosition(moveLine.X2, moveLine.Y2, g.player.PosZ())
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	playerPosition := g.player.Pos()
	strafeLine := geom.LineFromAngle(playerPosition.X, playerPosition.Y, g.player.Heading()-strafeAngle, math.Abs(sSpeed))
	g.updatePlayerPosition(strafeLine.X2, strafeLine.Y2, g.player.PosZ())
}

// Move player by vertical speed in the up/down direction
func (g *Game) VerticalMove(vSpeed float64) {
	pos := g.player.Pos()
	newPosZ := g.player.PosZ() + vSpeed
	g.updatePlayerPosition(pos.X, pos.Y, newPosZ)
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	angle := g.player.Heading() + rSpeed

	if angle >= geom.Pi2 {
		angle = geom.Pi2 - angle
	} else if angle < 0 {
		angle = angle + geom.Pi2
	}

	if angle < 0 {
		// handle rounding errors
		angle = 0
	}

	g.player.SetHeading(angle)
	g.player.moved = true
}

// Rotate player turret angle, relative to body heading, by rotation speed
func (g *Game) RotateTurret(rSpeed float64) {
	if !g.player.HasTurret() {
		g.Rotate(rSpeed)
		return
	}

	angle := g.player.TurretAngle() + rSpeed

	if angle > geom.HalfPi {
		angle = geom.HalfPi
	} else if angle < -geom.HalfPi {
		angle = -geom.HalfPi
	}

	g.player.SetTurretAngle(angle)
	g.player.moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	// current raycasting method can only allow up to 45 degree pitch in either direction
	g.player.SetPitch(geom.Clamp(pSpeed+g.player.Pitch(), -geom.Pi/8, geom.Pi/4))
	g.player.moved = true
}

func (g *Game) updatePlayer() {
	if g.player.Update() {

		// TODO: refactor to use same update position function as sprites (g.updateMechPosition, etc.)

		position, posZ := g.player.Pos(), g.player.PosZ()
		velocity, velocityZ := g.player.Velocity(), g.player.VelocityZ()
		if velocityZ != 0 {
			posZ += velocityZ
		}

		moveHeading := g.player.Heading()
		if g.player.JumpJetsActive() || (posZ > 0 && g.player.JumpJets() > 0) {
			// while jumping, or still in air after jumping, continue from last jump jet active heading and velocity
			moveHeading = g.player.JumpJetHeading()
			velocity = g.player.JumpJetVelocity()
		}
		moveLine := geom.LineFromAngle(position.X, position.Y, moveHeading, velocity)

		g.updatePlayerPosition(moveLine.X2, moveLine.Y2, posZ)
		g.player.moved = true
	}

	target := g.player.Target()
	if target != nil {
		// only increment lock percent on target if reticle near target area and in weapon range
		s := g.getSpriteFromEntity(target)
		if s == nil {
			return
		}

		acquireLock := false
		crosshairLockSize := int(math.Ceil(float64(g.width) * 0.05))
		midW, midH := g.width/2, g.height/2
		crosshairBounds := image.Rect(
			midW-crosshairLockSize/2, midH-crosshairLockSize/2,
			midW+crosshairLockSize/2, midH+crosshairLockSize/2,
		)
		targetBounds := s.ScreenRect()
		if targetBounds != nil {
			acquireLock = targetBounds.Overlaps(crosshairBounds)
		}

		targetDistance := model.EntityDistance(g.player, target) - g.player.CollisionRadius() - target.CollisionRadius()
		lockOnRange := 1000.0 / model.METERS_PER_UNIT

		if int(targetDistance) <= int(lockOnRange) {
			// TODO: decrease lock percent delta if further from target
			lockDelta := 0.25 / model.TICKS_PER_SECOND
			if !acquireLock {
				lockDelta = -0.15 / model.TICKS_PER_SECOND
			}

			targetLock := g.player.TargetLock() + lockDelta
			if targetLock > 1.0 {
				targetLock = 1.0
			} else if targetLock < 0 {
				targetLock = 0
			}
			g.player.SetTargetLock(targetLock)
		}
	}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.moved = false

	g.camera.SetPosition(g.player.Pos().Copy())
	g.camera.SetPositionZ(g.player.cameraZ)
	g.camera.SetPitchAngle(g.player.Pitch())

	if g.player.HasTurret() {
		g.camera.SetHeadingAngle(g.player.Heading() + g.player.TurretAngle())
	} else {
		g.camera.SetHeadingAngle(g.player.Heading())
	}
}

func (g *Game) updatePlayerPosition(setX, setY, setZ float64) {
	// Update player position
	newPos, newZ, isCollision, collisions := g.getValidMove(g.player.Unit, setX, setY, setZ, true)
	if !(newPos.Equals(g.player.Pos()) && newZ == g.player.PosZ()) {
		g.player.SetPos(newPos)
		g.player.SetPosZ(newZ)
		g.player.moved = true
	}

	if isCollision && len(collisions) > 0 {
		// apply damage to the first sprite entity that was hit
		collisionEntity := collisions[0]

		//fmt.Printf("collided with entity at %v (z: %v)\n", collisionEntity.entity.Pos(), collisionEntity.entity.PosZ())
		collisionDamage := 0.1 // TODO: determine collision damage based on player mech and speed
		collisionEntity.entity.ApplyDamage(collisionDamage)
		fmt.Printf("collided for %0.1f (HP: %0.1f)\n", collisionDamage, collisionEntity.entity.ArmorPoints())
	}
}

func (g *Game) navPointCycle() {
	if len(g.mission.NavPoints) == 0 {
		return
	}

	if g.player.Target() != nil {
		// unset player target so status display can show nav selection
		g.player.SetTarget(nil)
		if g.player.navPoint != nil {
			// on the first time after unset target have it not cycle to next nav
			return
		}
	}

	var newNav *model.NavPoint
	navPoints := g.mission.NavPoints
	currentNav := g.player.navPoint

	for _, n := range navPoints {
		if currentNav == nil {
			newNav = n
			break
		}

		if currentNav.NavPoint == n {
			// allow next loop iteration to select as new nav point
			currentNav = nil
			continue
		}
	}

	if newNav == nil {
		newNav = navPoints[0]
	}

	g.player.navPoint = render.NewNavSprite(newNav, 1.0)
}

func (g *Game) targetCrosshairs() {
	newTarget := g.player.convergenceSprite
	if newTarget != nil {
		g.player.SetTarget(newTarget.Entity)
	}
}

func (g *Game) targetCycle(cycleType TargetCycleType) {
	targetables := make([]*render.Sprite, 0, 64)

	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {
			if !g.isInteractiveType(spriteType) {
				// only cycle on certain sprite types (skip projectiles, effects, etc.)
				return true
			}

			s := getSpriteFromInterface(k.(raycaster.Sprite))
			targetables = append(targetables, s)

			return true
		})
	}

	if len(targetables) == 0 {
		g.player.SetTarget(nil)
		return
	}

	// sort by rough estimate of distance to player
	playerPos := g.player.Pos()

	if cycleType == TARGET_PREVIOUS {
		sort.Slice(targetables, func(a, b int) bool {
			sA, sB := targetables[a], targetables[b]
			dA := math.Abs(sA.Pos().X-playerPos.X) + math.Abs(sA.Pos().Y-playerPos.Y)
			dB := math.Abs(sB.Pos().X-playerPos.X) + math.Abs(sB.Pos().Y-playerPos.Y)
			return dA > dB
		})
	} else {
		sort.Slice(targetables, func(a, b int) bool {
			sA, sB := targetables[a], targetables[b]
			dA := math.Abs(sA.Pos().X-playerPos.X) + math.Abs(sA.Pos().Y-playerPos.Y)
			dB := math.Abs(sB.Pos().X-playerPos.X) + math.Abs(sB.Pos().Y-playerPos.Y)
			return dA < dB
		})
	}

	var newTarget *render.Sprite

	if cycleType != TARGET_NEAREST {
		currentTarget := g.player.Target()
		for _, t := range targetables {
			if currentTarget == nil {
				newTarget = t
				break
			}

			if currentTarget == t.Entity {
				// allow next loop iteration to select as new target
				currentTarget = nil
				continue
			}
		}
	}

	if newTarget == nil {
		newTarget = targetables[0]
	}

	g.player.SetTarget(newTarget.Entity)
}

func (g *Game) updateSprites() {
	// Update for animated sprite movement
	for spriteType := range g.sprites.sprites {
		g.sprites.sprites[spriteType].Range(func(k, _ interface{}) bool {

			switch spriteType {
			case MapSpriteType:
				s := k.(*render.Sprite)
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement sprite destruction animation
					g.sprites.deleteMapSprite(s)
				}

				g.updateSpritePosition(s)
				s.Update(g.player.Pos())

			case MechSpriteType:
				s := k.(*render.MechSprite)
				// TODO: implement mech armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteMechSprite(s)
				}

				g.updateMechPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				// TODO: implement vehicle armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteVehicleSprite(s)
				}

				g.updateVehiclePosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				// TODO: implement vtol armor and structure instead of direct HP
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteVTOLSprite(s)
				}

				g.updateVTOLPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				if s.ArmorPoints() <= 0 && s.StructurePoints() <= 0 {
					// TODO: implement unit destruction animation
					g.sprites.deleteInfantrySprite(s)
				}

				g.updateInfantryPosition(s)
				s.Update(g.player.Pos())
				g.updateWeaponCooldowns(model.EntityUnit(s.Entity))
			}

			return true
		})
	}
}

func (g *Game) updateMechPosition(s *render.MechSprite) {
	// TODO: give mechs a bit more of a brain than this
	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateMechPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateVehiclePosition(s *render.VehicleSprite) {
	// TODO: give units a bit more of a brain than this
	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateVehiclePosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateVTOLPosition(s *render.VTOLSprite) {
	// TODO: give units a bit more of a brain than this
	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateVTOLPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateInfantryPosition(s *render.InfantrySprite) {
	// TODO: give mechs a bit more of a brain than this
	sPosition := s.Pos()
	if len(s.PatrolPath) > 0 {
		// make sure there's movement towards the next patrol point
		patrolX, patrolY := s.PatrolPath[s.PatrolPathIndex][0], s.PatrolPath[s.PatrolPathIndex][1]
		patrolLine := geom.Line{X1: sPosition.X, Y1: sPosition.Y, X2: patrolX, Y2: patrolY}

		// TODO: do something about this velocity
		s.SetVelocity(0.02 * s.Scale())

		angle := patrolLine.Angle()
		dist := patrolLine.Distance()

		if dist <= s.Velocity() {
			// start movement towards next patrol point
			s.PatrolPathIndex += 1
			if s.PatrolPathIndex >= len(s.PatrolPath) {
				// loop back towards first patrol point
				s.PatrolPathIndex = 0
			}
			g.updateInfantryPosition(s)
		} else {
			// keep movements towards current patrol point
			s.SetHeading(angle)
		}
	}

	if s.Velocity() != 0 {
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateSpritePosition(s *render.Sprite) {
	if s.Velocity() != 0 {
		sPosition := s.Pos()
		vLine := geom.LineFromAngle(sPosition.X, sPosition.Y, s.Heading(), s.Velocity())

		xCheck := vLine.X2
		yCheck := vLine.Y2

		newPos, newPosZ, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, s.PosZ(), false)
		if isCollision {
			// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
			s.SetHeading(randFloat(-geom.Pi, geom.Pi))
			s.SetVelocity(randFloat(0.005, 0.009))
		} else {
			s.SetPos(newPos)
			s.SetPosZ(newPosZ)
		}
	}
}

func (g *Game) updateWeaponCooldowns(unit model.Unit) {
	if unit == nil {
		return
	}
	armament := unit.Armament()
	if len(armament) == 0 {
		return
	}

	for _, weapon := range armament {
		weapon.DecreaseCooldown(model.SECONDS_PER_TICK)
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	os.Exit(rc)
}
