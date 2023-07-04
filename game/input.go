package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/harbdog/pixelmek-3d/game/model"
	input "github.com/quasilyte/ebitengine-input"
)

type MouseMode int

const (
	MouseModeTurret MouseMode = iota
	MouseModeBody
	MouseModeCursor
)

const (
	ActionUnknown input.Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionMenu
	ActionBack
	ActionThrottleReverse
	ActionThrottle0
	ActionJumpJet
	ActionWeaponGroupFireToggle
	ActionWeaponGroupSetModifier
	ActionWeaponGroup1
	ActionWeaponGroup2
	ActionWeaponGroup3
	ActionNavCycle
	ActionTargetCrosshairs
	ActionTargetNearest
	ActionTargetNext
	ActionTargetPrevious
	ActionZoomToggle
	actionCount
)

var (
	stringToAction map[string]input.Action
)

func stringAction(aName string) input.Action {
	a, ok := stringToAction[aName]
	if !ok {
		return ActionUnknown
	}
	return a
}

func actionString(a input.Action) string {
	switch a {
	case ActionUp:
		return "up"
	case ActionDown:
		return "down"
	case ActionLeft:
		return "left"
	case ActionRight:
		return "right"
	case ActionMenu:
		return "menu"
	case ActionBack:
		return "back"
	case ActionThrottleReverse:
		return "throttle_reverse"
	case ActionThrottle0:
		return "throttle_0"
	case ActionJumpJet:
		return "jump_jet"
	case ActionWeaponGroupFireToggle:
		return "weapon_group_toggle"
	case ActionWeaponGroupSetModifier:
		return "weapon_group_set"
	case ActionWeaponGroup1:
		return "weapon_group_1"
	case ActionWeaponGroup2:
		return "weapon_group_2"
	case ActionWeaponGroup3:
		return "weapon_group_3"
	case ActionNavCycle:
		return "nav_cycle"
	case ActionTargetCrosshairs:
		return "target_crosshairs"
	case ActionTargetNearest:
		return "target_nearest"
	case ActionTargetNext:
		return "target_next"
	case ActionTargetPrevious:
		return "target_prev"
	case ActionZoomToggle:
		return "zoom_toggle"
	default:
		panic(fmt.Errorf("currently unable to handle actionString for input.Action: %v", a))
	}
}

func (g *Game) initControls() {
	// TODO: read keymap config from viper
	// https://github.com/quasilyte/ebitengine-input/blob/master/_examples/configfile/main.go

	// Build a reverse index to get an action by its name
	stringToAction = map[string]input.Action{}
	for a := ActionUnknown + 1; a < actionCount; a++ {
		stringToAction[actionString(a)] = a
	}

	keymap := input.Keymap{
		ActionUp:    {input.KeyW, input.KeyUp, input.KeyGamepadUp, input.KeyGamepadLStickUp},
		ActionDown:  {input.KeyS, input.KeyDown, input.KeyGamepadDown, input.KeyGamepadLStickDown},
		ActionLeft:  {input.KeyA, input.KeyLeft, input.KeyGamepadLeft, input.KeyGamepadLStickLeft},
		ActionRight: {input.KeyD, input.KeyRight, input.KeyGamepadRight, input.KeyGamepadLStickRight},

		ActionMenu: {input.KeyEscape, input.KeyF1, input.KeyGamepadStart},
		ActionBack: {input.KeyEscape, input.KeyGamepadBack},

		ActionThrottleReverse: {input.KeyBackspace},
		ActionThrottle0:       {input.KeyX, input.KeyGamepadLStick},
		ActionJumpJet:         {input.KeySpace},

		ActionWeaponGroupFireToggle:  {input.KeyBackspace},
		ActionWeaponGroupSetModifier: {input.KeyShift},
		ActionWeaponGroup1:           {input.Key1},
		ActionWeaponGroup2:           {input.Key2},
		ActionWeaponGroup3:           {input.Key3},

		ActionNavCycle:         {input.KeyN},
		ActionTargetCrosshairs: {input.KeyQ},
		ActionTargetNearest:    {input.KeyE},
		ActionTargetNext:       {input.KeyT},
		ActionTargetPrevious:   {input.KeyR},
		ActionZoomToggle:       {input.KeyZ},
	}

	g.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	g.input = g.inputSystem.NewHandler(0, keymap)
}

