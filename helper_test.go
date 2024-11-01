package scene_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func runForTest(t *testing.T, game ebiten.Game) {
	dummyScreen := ebiten.NewImage(3, 3)
	dummyFinalScreen := ebiten.NewImage(3, 3)

	for range 100 { // loop 100 times to avoid inf loop
		if l, ok := game.(ebiten.LayoutFer); ok {
			l.LayoutF(0, 0)
		} else {
			game.Layout(0, 0)
		}

		err := game.Update()

		if errors.Is(err, ebiten.Termination) {
			break
		}
		if err != nil {
			t.Fatalf("unexpected err on Game.Update(): %v", err)
		}

		game.Draw(dummyScreen)
		if f, ok := game.(ebiten.FinalScreenDrawer); ok {
			f.DrawFinalScreen(dummyFinalScreen, dummyScreen, ebiten.GeoM{})
		}
	}
}

type recorder struct {
	logs []string
}

func (r *recorder) Append(name, logType string) {
	r.logs = append(r.logs, fmt.Sprintf("%s:%s", name, logType))
}

type gameForTest struct {
	Name             string
	UpdateFn         func() error
	recorder         *recorder
	layoutW, layoutH int
}

func (g *gameForTest) append(logType string) {
	g.recorder.Append(g.Name, logType)
}

func (g *gameForTest) Update() error {
	g.append("update")
	if g.UpdateFn == nil {
		return ebiten.Termination
	}
	return g.UpdateFn()
}

func (g *gameForTest) Draw(screen *ebiten.Image) {
	g.append("draw")
}

func (g *gameForTest) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	g.append("layout")
	return g.layoutW, g.layoutH
}

type eventsForTest struct {
	gameForTest
}

func (e *eventsForTest) OnStart() {
	e.gameForTest.append("onstart")
}

func (e *eventsForTest) OnArrival() {
	e.gameForTest.append("onarrival")
}

func (e *eventsForTest) OnDeparture() {
	e.gameForTest.append("ondeparture")
}

func (e *eventsForTest) OnEnd() {
	e.gameForTest.append("onend")
}

type finalScreenDrawerForTest struct {
	gameForTest
}

func (f *finalScreenDrawerForTest) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	f.gameForTest.append("drawfinalscreen")
}

type layoutFerForTest struct {
	gameForTest
	layoutFW, layoutFH float64
}

func (l *layoutFerForTest) LayoutF(outsideWidth float64, outsideHeight float64) (screenWidth float64, screenHeight float64) {
	l.gameForTest.append("layoutf")
	return l.layoutFW, l.layoutFH
}

type transitionForTest struct {
	Name         string
	recorder     *recorder
	switchFrames int
	maxFrames    int
	currentFrame int
}

func (t *transitionForTest) Reset() {
	t.currentFrame = 0
	t.recorder.Append(t.Name, "reset")
}

func (t *transitionForTest) Update() error {
	t.recorder.Append(t.Name, "update")
	if t.currentFrame < t.maxFrames {
		t.currentFrame++
	}
	return nil
}

func (t *transitionForTest) Draw(screen *ebiten.Image) {
	t.recorder.Append(t.Name, "draw")
}

func (t *transitionForTest) Completed() bool {
	return t.currentFrame >= t.maxFrames
}

func (t *transitionForTest) CanSwitchScenes() bool {
	return t.currentFrame >= t.switchFrames
}

// from: https://github.com/hajimehoshi/ebiten/blob/main/internal/testing/testing.go
type game struct {
	m    *testing.M
	code int
}

func (g *game) Update() error {
	g.code = g.m.Run()
	return ebiten.Termination
}

func (*game) Draw(*ebiten.Image) {
}

func (*game) Layout(int, int) (int, int) {
	return 3, 3
}

func MainWithRunLoop(m *testing.M) {
	// Run an Ebiten process so that (*Image).At is available.
	g := &game{
		m:    m,
		code: 1,
	}
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
	if g.code != 0 {
		os.Exit(g.code)
	}
}

func TestMain(m *testing.M) {
	MainWithRunLoop(m)
}
