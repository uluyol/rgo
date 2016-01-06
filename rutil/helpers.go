package rutil

import (
	"strings"

	"github.com/uluyol/rgo"
)

type GraphCfg struct {
	v []string
}

func (g GraphCfg) addKV(k, v string) GraphCfg {
	g.v = append(g.v, k+`="`+v+`"`)
	return g
}

func (g GraphCfg) params() string {
	var s string
	if len(g.v) > 0 {
		s = ", " + strings.Join(g.v, ", ")
	}
	return s
}

func (g GraphCfg) WithCol(color string) GraphCfg { return g.addKV("col", color) }
func (g GraphCfg) WithType(t string) GraphCfg    { return g.addKV("type", t) }

func plotCommon(rc *rgo.Conn, funcName string, x, y []float64, g GraphCfg) error {
	rc.Send(x, "go.x")
	rc.Send(y, "go.y")
	return rc.Rf("%s(go.x, go.y%s)", funcName, g.params())
}

func genRange(n int) []float64 {
	r := make([]float64, n)
	for i := 1; i <= n; i++ {
		r[i-1] = float64(i)
	}
	return r
}

func Plot(rc *rgo.Conn, x, y []float64, cfg GraphCfg) error {
	return plotCommon(rc, "plot", x, y, cfg)
}

func Lines(rc *rgo.Conn, x, y []float64, cfg GraphCfg) error {
	return plotCommon(rc, "lines", x, y, cfg)
}

func PlotX(rc *rgo.Conn, x []float64, cfg GraphCfg) error {
	return Plot(rc, genRange(len(x)), x, cfg)
}

func LinesX(rc *rgo.Conn, x []float64, cfg GraphCfg) error {
	return Lines(rc, genRange(len(x)), x, cfg)
}