func (g *Game) handleInput() {
	menuKeyPressed := g.input.ActionIsJustPressed(ActionMenu)
	if menuKeyPressed {
		if g.menu.Active() {
			if g.osType == osTypeBrowser && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				// do not allow Esc key close menu in browser, since Esc key releases browser mouse capture
			} else {
				g.closeMenu()
			}
		} else {
			g.openMenu()
		}
	}

	if g.paused {
		return
	}

	_, isInfantry := g.player.Unit.(*model.Infantry)
	_, isMech := g.player.Unit.(*model.Mech)
	_, isVTOL := g.player.Unit.(*model.VTOL)

	var stop, forward, backward bool
	var rotLeft, rotRight bool

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		// TESTING purposes only
		g.fireTestWeaponAtPlayer()
	}

	if ebiten.IsKeyPressed(ebiten.KeyAlt) {
		if g.mouseMode != MouseModeBody {
			g.mouseMode = MouseModeBody
		}
	} else if inpututil.IsKeyJustReleased(ebiten.KeyAlt) {
		if g.mouseMode == MouseModeBody {
			g.mouseMode = MouseModeTurret

			// reset relative heading target when no longer using mouse turn
			g.player.SetTargetRelativeHeading(0)
		}
	}

	if (g.mouseMode == MouseModeTurret || g.mouseMode == MouseModeBody) && ebiten.CursorMode() != ebiten.CursorModeCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)

		// reset initial mouse capture position
		g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	}

	switch g.mouseMode {
	case MouseModeBody:
		x, y := ebiten.CursorPosition()

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				turnAmount := 0.01 * float64(dx) / g.zoomFovDepth
				g.player.SetTargetRelativeHeading(turnAmount)
			} else {
				// reset relative heading target when mouse stops
				g.player.SetTargetRelativeHeading(0)
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy) / g.zoomFovDepth)
			}
		}
	case MouseModeTurret:
		x, y := ebiten.CursorPosition()

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				if g.player.HasTurret() {
					g.RotateTurret(0.005 * float64(dx) / g.zoomFovDepth)
				} else {
					turnAmount := 0.01 * float64(dx) / g.zoomFovDepth
					g.player.SetTargetRelativeHeading(turnAmount)
				}
			} else {
				if !g.player.HasTurret() {
					// reset relative heading target when mouse stops
					g.player.SetTargetRelativeHeading(0)
				}
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy))
			}
		}
	}

	if g.mouseMode == MouseModeTurret || g.mouseMode == MouseModeBody {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if g.player.fireMode == model.CHAIN_FIRE {
				// cycle to next weapon only in same group (g.player.selectedGroup)
				prevWeapon := g.player.Armament()[g.player.selectedWeapon]
				groupWeapons := g.player.weaponGroups[g.player.selectedGroup]

				var nextWeapon model.Weapon
				nextIndex := 0
				for i, w := range groupWeapons {
					if w == prevWeapon {
						nextIndex = i + 1
						break
					}
				}
				if nextIndex >= len(groupWeapons) {
					nextIndex = 0
				}
				nextWeapon = groupWeapons[nextIndex]

				for i, w := range g.player.Armament() {
					if w == nextWeapon {
						g.player.selectedWeapon = uint(i)
						break
					}
				}
			}
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			if g.player.fireMode == model.GROUP_FIRE {
				g.player.selectedGroup++
				if int(g.player.selectedGroup) >= len(g.player.weaponGroups) {
					g.player.selectedGroup = 0
				}

				// set next selectedGroup only if >0 weapons in it
				weaponsInGroup := len(g.player.weaponGroups[g.player.selectedGroup])
				for weaponsInGroup == 0 {
					g.player.selectedGroup++
					if int(g.player.selectedGroup) >= len(g.player.weaponGroups) {
						g.player.selectedGroup = 0
					}
					weaponsInGroup = len(g.player.weaponGroups[g.player.selectedGroup])
				}

			} else if g.player.fireMode == model.CHAIN_FIRE {
				g.player.selectedWeapon++
				if int(g.player.selectedWeapon) >= len(g.player.Armament()) {
					g.player.selectedWeapon = 0
				}

				// set selectedGroup if the newly selected weapon is in different group
				newSelectedWeapon := g.player.Armament()[g.player.selectedWeapon]
				groups := model.GetGroupsForWeapon(newSelectedWeapon, g.player.weaponGroups)
				if len(groups) == 0 {
					g.player.selectedGroup = 0
				} else {
					g.player.selectedGroup = groups[0]
				}
			}
		}
	}

	if g.player.fireMode == model.CHAIN_FIRE && g.input.ActionIsPressed(ActionWeaponGroupSetModifier) {
		// set group for selected weapon
		setGroupIndex := -1
		switch {
		case g.input.ActionIsJustPressed(ActionWeaponGroup1):
			setGroupIndex = 0
		case g.input.ActionIsJustPressed(ActionWeaponGroup1):
			setGroupIndex = 1
		case g.input.ActionIsJustPressed(ActionWeaponGroup1):
			setGroupIndex = 2
		}

		if setGroupIndex >= 0 {
			weapon := g.player.Armament()[g.player.selectedWeapon]
			groups := model.GetGroupsForWeapon(weapon, g.player.weaponGroups)
			for _, gIndex := range groups {
				if int(gIndex) == setGroupIndex {
					// already in group
					return
				} else {
					// remove from current group
					weaponsInGroup := g.player.weaponGroups[gIndex]
					g.player.weaponGroups[gIndex] = make([]model.Weapon, 0, len(weaponsInGroup)-1)
					for _, chkWeapon := range weaponsInGroup {
						if chkWeapon != weapon {
							g.player.weaponGroups[gIndex] = append(g.player.weaponGroups[gIndex], chkWeapon)
						}
					}
				}
			}

			// add to selected group
			g.player.weaponGroups[setGroupIndex] = append(g.player.weaponGroups[setGroupIndex], weapon)
			g.player.selectedGroup = uint(setGroupIndex)
		}
	}

	if g.input.ActionIsJustPressed(ActionWeaponGroupFireToggle) {
		// toggle group fire mode
		switch g.player.fireMode {
		case model.CHAIN_FIRE:
			g.player.fireMode = model.GROUP_FIRE
		case model.GROUP_FIRE:
			g.player.fireMode = model.CHAIN_FIRE
		}

		if g.player.fireMode == model.GROUP_FIRE {
			// select the first appropriate group from selected weapon when switching to group mode
			prevSelectedWeapon := g.player.Armament()[g.player.selectedWeapon]
			groups := model.GetGroupsForWeapon(prevSelectedWeapon, g.player.weaponGroups)
			if len(groups) == 0 {
				g.player.selectedGroup = 0
			} else {
				g.player.selectedGroup = groups[0]
			}
		} else if g.player.fireMode == model.CHAIN_FIRE {
			// select the first weapon of the group that was selected when switching to chain mode
			prevSelectedGroup := g.player.selectedGroup
			weapons := g.player.weaponGroups[prevSelectedGroup]
			if len(weapons) == 0 {
				g.player.selectedWeapon = 0
			} else {
				for i, w := range g.player.Armament() {
					if w == weapons[0] {
						g.player.selectedWeapon = uint(i)
						break
					}
				}
			}
		}
	}

	if g.input.ActionIsJustPressed(ActionNavCycle) {
		// cycle nav points
		g.navPointCycle()
	}

	if g.input.ActionIsJustPressed(ActionTargetCrosshairs) {
		// target on crosshairs
		g.targetCrosshairs()
	}

	if g.input.ActionIsJustPressed(ActionTargetNearest) {
		// target nearest to player
		g.targetCycle(TARGET_NEAREST)
	}

	if g.input.ActionIsJustPressed(ActionTargetNext) {
		// cycle player targets
		g.targetCycle(TARGET_NEXT)
	}

	if g.input.ActionIsJustPressed(ActionTargetPrevious) {
		// cycle player targets in reverse order
		g.targetCycle(TARGET_PREVIOUS)
	}

	if g.input.ActionIsJustPressed(ActionZoomToggle) {
		// toggle zoom
		if g.camera.FovDepth() != g.zoomFovDepth {
			// zoom in
			zoomFovDegrees := g.fovDegrees / g.zoomFovDepth
			g.camera.SetFovAngle(zoomFovDegrees, g.zoomFovDepth)
			g.camera.SetPitchAngle(g.player.Pitch())
		} else {
			// zoom out
			g.camera.SetFovAngle(g.fovDegrees, 1.0)
			g.camera.SetPitchAngle(g.player.Pitch())
		}
	}

	if g.input.ActionIsJustPressed(ActionThrottleReverse) {
		// toggle reverse throttle
		if g.player.TargetVelocity() > 0 {
			// switch to reverse
			vPercent := g.player.TargetVelocity() / g.player.MaxVelocity()
			g.player.SetTargetVelocity(-vPercent * g.player.MaxVelocity() / 2)
		} else if g.player.TargetVelocity() < 0 {
			// switch to forward
			vPercent := math.Abs(g.player.TargetVelocity()) / (g.player.MaxVelocity() / 2)
			g.player.SetTargetVelocity(vPercent * g.player.MaxVelocity())
		}
	}

	if g.input.ActionIsPressed(ActionJumpJet) {
		switch {
		case isVTOL:
			// TODO: use unit tonnage and gravity to determine ascent speed
			g.player.SetTargetVelocityZ(g.player.MaxVelocity() / 2)
		case isMech:
			if g.player.Unit.JumpJets() > 0 {
				if g.player.Unit.JumpJetsActive() {
					// continue jumping until jets run out of charge
					if g.player.JumpJetDuration() >= g.player.Unit.MaxJumpJetDuration() {
						g.player.Unit.SetJumpJetsActive(false)
						g.player.SetTargetVelocityZ(0)
					}
				} else if g.player.JumpJetDuration() < 0.9*g.player.Unit.MaxJumpJetDuration() {
					// only allow jump jets to reengage if not close to the max jet charge usage
					g.player.Unit.SetJumpJetsActive(true)
					g.player.SetTargetVelocityZ(0.05)
				}
			}
		}
		// TODO: infantry jump (or jump jet infantry)

	} else if ebiten.IsKeyPressed(ebiten.KeyControl) {
		switch {
		case isVTOL:
			// TODO: use unit tonnage and gravity to determine ascent speed
			g.player.SetTargetVelocityZ(-g.player.MaxVelocity() / 2)
		}

	} else if g.player.TargetVelocityZ() != 0 {
		g.player.SetTargetVelocityZ(0)
		switch {
		case isMech:
			g.player.Unit.SetJumpJetsActive(false)
		}
	}

	if g.input.ActionIsPressed(ActionLeft) {
		rotLeft = true
	}
	if g.input.ActionIsPressed(ActionRight) {
		rotRight = true
	}

	if g.input.ActionIsPressed(ActionUp) {
		forward = true
	}
	if g.input.ActionIsPressed(ActionDown) {
		backward = true
	}

	if g.input.ActionIsPressed(ActionThrottle0) {
		stop = true
	}

	switch g.throttleDecay {
	case true:
		if forward {
			g.player.SetTargetVelocity(g.player.MaxVelocity())
		} else if backward {
			g.player.SetTargetVelocity(-g.player.MaxVelocity() / 2)
		} else {
			g.player.SetTargetVelocity(0)
		}
	case false:
		deltaV := 0.0004 // FIXME: testing
		if forward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() + deltaV)
		} else if backward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() - deltaV)
		} else if stop {
			g.player.SetTargetVelocity(0)
		}
	}

	isStrafe := false
	if !g.player.HasTurret() && (rotLeft || rotRight) {
		// only infantry/battle armor and VTOL can strafe
		if isInfantry || isVTOL {
			// strafe instead of rotate
			isStrafe = true
		}
	}

	if isStrafe {
		// TODO: use unit max velocity to determine strafe speed
		if rotLeft {
			g.Strafe(-0.05)
		} else if rotRight {
			g.Strafe(0.05)
		}
	} else {
		if rotLeft {
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(turnAmount)
		} else if rotRight {
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(-turnAmount)
		}
	}
}
