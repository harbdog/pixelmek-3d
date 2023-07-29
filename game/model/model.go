package model

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/harbdog/pixelmek-3d/game/resources"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

const (
	METERS_PER_UNIT  float64 = 20
	TICKS_PER_SECOND float64 = 60
	SECONDS_PER_TICK float64 = 1 / TICKS_PER_SECOND

	VELOCITY_TO_KPH float64 = (METERS_PER_UNIT / 1000) * (TICKS_PER_SECOND * 60 * 60)
	KPH_TO_VELOCITY float64 = 1 / VELOCITY_TO_KPH

	GRAVITY_METERS_PSS float64 = 9.80665
	GRAVITY_UNITS_PTT  float64 = GRAVITY_METERS_PSS / METERS_PER_UNIT / (TICKS_PER_SECOND * TICKS_PER_SECOND)

	CEILING_JUMP float64 = 2.0
	CEILING_VTOL float64 = 5.0 // TODO: set flight ceiling in map yaml
)

var (
	Randish *rand.Rand
)

func init() {
	Randish = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type ModelResources struct {
	Mechs            map[string]*ModelMechResource
	Vehicles         map[string]*ModelVehicleResource
	VTOLs            map[string]*ModelVTOLResource
	Infantry         map[string]*ModelInfantryResource
	Emplacements     map[string]*ModelEmplacementResource
	EnergyWeapons    map[string]*ModelEnergyWeaponResource
	MissileWeapons   map[string]*ModelMissileWeaponResource
	BallisticWeapons map[string]*ModelBallisticWeaponResource
}

const (
	MechResourceType        string = "mechs"
	VehicleResourceType     string = "vehicles"
	VTOLResourceType        string = "vtols"
	InfantryResourceType    string = "infantry"
	EmplacementResourceType string = "emplacements"
	ProjectilesResourceType string = "projectiles"
	EffectsResourceType     string = "effects"
	EnergyResourceType      string = "energy"
	MissileResourceType     string = "missile"
	BallisticResourceType   string = "ballistic"
)

type TechBase int

const (
	CLAN TechBase = iota
	IS
)

func TechBaseString(t TechBase) string {
	switch t {
	case CLAN:
		return "clan"
	case IS:
		return "is"
	}
	return "unknown"
}

type ModelTech struct {
	TechBase
}

type HeatSinkType int

const (
	NONE   HeatSinkType = iota // 0
	SINGLE                     // 1
	DOUBLE                     // 2
)

type ModelHeatSinkType struct {
	HeatSinkType
}

type Location int

const (
	HEAD Location = iota
	CENTER_TORSO
	LEFT_TORSO
	RIGHT_TORSO
	LEFT_ARM
	RIGHT_ARM
	LEFT_LEG
	RIGHT_LEG
	FRONT
	RIGHT
	LEFT
	TURRET
)

type ModelLocation struct {
	Location
}

type ModelMechResource struct {
	File              string                   `yaml:"-"`
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=200"`
	Height            float64                  `yaml:"height" validate:"gt=0"`
	HeightPxRatio     float64                  `yaml:"heightRatio" validate:"gte=0"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int                      `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	CockpitPxOffset   [2]float64               `yaml:"cockpitOffset" validate:"required"`
	HeatSinks         *ModelResourceHeatSinks  `yaml:"heatSinks"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelVehicleResource struct {
	File              string                   `yaml:"-"`
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=200"`
	Height            float64                  `yaml:"height" validate:"gt=0"`
	HeightPxRatio     float64                  `yaml:"heightRatio" validate:"gte=0"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	CockpitPxOffset   [2]float64               `yaml:"cockpitOffset" validate:"required"`
	HeatSinks         *ModelResourceHeatSinks  `yaml:"heatSinks"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelVTOLResource struct {
	File              string                   `yaml:"-"`
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Height            float64                  `yaml:"height" validate:"gt=0"`
	HeightPxRatio     float64                  `yaml:"heightRatio" validate:"gte=0"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	CockpitPxOffset   [2]float64               `yaml:"cockpitOffset" validate:"required"`
	HeatSinks         *ModelResourceHeatSinks  `yaml:"heatSinks"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelInfantryResource struct {
	File              string                   `yaml:"-"`
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Height            float64                  `yaml:"height" validate:"gt=0"`
	HeightPxRatio     float64                  `yaml:"heightRatio" validate:"gte=0"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int                      `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	CockpitPxOffset   [2]float64               `yaml:"cockpitOffset" validate:"required"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelEmplacementResource struct {
	File              string                   `yaml:"-"`
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Height            float64                  `yaml:"height" validate:"gt=0"`
	HeightPxRatio     float64                  `yaml:"heightRatio" validate:"gte=0"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	CockpitPxOffset   [2]float64               `yaml:"cockpitOffset" validate:"required"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelEnergyWeaponResource struct {
	Name            string                   `yaml:"name" validate:"required"`
	ShortName       string                   `yaml:"short" validate:"required"`
	Tech            ModelTech                `yaml:"tech" validate:"required"`
	Tonnage         float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Damage          float64                  `yaml:"damage" validate:"gt=0"`
	Heat            float64                  `yaml:"heat" validate:"gte=0"`
	Distance        float64                  `yaml:"distance" validate:"gt=0"`
	ExtremeDistance float64                  `yaml:"extremeDistance" validate:"gte=0"`
	Velocity        float64                  `yaml:"velocity" validate:"gt=0"`
	Cooldown        float64                  `yaml:"cooldown" validate:"gt=0"`
	ProjectileCount int                      `yaml:"projectileCount" validate:"gt=0"`
	ProjectileDelay float64                  `yaml:"projectileDelay" validate:"gte=0"`
	Projectile      *ModelProjectileResource `yaml:"projectile"`
	Audio           string                   `yaml:"audio" validate:"required"`
}

type ModelMissileWeaponResource struct {
	Name            string                    `yaml:"name" validate:"required"`
	ShortName       string                    `yaml:"short" validate:"required"`
	Tech            ModelTech                 `yaml:"tech" validate:"required"`
	Tonnage         float64                   `yaml:"tonnage" validate:"gt=0,lte=100"`
	Damage          float64                   `yaml:"damage" validate:"gt=0"`
	Heat            float64                   `yaml:"heat" validate:"gte=0"`
	Distance        float64                   `yaml:"distance" validate:"gt=0"`
	ExtremeDistance float64                   `yaml:"extremeDistance" validate:"gte=0"`
	Velocity        float64                   `yaml:"velocity" validate:"gt=0"`
	Cooldown        float64                   `yaml:"cooldown" validate:"gt=0"`
	ProjectileCount int                       `yaml:"projectileCount" validate:"gt=0"`
	ProjectileDelay float64                   `yaml:"projectileDelay" validate:"gte=0"`
	Projectile      *ModelProjectileResource  `yaml:"projectile"`
	LockOn          *ModelMissileWeaponLockOn `yaml:"lockOn,omitempty"`
	Audio           string                    `yaml:"audio" validate:"required"`
}

type ModelBallisticWeaponResource struct {
	Name            string                   `yaml:"name" validate:"required"`
	ShortName       string                   `yaml:"short" validate:"required"`
	Tech            ModelTech                `yaml:"tech" validate:"required"`
	Tonnage         float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Damage          float64                  `yaml:"damage" validate:"gt=0"`
	Heat            float64                  `yaml:"heat" validate:"gte=0"`
	Distance        float64                  `yaml:"distance" validate:"gt=0"`
	ExtremeDistance float64                  `yaml:"extremeDistance" validate:"gte=0"`
	Velocity        float64                  `yaml:"velocity" validate:"gt=0"`
	Cooldown        float64                  `yaml:"cooldown" validate:"gt=0"`
	ProjectileCount int                      `yaml:"projectileCount" validate:"gt=0"`
	ProjectileDelay float64                  `yaml:"projectileDelay" validate:"gte=0"`
	Projectile      *ModelProjectileResource `yaml:"projectile"`
	Audio           string                   `yaml:"audio" validate:"required"`
}

type ModelProjectileResource struct {
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
	ImpactEffect      *ModelEffectResource     `yaml:"impactEffect"`
}

type ModelMissileWeaponLockOn struct {
	LockRequired bool    `yaml:"lockRequired"`
	TurnRate     float64 `yaml:"turnRate" validate:"gt=0"`
	GroupRadius  float64 `yaml:"groupRadius" validate:"gt=0"`
}

type ModelEffectResource struct {
	Image      string                   `yaml:"image" validate:"required"`
	ImageSheet *ModelResourceImageSheet `yaml:"imageSheet"`
	Scale      float64                  `yaml:"scale" validate:"gt=0"`
	Audio      string                   `yaml:"audio" validate:"required"`
}

type ModelResourceImageSheet struct {
	Columns        int             `yaml:"columns" validate:"gt=0"`
	Rows           int             `yaml:"rows" validate:"gt=0"`
	AnimationRate  int             `yaml:"animationRate" validate:"gte=0"`
	AngleFacingRow map[float64]int `yaml:"angleFacingRow"`
	StaticIndex    int             `yaml:"staticIndex" validate:"gte=0"`
}

type ModelResourceHeatSinks struct {
	Quantity int               `yaml:"quantity" validate:"gte=0"`
	Type     ModelHeatSinkType `yaml:"type" validate:"required"`
}

type ModelWeaponType struct {
	WeaponType
}

type ModelResourceArmament struct {
	Weapon   string          `yaml:"weapon" validate:"required"`
	Type     ModelWeaponType `yaml:"type" validate:"required"`
	Location ModelLocation   `yaml:"location" validate:"required"`
	Offset   [2]float64      `yaml:"offset" validate:"required"`
}

// Unmarshals into TechBase
func (t *ModelTech) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	clan, is := TechBaseString(CLAN), TechBaseString(IS)

	switch str {
	case clan:
		t.TechBase = CLAN
	case is:
		t.TechBase = IS
	default:
		return fmt.Errorf("unknown tech value '%s', must be one of: [%s, %s]", str, clan, is)
	}

	return nil
}

// Unmarshals into HeatSinkType
func (t *ModelHeatSinkType) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	single, double := "single", "double"

	switch str {
	case single:
		t.HeatSinkType = SINGLE
	case double:
		t.HeatSinkType = DOUBLE
	default:
		return fmt.Errorf("unknown heat sink type value '%s', must be one of: [%s, %s]", str, single, double)
	}

	return nil
}

// Unmarshals into WeaponType
func (t *ModelWeaponType) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	energy, ballistic, missile := "energy", "ballistic", "missile"

	switch str {
	case energy:
		t.WeaponType = ENERGY
	case ballistic:
		t.WeaponType = BALLISTIC
	case missile:
		t.WeaponType = MISSILE
	default:
		return fmt.Errorf(
			"unknown weapon type value '%s', must be one of: [%s, %s, %s]", str, energy, ballistic, missile,
		)
	}

	return nil
}

// Unmarshals into Location
func (t *ModelLocation) UnmarshalText(b []byte) error {
	str := strings.Trim(string(b), `"`)

	hd, ct, lt, rt, la, ra, ll, rl := "hd", "ct", "lt", "rt", "la", "ra", "ll", "rl"
	front, left, right, turret := "front", "left", "right", "turret"

	switch str {
	case hd:
		t.Location = HEAD
	case ct:
		t.Location = CENTER_TORSO
	case lt:
		t.Location = LEFT_TORSO
	case rt:
		t.Location = RIGHT_TORSO
	case la:
		t.Location = LEFT_ARM
	case ra:
		t.Location = RIGHT_ARM
	case ll:
		t.Location = LEFT_LEG
	case rl:
		t.Location = RIGHT_LEG
	case front:
		t.Location = FRONT
	case left:
		t.Location = LEFT
	case right:
		t.Location = RIGHT
	case turret:
		t.Location = TURRET
	default:
		return fmt.Errorf("unknown location value '%s'", str)
	}

	return nil
}

func LoadModelResources() (*ModelResources, error) {
	resources := &ModelResources{}

	err := resources.loadWeaponResources()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = resources.loadUnitResources()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return resources, nil
}

func (r *ModelResources) loadUnitResources() error {
	// load and validate all units
	v := validator.New()

	unitsPath := "units"
	unitsTypes, err := resources.FilesInPath(unitsPath)
	if err != nil {
		return err
	}

	for _, t := range unitsTypes {
		if !t.IsDir() {
			// only folders with unit type name expected
			continue
		}

		unitType := t.Name()
		unitTypePath := filepath.Join(unitsPath, unitType)
		unitFiles, err := resources.FilesInPath(unitTypePath)
		if err != nil {
			return err
		}

		// initialize map for each recognized unit type
		switch unitType {
		case MechResourceType:
			r.Mechs = make(map[string]*ModelMechResource, len(unitFiles))

		case VehicleResourceType:
			r.Vehicles = make(map[string]*ModelVehicleResource, len(unitFiles))

		case VTOLResourceType:
			r.VTOLs = make(map[string]*ModelVTOLResource, len(unitFiles))

		case InfantryResourceType:
			r.Infantry = make(map[string]*ModelInfantryResource, len(unitFiles))

		case EmplacementResourceType:
			r.Emplacements = make(map[string]*ModelEmplacementResource, len(unitFiles))
		}

		for _, u := range unitFiles {
			if u.IsDir() {
				// TODO: support recursive directory structure?
				continue
			}

			fileName := u.Name()
			filePath := filepath.Join(unitTypePath, fileName)
			unitYaml, err := resources.ReadFile(filePath)
			if err != nil {
				return err
			}

			switch unitType {
			case MechResourceType:
				m := &ModelMechResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				m.File = fileName
				r.Mechs[fileName] = m

			case VehicleResourceType:
				m := &ModelVehicleResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				m.File = fileName
				r.Vehicles[fileName] = m

			case VTOLResourceType:
				m := &ModelVTOLResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				m.File = fileName
				r.VTOLs[fileName] = m

			case InfantryResourceType:
				m := &ModelInfantryResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				m.File = fileName
				r.Infantry[fileName] = m

			case EmplacementResourceType:
				m := &ModelEmplacementResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				m.File = fileName
				r.Emplacements[fileName] = m
			}
		}
	}

	return nil
}

func (r *ModelResources) loadWeaponResources() error {
	// load and validate all weapons, projectiles and impact efffects
	v := validator.New()

	weaponsPath := "weapons"
	weaponsTypes, err := resources.FilesInPath(weaponsPath)
	if err != nil {
		return err
	}

	for _, t := range weaponsTypes {
		if !t.IsDir() {
			// only folders with weapon type name expected
			continue
		}

		weaponType := t.Name()
		weaponTypePath := filepath.Join(weaponsPath, weaponType)
		weaponFiles, err := resources.FilesInPath(weaponTypePath)
		if err != nil {
			return err
		}

		// initialize map for each recognized weapon type
		switch weaponType {
		case EnergyResourceType:
			r.EnergyWeapons = make(map[string]*ModelEnergyWeaponResource, len(weaponFiles))
		case MissileResourceType:
			r.MissileWeapons = make(map[string]*ModelMissileWeaponResource, len(weaponFiles))
		case BallisticResourceType:
			r.BallisticWeapons = make(map[string]*ModelBallisticWeaponResource, len(weaponFiles))
		}

		for _, u := range weaponFiles {
			if u.IsDir() {
				continue
			}

			fileName := u.Name()
			weaponFilePath := filepath.Join(weaponTypePath, fileName)
			weaponYaml, err := resources.ReadFile(weaponFilePath)
			if err != nil {
				return err
			}

			switch weaponType {
			case EnergyResourceType:
				m := &ModelEnergyWeaponResource{}
				err = yaml.Unmarshal(weaponYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", weaponFilePath, err.Error())
				}

				r.EnergyWeapons[fileName] = m

			case MissileResourceType:
				m := &ModelMissileWeaponResource{}
				err = yaml.Unmarshal(weaponYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", weaponFilePath, err.Error())
				}

				r.MissileWeapons[fileName] = m

			case BallisticResourceType:
				m := &ModelBallisticWeaponResource{}
				err = yaml.Unmarshal(weaponYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", weaponFilePath, err.Error())
				}

				r.BallisticWeapons[fileName] = m

			}
		}
	}

	return nil
}

func (r *ModelResources) GetMechResource(unit string) *ModelMechResource {
	if m, ok := r.Mechs[unit]; ok {
		return m
	}
	log.Errorf("mech unit resource does not exist %s", unit)
	return nil
}

// GetMechResourceList gets mech resources as sorted list
func (r *ModelResources) GetMechResourceList() []*ModelMechResource {
	resourceList := make([]*ModelMechResource, 0, len(r.Mechs))
	for _, v := range r.Mechs {
		resourceList = append(resourceList, v)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		rI, rJ := resourceList[i], resourceList[j]
		return rI.Name < rJ.Name || (rI.Name == rJ.Name && rI.Variant < rJ.Variant)
	})

	return resourceList
}

func (r *ModelResources) GetVehicleResource(unit string) *ModelVehicleResource {
	if m, ok := r.Vehicles[unit]; ok {
		return m
	}
	log.Errorf("vehicle unit resource does not exist %s", unit)
	return nil
}

// GetVehicleResourceList gets vehicle resources as sorted list
func (r *ModelResources) GetVehicleResourceList() []*ModelVehicleResource {
	resourceList := make([]*ModelVehicleResource, 0, len(r.Mechs))
	for _, v := range r.Vehicles {
		resourceList = append(resourceList, v)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		rI, rJ := resourceList[i], resourceList[j]
		return rI.Name < rJ.Name || (rI.Name == rJ.Name && rI.Variant < rJ.Variant)
	})

	return resourceList
}

func (r *ModelResources) GetVTOLResource(unit string) *ModelVTOLResource {
	if m, ok := r.VTOLs[unit]; ok {
		return m
	}
	log.Errorf("vtol unit resource does not exist %s", unit)
	return nil
}

// GetVTOLResourceList gets vtol resources as sorted list
func (r *ModelResources) GetVTOLResourceList() []*ModelVTOLResource {
	resourceList := make([]*ModelVTOLResource, 0, len(r.VTOLs))
	for _, v := range r.VTOLs {
		resourceList = append(resourceList, v)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		rI, rJ := resourceList[i], resourceList[j]
		return rI.Name < rJ.Name || (rI.Name == rJ.Name && rI.Variant < rJ.Variant)
	})

	return resourceList
}

