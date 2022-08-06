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

type mechAnimatePart struct {
	image   *ebiten.Image
	travelY float64
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
	centerX, bottomY := float64(uSize)/2-float64(uWidth)/2, float64(uSize-uHeight)

	// maxCols will be determined later based on how many frames needed by any single animation row
	maxRows, maxCols := int(NUM_ANIMATIONS), 1

	// separate out each limb part from source image
	srcParts := make([]*mechAnimatePart, int(NUM_PARTS))
	for c := 0; c < int(NUM_PARTS); c++ {
		x, y := c*uWidth, 0
		cellRect := image.Rect(x, y, x+uWidth-1, y+uHeight-1)
		cellImg := srcImage.SubImage(cellRect).(*ebiten.Image)
		srcParts[c] = &mechAnimatePart{image: cellImg, travelY: 0}
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

	// first frame of idle animation is static image
	row, col := int(ANIMATE_IDLE), 0
	//m.drawMechAnimFrame(offX, offY, ct.image, 0, la.image, 0, ra.image, 0, ll.image, 0, rl.image, 0)
	m.drawMechAnimationParts(row, col, 1, uSize, centerX, bottomY, ct, 0, la, 0, ra, 0, ll, 0, rl, 0)
	col++

	// 2x arms up
	//m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb/2, ra, -idlePxPerLimb/2, ll, 0, rl, 0)
	//m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb, ra, -idlePxPerLimb, ll, 0, rl, 0)
	m.drawMechAnimationParts(row, col, 2, uSize, centerX, bottomY, ct, 0, la, -idlePxPerLimb, ra, -idlePxPerLimb, ll, 0, rl, 0)
	col += 2

	// 2x arms down
	//m.drawMechAnimFrame(offX, offY, ct, 0, la, -idlePxPerLimb/2, ra, -idlePxPerLimb/2, ll, 0, rl, 0)
	//m.drawMechAnimFrame(offX, offY, ct, 0, la, 0, ra, 0, ll, 0, rl, 0)
	m.drawMechAnimationParts(row, col, 2, uSize, centerX, bottomY, ct, 0, la, idlePxPerLimb, ra, idlePxPerLimb, ll, 0, rl, 0)
	col += 2

	// 2x arms down + 2x ct down
	//m.drawMechAnimFrame(offX, offY, ct.image, idlePxPerLimb/2, la.image, idlePxPerLimb/2, ra.image, idlePxPerLimb/2, ll.image, 0, rl.image, 0)
	//m.drawMechAnimFrame(offX, offY, ct.image, idlePxPerLimb, la.image, idlePxPerLimb, ra.image, idlePxPerLimb, ll.image, 0, rl.image, 0)
	m.drawMechAnimationParts(row, col, 2, uSize, centerX, bottomY, ct, idlePxPerLimb, la, idlePxPerLimb, ra, idlePxPerLimb, ll, 0, rl, 0)
	col += 2

	// 1x arms and ct back up again
	//m.drawMechAnimFrame(offX, offY, ct.image, idlePxPerLimb/2, la.image, idlePxPerLimb/2, ra.image, idlePxPerLimb/2, ll.image, 0, rl.image, 0)
	m.drawMechAnimationParts(row, col, 1, uSize, centerX, bottomY, ct, -idlePxPerLimb/2, la, -idlePxPerLimb/2, ra, -idlePxPerLimb/2, ll, 0, rl, 0)

	// TODO: second row shall be strut animation
	if strutPxPerArm >= 0 && strutCols >= 0 {
		// TODO
	}

	return m
}

// drawMechAnimationParts draws onto the sheet each mech part with total pixel travel over a number of given frames
// starting at the given column within the given row in the sheet of frames
func (m *MechSpriteAnimate) drawMechAnimationParts(
	row, col, frames, uSize int, adjustX, adjustY float64, ct *mechAnimatePart, pxCT float64,
	la *mechAnimatePart, pxLA float64, ra *mechAnimatePart, pxRA float64,
	ll *mechAnimatePart, pxLL float64, rl *mechAnimatePart, pxRL float64,
) {
	offsetY := float64(row*uSize) + adjustY

	// use previously tracked offsets in parts as starting point
	pxPerCT := ct.travelY
	pxPerLA := la.travelY
	pxPerRA := ra.travelY
	pxPerLL := ll.travelY
	pxPerRL := rl.travelY

	for c := col; c < col+frames; c++ {
		offsetX := float64(c*uSize) + adjustX
		pxPerCT += pxCT / float64(frames)
		pxPerLA += pxLA / float64(frames)
		pxPerRA += pxRA / float64(frames)
		pxPerLL += pxLL / float64(frames)
		pxPerRL += pxRL / float64(frames)

		m.drawMechAnimFrame(offsetX, offsetY, ct.image, pxPerCT, la.image, pxPerLA, ra.image, pxPerRA, ll.image, pxPerLL, rl.image, pxPerRL)
	}

	// keep track of offsets in parts for next animation
	ct.travelY += pxCT
	la.travelY += pxLA
	ra.travelY += pxRA
	ll.travelY += pxLL
	rl.travelY += pxRL
}

// drawMechAnimFrame draws onto the sheet each mech part each with given offet for the frame (offX, offY),
// and individual offsets specific for each part
func (m *MechSpriteAnimate) drawMechAnimFrame(
	offX, offY float64, ct *ebiten.Image, offCT float64, la *ebiten.Image, offLA float64,
	ra *ebiten.Image, offRA float64, ll *ebiten.Image, offLL float64, rl *ebiten.Image, offRL float64,
) {
	offset := ebiten.GeoM{}
	offset.Translate(offX, offY)

	op_ct := &ebiten.DrawImageOptions{GeoM: offset}
	op_ct.GeoM.Translate(0, offCT)
	m.sheet.DrawImage(ct, op_ct)

	op_ll := &ebiten.DrawImageOptions{GeoM: offset}
	op_ll.GeoM.Translate(0, offLL)
	m.sheet.DrawImage(ll, op_ll)

	op_rl := &ebiten.DrawImageOptions{GeoM: offset}
	op_rl.GeoM.Translate(0, offRL)
	m.sheet.DrawImage(rl, op_rl)

	op_la := &ebiten.DrawImageOptions{GeoM: offset}
	op_la.GeoM.Translate(0, offLA)
	m.sheet.DrawImage(la, op_la)

	op_ra := &ebiten.DrawImageOptions{GeoM: offset}
	op_ra.GeoM.Translate(0, offRA)
	m.sheet.DrawImage(ra, op_ra)
}
