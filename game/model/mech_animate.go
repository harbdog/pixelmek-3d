package model

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type MechAnimationIndex int

const (
	ANIMATE_IDLE  MechAnimationIndex = 0
	ANIMATE_STRUT MechAnimationIndex = 1
	// TODO: ANIMATE_SHUTDOWN, ANIMATE_JUMP?
	NUM_ANIMATIONS MechAnimationIndex = 1
)

type MechSpriteAnimate struct {
	sheet            *ebiten.Image
	maxCols, maxRows int
	numColsAtRow     [NUM_ANIMATIONS]int
}

// NewMechAnimationSheetFromImage creates a new image sheet with generated image frames for mech sprite animation
func NewMechAnimationSheetFromImage(srcImage *ebiten.Image) *MechSpriteAnimate {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	srcWidth, srcHeight := srcImage.Size()
	uWidth, uHeight := srcWidth/int(NUM_PARTS), srcHeight

	uSize := uWidth
	if uHeight > uWidth {
		// adjust size to square it off as needed by the raycasting of sprites
		uSize = uHeight
	}

	// determine offsets for center/bottom within each frame
	centerX, bottomY := float64(uSize/2-uWidth/2), float64(uSize-uHeight)

	// maxCols will be determined later based on how many frames needed by any single animation row
	maxRows, maxCols := int(NUM_ANIMATIONS), 1

	// separate out each limb part from source image
	srcParts := make([]*ebiten.Image, int(NUM_PARTS))
	for c := 0; c < int(NUM_PARTS); c++ {
		x, y := c*uWidth, 0
		cellRect := image.Rect(x, y, x+uWidth-1, y+uHeight-1)
		cellImg := srcImage.SubImage(cellRect).(*ebiten.Image)
		srcParts[c] = cellImg
	}

	// static := srcParts[PART_STATIC]
	ct := srcParts[PART_CT]
	la := srcParts[PART_LA]
	ra := srcParts[PART_RA]
	ll := srcParts[PART_LL]
	rl := srcParts[PART_RL]

	// calculate number of animations (rows) and frames for each animation (cols)

	// idle animation: only arms and torso limbs move, for now just going with 4% pixel movement for both
	idlePxPerLimb := 0.02 * float64(uHeight)
	idleCols := 8 // x4 = up -> down -> down -> up (both arms only)
	if idleCols > maxCols {
		maxCols = idleCols
	}

	// strut animation: for now just going with 7% pixel movement for legs, 5% for arms
	strutPxPerLeg, strutPxPerArm := 0.07*float64(uHeight), 0.05*float64(uHeight)
	strutCols := int(strutPxPerLeg) * 4 // x4 = up -> down -> down -> up (starting with left arm, reverse right arm)
	// if strutCols > maxCols {
	// 	maxCols = strutCols
	// }

	mechSheet := ebiten.NewImage(maxCols*uSize, maxRows*uSize)

	m := &MechSpriteAnimate{
		sheet:        mechSheet,
		maxCols:      maxCols,
		maxRows:      maxRows,
		numColsAtRow: [NUM_ANIMATIONS]int{1},
	}

	// first row shall be idle animation

	// TODO: turn into a proper function to deal with each set of movements per frame

	// first frame of idle animation is static image
	row, col := int(ANIMATE_IDLE), 0
	offX, offY := float64(col*uSize)+centerX, float64(row*uSize)+bottomY
	m.drawMechAnimFrame(offX, offY, ct, 0, la, 0, ra, 0, ll, 0, rl, 0)

	// 2x arms up
	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb/2, ra, -idlePxPerLimb/2, ll, 0, rl, 0)

	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb, ra, -idlePxPerLimb, ll, 0, rl, 0)

	// 4x arms down
	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb/2, ra, -idlePxPerLimb/2, ll, 0, rl, 0)

	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, 0, la, 0, ra, 0, ll, 0, rl, 0)

	// 2x ct down also
	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, idlePxPerLimb/2, la, idlePxPerLimb/2, ra, idlePxPerLimb/2, ll, 0, rl, 0)

	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, idlePxPerLimb, la, idlePxPerLimb, ra, idlePxPerLimb, ll, 0, rl, 0)

	// arms and ct back up again
	col++
	offX = float64(col*uSize) + centerX
	m.drawMechAnimFrame(offX, offY, ct, idlePxPerLimb/2, la, idlePxPerLimb/2, ra, idlePxPerLimb/2, ll, 0, rl, 0)

	// TODO: second row shall be strut animation
	if strutPxPerArm >= 0 && strutCols >= 0 {
		// TODO
	}

	return m
}

func (m *MechSpriteAnimate) drawMechAnimFrame(offX, offY float64, ct *ebiten.Image, offCT float64, la *ebiten.Image, offLA float64, ra *ebiten.Image, offRA float64, ll *ebiten.Image, offLL float64, rl *ebiten.Image, offRL float64) {
	offset := ebiten.GeoM{}
	offset.Translate(offX, offY)

	op_ct := &ebiten.DrawImageOptions{GeoM: offset}
	op_ct.GeoM.Translate(0, offCT)
	m.sheet.DrawImage(ct, op_ct)

	op_la := &ebiten.DrawImageOptions{GeoM: offset}
	op_la.GeoM.Translate(0, offLA)
	m.sheet.DrawImage(la, op_la)

	op_ra := &ebiten.DrawImageOptions{GeoM: offset}
	op_ra.GeoM.Translate(0, offRA)
	m.sheet.DrawImage(ra, op_ra)

	op_ll := &ebiten.DrawImageOptions{GeoM: offset}
	op_ll.GeoM.Translate(0, offLL)
	m.sheet.DrawImage(ll, op_ll)

	op_rl := &ebiten.DrawImageOptions{GeoM: offset}
	op_rl.GeoM.Translate(0, offRL)
	m.sheet.DrawImage(rl, op_rl)
}
