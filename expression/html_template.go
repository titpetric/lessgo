package expression

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/titpetric/lessgo/expression/functions"
)

var funcMap template.FuncMap
var cachedFunctionNames []string

func init() {
	funcMap = make(template.FuncMap)
	registerFunctions()
	// Pre-compute function names once
	cachedFunctionNames = make([]string, 0, len(funcMap))
	for k := range funcMap {
		cachedFunctionNames = append(cachedFunctionNames, k)
	}
}

func FuncMap() template.FuncMap {
	return funcMap
}

// GetRegisteredFunctionNames returns a list of all registered function names
func GetRegisteredFunctionNames() []string {
	return cachedFunctionNames
}

func registerFunctions() {
	// Register all functions from the "functions" package
	// This is a simplified example; a real implementation might use reflection
	// to discover functions automatically.
	// These functions are part of the 'expr' package's evaluation logic or Value methods, not directly exposed as standalone functions in the 'functions' package for template use.
	// register("add", functions.Add)
	// register("subtract", functions.Subtract)
	// register("multiply", functions.Multiply) // Conflicts with func_colors.go's Multiply
	// register("divide", functions.Divide)

	register("percentage", functions.Percentage)
	register("rgb", functions.RGB)
	register("hsl", functions.HSL)
	register("hsla", functions.HSLA)
	register("hsv", functions.HSV)
	register("hsva", functions.HSVA)
	register("saturate", functions.Saturate)
	register("desaturate", functions.Desaturate)
	register("lighten", functions.Lighten)
	register("darken", functions.Darken)
	register("fadein", functions.Fadein)
	register("fadeout", functions.Fadeout)
	register("fade", functions.Fade)
	register("spin", functions.Spin)
	register("mix", functions.Mix)
	register("hue", functions.Hue)
	register("saturation", functions.Saturation)
	register("lightness", functions.Lightness)
	register("alpha", functions.Alpha)
	register("luma", functions.LumaFunction)
	register("luminance", functions.Luminance)
	register("greyscale", functions.Greyscale)
	register("shade", functions.Shade)
	register("tint", functions.Tint)
	register("multiply", functions.Multiply)
	register("screen", functions.Screen)
	register("overlay", functions.Overlay)
	register("softlight", functions.Softlight)
	register("hardlight", functions.Hardlight)
	register("difference", functions.Difference)
	register("exclusion", functions.Exclusion)
	register("average", functions.Average)
	register("contrast", functions.Contrast)
	register("red", functions.Red)
	register("green", functions.Green)
	register("blue", functions.Blue)
	register("argb", functions.ARGB)
	register("length", functions.Length)
	register("isnumber", functions.IsNumberFunction)
	register("isstring", functions.IsStringFunction)
	register("iscolor", functions.IsColorFunction)
	register("iskeyword", functions.IsKeywordFunction)
	register("isurl", functions.IsURLFunction)
	register("ispixel", functions.IsPixelFunction)
	register("ispercentage", functions.IsPercentageFunction)
	register("isem", functions.IsEmFunction)
	// register("isrem", functions.IsRem) // Not found in functions package
	register("isunit", functions.IsUnitFunction)
	register("boolean", functions.Boolean)
	register("round", functions.Round)
	register("ceil", functions.Ceil)
	register("floor", functions.Floor)
	register("abs", functions.Abs)
	register("min", functions.Min)
	register("max", functions.Max)
	register("sqrt", functions.Sqrt)
	register("pow", functions.Pow)
	register("mod", functions.Mod)
	register("sin", functions.Sin)
	register("cos", functions.Cos)
	register("tan", functions.Tan)
	register("asin", functions.Asin)
	register("acos", functions.Acos)
	register("atan", functions.Atan)
	register("pi", functions.Pi)
	register("escape", functions.Escape)
	register("e", functions.E)
	register("replace", functions.Replace)
	register("format", functions.Format)
	register("if", functions.If)
	register("range", functions.Range)
	register("extract", functions.Extract)
	register("unit", functions.Unit)
	register("convert", functions.Convert)
	register("getunit", functions.GetUnit)

	// Additional missing functions
	register("negation", functions.Negation)
	register("colorfunction", functions.ColorFunction)
	register("color", functions.ColorFunction) // alias
	register("hsvhue", functions.HSVHue)
	register("hsvsaturation", functions.HSVSaturation)
	register("hsvvalue", functions.HSVValue)
	register("isdefined", functions.IsDefined)
	register("islist", functions.IsList)
	register("islistfunction", functions.IsListFunction)
	register("isnumberfunction", functions.IsNumberFunction)
	register("isstringfunction", functions.IsStringFunction)
	register("iscolorfunction", functions.IsColorFunction)
	register("iskeywordfunction", functions.IsKeywordFunction)
	register("isurlfunc", functions.IsURLFunction)
	register("ispixelfunction", functions.IsPixelFunction)
	register("isemfunction", functions.IsEmFunction)
	register("ispercentagefunction", functions.IsPercentageFunction)
	register("isunitfunction", functions.IsUnitFunction)
	register("isruleset", functions.IsRuleset)
	register("isrulesetfunction", functions.IsRulesetFunction)
}

func register(name string, fn any) {
	funcMap[name] = func(args ...any) (any, error) {
		fnVal := reflect.ValueOf(fn)
		fnType := fnVal.Type()

		if fnType.Kind() != reflect.Func {
			return nil, fmt.Errorf("not a function: %s", name)
		}

		// Check if the number of arguments is valid
		if fnType.IsVariadic() {
			if len(args) < fnType.NumIn()-1 {
				return nil, fmt.Errorf("not enough arguments for %s", name)
			}
		} else {
			if len(args) != fnType.NumIn() {
				return nil, fmt.Errorf("wrong number of arguments for %s: got %d, want %d", name, len(args), fnType.NumIn())
			}
		}

		// Convert arguments to the required types
		in := make([]reflect.Value, 0)
		for i, arg := range args {
			var targetType reflect.Type
			if fnType.IsVariadic() && i >= fnType.NumIn()-1 {
				// For variadic args, use the element type
				targetType = fnType.In(fnType.NumIn() - 1).Elem()
			} else {
				targetType = fnType.In(i)
			}
			argVal := reflect.ValueOf(arg)
			if argVal.Type().ConvertibleTo(targetType) {
				in = append(in, argVal.Convert(targetType))
			} else {
				return nil, fmt.Errorf("cannot convert argument %d for %s from %s to %s", i+1, name, argVal.Type(), targetType)
			}
		}

		// Call the function
		out := fnVal.Call(in)

		// Handle the return values
		if len(out) == 0 {
			return nil, nil
		}

		if len(out) > 2 {
			return nil, fmt.Errorf("function %s returns too many values", name)
		}

		var result any
		var err error

		// The first return value is the result
		result = out[0].Interface()

		// The second return value (optional) is an error
		if len(out) == 2 {
			if errVal, ok := out[1].Interface().(error); ok {
				err = errVal
			}
		}
		return result, err
	}
}

func Call(name string, args ...any) (any, error) {
	name = strings.ToLower(name)
	// Normalize function names: remove hyphens
	name = strings.ReplaceAll(name, "-", "")
	if fn, ok := funcMap[name]; ok {
		return fn.(func(...any) (any, error))(args...)
	}
	return nil, fmt.Errorf("unknown function call: %s", name)
}
