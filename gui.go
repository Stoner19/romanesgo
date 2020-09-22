// +build gui

package main

import (
    "flag"
    "fmt"
    "log"
    "runtime"
    "strconv"
    "strings"
    "time"

    rv "github.com/TheGrum/renderview"
    "github.com/TheGrum/renderview/driver"
)

func ParseConstants(value string) []float64 {
    constants := strings.Split(value, ",")
    var c []float64 = make([]float64, 0, len(constants))
    for _, s := range constants {
        val, err := strconv.ParseFloat(s, 64)
        if err != nil {
            log.Printf("Error parsing %v as a float64", s)
        } else {
            c = append(c, val)
        }
    }
    return c
}

// If people are going to be typing these values,
// we need to be more tolerant of invalid values
func getFractalFunctionTolerant(fractalName, colouringFuncName string, constants []float64) func(float64, float64, int) (R, G, B, A float64) {
    fractalFuncUnasserted, valid := fractals[fractalName] //Asserted after validation because if the fractal function's wrong, we'd try to assert nil.
    if valid != true {
        fmt.Println("Invalid fractal function.")
        return nil
    }
    fractalFunc := fractalFuncUnasserted.(map[string]interface{})["func"].(func(interface{}, []float64) func(float64, float64, int) (float64, float64, float64, float64))

    if len(constants) != fractals[fractalName].(map[string]interface{})["constants"].(int) {
        fmt.Println("Invalid amount of constants.")
        return nil
    }

    colouringFunc, valid := fractals[fractalName].(map[string]interface{})["colourfuncs"].(map[string]interface{})[colouringFuncName]

    if valid != true {
        fmt.Println("Invalid colouring function.")
        return nil
    }

    return fractalFunc(colouringFunc, constants)
}

func main() {
    flag.Parse()

    m := rv.NewBasicRenderModel()
    m.AddParameters(rv.SetHints(rv.HINT_SIDEBAR,
        rv.NewStringRP("fractal function", "mandelbrot"),
        rv.NewStringRP("colour function", "default"),
        rv.NewStringRP("constants", ""),
        rv.NewIntRP("iterations", 128),
        rv.NewIntRP("samples", 1),
        rv.NewFloat64RP("zoom", 1),
        rv.NewIntRP("routines", runtime.NumCPU()))...)
    m.AddParameters(rv.DefaultParameters(false, rv.HINT_SIDEBAR, rv.OPT_AUTO_ZOOM, 0, 0, 2, 2)...)

    fnP := m.GetParameter("fractal function")
    cfnP := m.GetParameter("colour function")
    iterP := m.GetParameter("iterations")
    smP := m.GetParameter("samples")
    wP := m.GetParameter("width")
    hP := m.GetParameter("height")
    lP := m.GetParameter("left")
    //  rP := m.GetParameter("right")
    tP := m.GetParameter("top")
    //  bP := m.GetParameter("bottom")
    rtP := m.GetParameter("routines")
    zP := m.GetParameter("zoom")
    constP := m.GetParameter("constants")

    m.InnerRender = func() {
        m.Rendering = true
        // Collect current parameter values
        //xc := ((rP.GetValueFloat64() - lP.GetValueFloat64()) / 2.0) + lP.GetValueFloat64()
        //yc := ((bP.GetValueFloat64() - tP.GetValueFloat64()) / 2.0) + tP.GetValueFloat64()
        xc := lP.GetValueFloat64()
        yc := tP.GetValueFloat64()
        f := getNewFractGen(wP.GetValueInt(), hP.GetValueInt(), rtP.GetValueInt(), iterP.GetValueInt(), xc, yc, zP.GetValueFloat64())
        consts := ParseConstants(constP.GetValueString())
        finalfun := getFractalFunctionTolerant(fnP.GetValueString(), cfnP.GetValueString(), consts)
        if finalfun != nil {
            f.generate(finalfun, smP.GetValueInt())
        }
        m.Img = f.fractImg
        m.Rendering = false
    }
    go func(m *rv.BasicRenderModel) {
        ticker := time.NewTicker(time.Millisecond * 250)
        for _ = range ticker.C {
            m.RequestPaint()
        }
    }(m)
    driver.Main(m)
}

func handleError(err error) {
    if !(err == nil) {
        log.Fatal(err)
    }
}
