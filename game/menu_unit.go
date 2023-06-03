package game

import (
	"fmt"
	"sort"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
)

type UnitMenu struct {
	*MenuModel
}

func createUnitMenu(g *Game) *UnitMenu {
	var ui *ebitenui.UI = &ebitenui.UI{}

	menu := &UnitMenu{
		MenuModel: &MenuModel{
			game:   g,
			ui:     ui,
			active: true,
		},
	}

	menu.initResources()
	menu.initMenu()

	return menu
}

func (m *UnitMenu) initMenu() {
	m.MenuModel.initMenu()
	m.root.BackgroundImage = m.Resources().background

	// menu title
	titleBar := unitTitleContainer(m)
	m.root.AddChild(titleBar)

	// unit selection
	selection := unitSelectionContainer(m)
	m.root.AddChild(selection)

	// footer
	footer := unitMenuFooterContainer(m)
	m.root.AddChild(footer)
}

func (m *UnitMenu) Update() {
	m.ui.Update()
}

func (m *UnitMenu) Draw(screen *ebiten.Image) {
	m.ui.Draw(screen)
}

func unitTitleContainer(m *UnitMenu) *widget.Container {
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	c.AddChild(widget.NewText(
		widget.TextOpts.Text("Unit Selection", res.text.bigTitleFace, res.text.idleColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	))

	return c
}

func unitMenuFooterContainer(m *UnitMenu) *widget.Container {
	game := m.Game()
	res := m.Resources()

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.panel.titleBar),
		widget.ContainerOpts.Layout(widget.NewGridLayout(widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Stretch([]bool{false, true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:   m.Padding(),
				Right:  m.Padding(),
				Top:    m.Padding(),
				Bottom: m.Padding(),
			}))))

	back := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Back", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			iScene, _ := game.scene.(*InstantActionScene)
			iScene.back()
		}),
	)
	c.AddChild(back)

	c.AddChild(newBlankSeparator(m, widget.RowLayoutData{
		Stretch: true,
	}))

	next := widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.button.image),
		widget.ButtonOpts.Text("Next", res.button.face, res.button.text),
		widget.ButtonOpts.TextPadding(res.button.padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			iScene, _ := game.scene.(*InstantActionScene)
			iScene.next()
		}),
	)
	c.AddChild(next)

	return c
}

type unitPageContainer struct {
	widget    widget.PreferredSizeLocateableWidget
	titleText *widget.Text
	flipBook  *widget.FlipBook
}

type unitPage struct {
	title   string
	content widget.PreferredSizeLocateableWidget
	unit    *model.ModelMechResource // TODO: any resource
}

func newUnitPageContainer(m *UnitMenu) *unitPageContainer {
	res := m.Resources()

	c := widget.NewContainer(
		// background image will instead be set based on which page is showing
		//widget.ContainerOpts.BackgroundImage(res.panel.image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(res.panel.padding),
			widget.RowLayoutOpts.Spacing(m.Spacing()))),
	)

	titleText := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text("", res.text.titleFace, res.text.idleColor))
	c.AddChild(titleText)

	flipBook := widget.NewFlipBook(
		widget.FlipBookOpts.ContainerOpts(widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		}))),
	)
	c.AddChild(flipBook)

	return &unitPageContainer{
		widget:    c,
		titleText: titleText,
		flipBook:  flipBook,
	}
}

func (p *unitPageContainer) setPage(page *unitPage) {
	p.titleText.Label = page.unit.Name
	p.flipBook.SetPage(page.content)
	p.flipBook.RequestRelayout()
}

func unitSelectionPage(m *UnitMenu, unit *model.ModelMechResource) *unitPage {
	c := newPageContentContainer()
	// res := m.Resources()
	// game := m.Game()

	// TODO: more content

	return &unitPage{
		title:   fmt.Sprintf("%0.0f - %s", unit.Tonnage, unit.Name),
		content: c,
		unit:    unit,
	}
}

func unitSelectionContainer(m *UnitMenu) widget.PreferredSizeLocateableWidget {
	res := m.Resources()
	game := m.Game()

	chassisList := []string{}
	chassisMap := make(map[string][]*model.ModelMechResource, 32)
	for _, unit := range game.resources.GetMechResourceList() {
		chassis := unit.Name
		_, found := chassisMap[chassis]
		if !found {
			chassisList = append(chassisList, chassis)
			chassisMap[chassis] = make([]*model.ModelMechResource, 0, 4)
		}
		chassisMap[chassis] = append(chassisMap[chassis], unit)
	}

	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Padding(widget.Insets{
				Left:  m.Spacing(),
				Right: m.Spacing(),
			}),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, []bool{true}),
			widget.GridLayoutOpts.Spacing(m.Spacing(), 0),
		)))

	// sort by weight and then chassis name
	sort.Slice(chassisList, func(i, j int) bool {
		unitA, unitB := chassisMap[chassisList[i]][0], chassisMap[chassisList[j]][0]
		if unitA.Tonnage == unitB.Tonnage {
			return unitA.Name < unitB.Name
		}
		return unitA.Tonnage < unitB.Tonnage
	})

	pages := make([]interface{}, 0, len(chassisMap))
	for _, chassis := range chassisList {
		unitList := chassisMap[chassis]
		unitPage := unitSelectionPage(m, unitList[0]) // TODO: handle variant selection
		pages = append(pages, unitPage)
	}

	pageContainer := newUnitPageContainer(m)

	pageList := widget.NewList(
		widget.ListOpts.Entries(pages),
		widget.ListOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(*unitPage).title
		}),
		widget.ListOpts.ScrollContainerOpts(widget.ScrollContainerOpts.Image(res.list.image)),
		widget.ListOpts.SliderOpts(
			widget.SliderOpts.Images(res.list.track, res.list.handle),
			widget.SliderOpts.MinHandleSize(res.list.handleSize),
			widget.SliderOpts.TrackPadding(res.list.trackPadding),
		),
		widget.ListOpts.EntryColor(res.list.entry),
		widget.ListOpts.EntryFontFace(res.list.face),
		widget.ListOpts.EntryTextPadding(res.list.entryPadding),
		widget.ListOpts.HideHorizontalSlider(),

		widget.ListOpts.EntrySelectedHandler(func(args *widget.ListEntrySelectedEventArgs) {
			nextPage := args.Entry.(*unitPage)
			pageContainer.setPage(nextPage)
			m.Root().RequestRelayout()
		}))

	c.AddChild(pageList)

	c.AddChild(pageContainer.widget)

	return c
}
