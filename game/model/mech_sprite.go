package model

import (
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type MechSprite struct {
	*Sprite
	static *ebiten.Image
	ct     *ebiten.Image
	la     *ebiten.Image
	ra     *ebiten.Image
	ll     *ebiten.Image
	rl     *ebiten.Image
}

type MechPart int

const (
	PART_STATIC MechPart = 0
	PART_CT     MechPart = 1
	PART_LA     MechPart = 2
	PART_RA     MechPart = 3
	PART_LL     MechPart = 4
	PART_RL     MechPart = 5
	NUM_PARTS   MechPart = 6
)

// func (s *MechSprite) Scale() float64 {
// 	return s.scale
// }

// func (s *MechSprite) VerticalAnchor() raycaster.SpriteAnchor {
// 	return s.anchor
// }

// func (s *MechSprite) Texture() *ebiten.Image {
// 	return s.texture
// }

// func (s *MechSprite) TextureRect() image.Rectangle {
// 	return s.texRect
// }

func NewMechSprite(
	x, y float64, img *ebiten.Image, collisionRadius float64,
) *MechSprite {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	mechAnimate := NewMechAnimationSheetFromImage(img)

	//p := NewSpriteFromSheet(x, y, 1.0, mechSheet, color.RGBA{}, numCols, numRows, 0, raycaster.AnchorBottom, collisionRadius)
	//p := NewSprite(x, y, 1.0, mechAnimate.sheet, color.RGBA{}, raycaster.AnchorBottom, collisionRadius)
	p := NewAnimatedSprite(x, y, 0.75, 5, mechAnimate.sheet, color.RGBA{}, mechAnimate.maxCols, mechAnimate.maxRows, raycaster.AnchorBottom, collisionRadius)

	// TODO: use function to split out the parts without using NewSpriteFromSheet, since NewMechAnimationSheetFromImage will replace the need
	s := &MechSprite{
		Sprite: p,
	}

	return s
}

func NewMechSpriteFromMech(x, y float64, origMech *MechSprite) *MechSprite {
	s := &MechSprite{}
	copier.Copy(s, origMech)

	s.Sprite = &Sprite{}
	copier.Copy(s.Sprite, origMech.Sprite)

	s.Position = &geom.Vector2{X: x, Y: y}

	return s
}