func (r *ModelResources) GetInfantryResource(unit string) *ModelInfantryResource {
	if m, ok := r.Infantry[unit]; ok {
		return m
	}
	log.Errorf("infantry unit resource does not exist %s", unit)
	return nil
}

// GetInfantryResourceList gets infantry resources as sorted list
func (r *ModelResources) GetInfantryResourceList() []*ModelInfantryResource {
	resourceList := make([]*ModelInfantryResource, 0, len(r.Infantry))
	for _, v := range r.Infantry {
		resourceList = append(resourceList, v)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		rI, rJ := resourceList[i], resourceList[j]
		return rI.Name < rJ.Name || (rI.Name == rJ.Name && rI.Variant < rJ.Variant)
	})

	return resourceList
}

func (r *ModelResources) GetEmplacementResource(unit string) *ModelEmplacementResource {
	if m, ok := r.Emplacements[unit]; ok {
		return m
	}
	log.Errorf("emplacement unit resource does not exist %s", unit)
	return nil
}

// GetEmplacementResourceList gets emplacement resources as sorted list
func (r *ModelResources) GetEmplacementResourceList() []*ModelEmplacementResource {
	resourceList := make([]*ModelEmplacementResource, 0, len(r.Emplacements))
	for _, v := range r.Emplacements {
		resourceList = append(resourceList, v)
	}

	sort.Slice(resourceList, func(i, j int) bool {
		rI, rJ := resourceList[i], resourceList[j]
		return rI.Name < rJ.Name || (rI.Name == rJ.Name && rI.Variant < rJ.Variant)
	})

	return resourceList
}

func (r *ModelResources) GetEnergyWeaponResource(weapon string) *ModelEnergyWeaponResource {
	if m, ok := r.EnergyWeapons[weapon]; ok {
		return m
	}
	log.Errorf("energy weapon resource does not exist %s", weapon)
	return nil
}

func (r *ModelResources) GetMissileWeaponResource(weapon string) *ModelMissileWeaponResource {
	if m, ok := r.MissileWeapons[weapon]; ok {
		return m
	}
	log.Errorf("missile weapon resource does not exist %s", weapon)
	return nil
}

func (r *ModelResources) GetBallisticWeaponResource(weapon string) *ModelBallisticWeaponResource {
	if m, ok := r.BallisticWeapons[weapon]; ok {
		return m
	}
	log.Errorf("ballistic weapon resource does not exist %s", weapon)
	return nil
}
