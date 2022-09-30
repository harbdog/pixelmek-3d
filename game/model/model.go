package model

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type ModelResources struct {
	Mechs         map[string]*ModelMechResource
	Vehicles      map[string]*ModelVehicleResource
	VTOLs         map[string]*ModelVTOLResource
	Infantry      map[string]*ModelInfantryResource
	EnergyWeapons map[string]*ModelEnergyWeaponResource
}

const (
	MechResourceType        string = "mechs"
	VehicleResourceType     string = "vehicles"
	VTOLResourceType        string = "vtols"
	InfantryResourceType    string = "infantry"
	ProjectilesResourceType string = "projectiles"
	EffectsResourceType     string = "effects"
	EnergyResourceType      string = "energy"
)

type TechBase int

const (
	CLAN TechBase = iota
	IS
)

type ModelTech struct {
	TechBase
}

type HeatSinkType int

const (
	SINGLE HeatSinkType = iota
	DOUBLE
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
)

type ModelLocation struct {
	Location
}

type ModelMechResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=200"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int                      `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
	CockpitOffset     [2]float64               `yaml:"cockpitOffset" validate:"required"`
	HeatSinks         *ModelResourceHeatSinks  `yaml:"heatSinks"`
	Armament          []*ModelResourceArmament `yaml:"armament"`
}

type ModelVehicleResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=200"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelVTOLResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Tonnage           float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelInfantryResource struct {
	Name              string                   `yaml:"name" validate:"required"`
	Variant           string                   `yaml:"variant" validate:"required"`
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	Tech              ModelTech                `yaml:"tech" validate:"required"`
	Speed             float64                  `yaml:"speed" validate:"gt=0,lte=250"`
	JumpJets          int                      `yaml:"jumpJets" validate:"gte=0,lte=20"`
	Armor             float64                  `yaml:"armor" validate:"gte=0"`
	Structure         float64                  `yaml:"structure" validate:"gt=0"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelEnergyWeaponResource struct {
	Name       string                   `yaml:"name" validate:"required"`
	ShortName  string                   `yaml:"short" validate:"required"`
	Tech       ModelTech                `yaml:"tech" validate:"required"`
	Tonnage    float64                  `yaml:"tonnage" validate:"gt=0,lte=100"`
	Damage     float64                  `yaml:"damage" validate:"gt=0"`
	Heat       float64                  `yaml:"heat" validate:"gte=0"`
	Distance   float64                  `yaml:"distance" validate:"gt=0"`
	Velocity   float64                  `yaml:"velocity" validate:"gt=0"`
	Cooldown   float64                  `yaml:"cooldown" validate:"gt=0"`
	Projectile *ModelProjectileResource `yaml:"projectile"`
}

type ModelProjectileResource struct {
	Image             string                   `yaml:"image" validate:"required"`
	ImageSheet        *ModelResourceImageSheet `yaml:"imageSheet"`
	CollisionPxRadius float64                  `yaml:"collisionRadius" validate:"gt=0"`
	CollisionPxHeight float64                  `yaml:"collisionHeight" validate:"gt=0"`
	Scale             float64                  `yaml:"scale" validate:"gt=0"`
	ImpactEffect      *ModelEffectResource     `yaml:"impactEffect"`
}

type ModelEffectResource struct {
	Image      string                   `yaml:"image" validate:"required"`
	ImageSheet *ModelResourceImageSheet `yaml:"imageSheet"`
	Scale      float64                  `yaml:"scale" validate:"gt=0"`
}

type ModelResourceImageSheet struct {
	Columns        int             `yaml:"columns" validate:"gt=0"`
	Rows           int             `yaml:"rows" validate:"gt=0"`
	AnimationRate  int             `yaml:"animationRate" validate:"gte=0"`
	AngleFacingRow map[float64]int `yaml:"angleFacingRow"`
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

	clan, is := "clan", "is"

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

	unitsPath := filepath.Join("game", "resources", "units")
	unitsTypes, err := filesInPath(unitsPath)
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
		unitFiles, err := filesInPath(unitTypePath)
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

		}

		for _, u := range unitFiles {
			if u.IsDir() {
				// TODO: support recursive directory structure?
				continue
			}

			fileName := u.Name()
			filePath := filepath.Join(unitTypePath, fileName)
			unitYaml, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			switch unitType {
			case MechResourceType:
				m := &ModelMechResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				r.Mechs[fileName] = m

			case VehicleResourceType:
				m := &ModelVehicleResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				r.Vehicles[fileName] = m

			case VTOLResourceType:
				m := &ModelVTOLResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				r.VTOLs[fileName] = m

			case InfantryResourceType:
				m := &ModelInfantryResource{}
				err = yaml.Unmarshal(unitYaml, m)
				if err != nil {
					return err
				}

				err = v.Struct(m)
				if err != nil {
					return fmt.Errorf("[%s] %s", filePath, err.Error())
				}

				r.Infantry[fileName] = m

			}
		}
	}

	return nil
}

func (r *ModelResources) loadWeaponResources() error {
	// load and validate all weapons, projectiles and impact efffects
	v := validator.New()

	weaponsPath := filepath.Join("game", "resources", "weapons")
	weaponsTypes, err := filesInPath(weaponsPath)
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
		weaponFiles, err := filesInPath(weaponTypePath)
		if err != nil {
			return err
		}

		// initialize map for each recognized weapon type
		switch weaponType {
		case EnergyResourceType:
			r.EnergyWeapons = make(map[string]*ModelEnergyWeaponResource, len(weaponFiles))

		}

		for _, u := range weaponFiles {
			if u.IsDir() {
				continue
			}

			fileName := u.Name()
			weaponYaml, err := ioutil.ReadFile(filepath.Join(weaponTypePath, fileName))
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
					return fmt.Errorf("[%s] %s", weaponTypePath, err.Error())
				}

				r.EnergyWeapons[fileName] = m

			}
		}
	}

	return nil
}

func (r *ModelResources) GetMechResource(unit string) *ModelMechResource {
	if m, ok := r.Mechs[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetVehicleResource(unit string) *ModelVehicleResource {
	if m, ok := r.Vehicles[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetVTOLResource(unit string) *ModelVTOLResource {
	if m, ok := r.VTOLs[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetInfantryResource(unit string) *ModelInfantryResource {
	if m, ok := r.Infantry[unit]; ok {
		return m
	}
	return nil
}

func (r *ModelResources) GetEnergyWeaponResource(weapon string) *ModelEnergyWeaponResource {
	if m, ok := r.EnergyWeapons[weapon]; ok {
		return m
	}
	return nil
}

func filesInPath(path string) ([]fs.DirEntry, error) {
	top, err := os.Open(path)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	files, err := top.ReadDir(0)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	return files, nil
}
