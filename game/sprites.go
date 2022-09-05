package game

import (
	"fmt"
	"math"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
)

type SpriteHandler struct {
	sprites map[SpriteType]map[raycaster.Sprite]struct{}
}

type SpriteType int

const (
	MapSpriteType SpriteType = iota
	MechSpriteType
	ProjectileSpriteType
	EffectSpriteType
	TotalSpriteTypes
)

func NewSpriteHandler() *SpriteHandler {
	s := &SpriteHandler{sprites: make(map[SpriteType]map[raycaster.Sprite]struct{}, TotalSpriteTypes)}
	s.sprites[MechSpriteType] = make(map[raycaster.Sprite]struct{}, 128)
	s.sprites[MapSpriteType] = make(map[raycaster.Sprite]struct{}, 512)
	s.sprites[ProjectileSpriteType] = make(map[raycaster.Sprite]struct{}, 1024)
	s.sprites[EffectSpriteType] = make(map[raycaster.Sprite]struct{}, 1024)

	return s
}

func (s *SpriteHandler) totalSprites() int {
	total := 0
	for _, spriteMap := range s.sprites {
		total += len(spriteMap)
	}

	return total
}

func (s *SpriteHandler) addMapSprite(sprite *model.Sprite) {
	s.sprites[MapSpriteType][sprite] = struct{}{}
}

func (s *SpriteHandler) deleteMapSprite(sprite *model.Sprite) {
	delete(s.sprites[MapSpriteType], sprite)
}

func (s *SpriteHandler) addMechSprite(mech *model.MechSprite) {
	s.sprites[MechSpriteType][mech] = struct{}{}
}

func (s *SpriteHandler) deleteMechSprite(mech *model.MechSprite) {
	delete(s.sprites[MechSpriteType], mech)
}

func (s *SpriteHandler) addProjectile(projectile *model.Projectile) {
	s.sprites[ProjectileSpriteType][projectile] = struct{}{}
}

func (s *SpriteHandler) deleteProjectile(projectile *model.Projectile) {
	delete(s.sprites[ProjectileSpriteType], projectile)
}

func (s *SpriteHandler) addEffect(effect *model.Effect) {
	s.sprites[EffectSpriteType][effect] = struct{}{}
}

func (s *SpriteHandler) deleteEffect(effect *model.Effect) {
	delete(s.sprites[EffectSpriteType], effect)
}

func (g *Game) getRaycastSprites() []raycaster.Sprite {
	numSprites := g.sprites.totalSprites() + len(g.clutter.sprites)
	raycastSprites := make([]raycaster.Sprite, numSprites)

	index := 0

	for _, spriteMap := range g.sprites.sprites {
		for spriteInterface := range spriteMap {
			sprite := getSpriteFromInterface(spriteInterface)
			// for now this is sufficient, but for much larger amounts of sprites may need goroutines to divide up the work
			// only include map sprites within fast approximation of render distance
			doSprite := g.renderDistance < 0 ||
				(math.Abs(sprite.Position.X-g.player.Position.X) <= g.renderDistance &&
					math.Abs(sprite.Position.Y-g.player.Position.Y) <= g.renderDistance)
			if doSprite {
				raycastSprites[index] = spriteInterface
				index += 1
			}
		}
	}
	for clutter := range g.clutter.sprites {
		raycastSprites[index] = clutter
		index += 1
	}

	return raycastSprites[:index]
}

func getSpriteFromInterface(sInterface raycaster.Sprite) *model.Sprite {
	switch interfaceType := sInterface.(type) {
	case *model.Sprite:
		return sInterface.(*model.Sprite)
	case *model.MechSprite:
		return sInterface.(*model.MechSprite).Sprite
	case *model.Projectile:
		return sInterface.(*model.Projectile).Sprite
	case *model.Effect:
		return sInterface.(*model.Effect).Sprite
	default:
		panic(fmt.Errorf("unable to get sprite entity for type %v", interfaceType))
	}
}

func getEntityFromInterface(sInterface raycaster.Sprite) *model.Entity {
	switch interfaceType := sInterface.(type) {
	case *model.Sprite:
		return sInterface.(*model.Sprite).Entity
	case *model.MechSprite:
		return sInterface.(*model.MechSprite).Entity
	case *model.Projectile:
		return sInterface.(*model.Projectile).Entity
	case *model.Effect:
		return sInterface.(*model.Effect).Entity
	default:
		panic(fmt.Errorf("unable to get sprite entity for type %v", interfaceType))
	}
}
