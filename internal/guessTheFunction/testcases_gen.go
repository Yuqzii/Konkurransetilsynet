//go:generate go run ../../tools/gen_guessTheFunction_testcases

package guessTheFunction

var TestCasesLinear = []*TestCase{
	{"x+10", Add{
		Variable{}, Number{Value: 10}}},

	{"x+1", Add{
		Variable{}, Number{Value: 1}}},

	{"x-31", Subtract{
		Variable{}, Number{Value: 31}}},

	{"x-9", Subtract{
		Variable{}, Number{Value: 9}}},

	{"x+13+3", Add{
		Variable{}, Add{Number{Value: 13}, Number{Value: 3}}}},

	{"12+x-2", Add{
		Number{Value: 12}, Subtract{Variable{}, Number{Value: 2}},
	}},

	{"x-210+x+1", Add{
		Subtract{
			Variable{},
			Number{Value: 210},
		},
		Add{
			Variable{},
			Number{Value: 1},
		}},
	},

	{"3*x+2", Add{
		Multiply{Number{Value: 3}, Variable{}},
		Number{Value: 2},
	}},

	{"10*x-9", Subtract{
		Multiply{Number{Value: 10}, Variable{}},
		Number{Value: 9},
	}},

	{"-10*x+3", Add{
		Multiply{Number{Value: -10}, Variable{}},
		Number{Value: 3},
	}},

	{"x*-5+10", Add{
		Multiply{Variable{}, Number{Value: -5}},
		Number{Value: 10},
	}},

	{"-1+7*x", Add{
		Number{Value: -1},
		Multiply{Number{Value: 7}, Variable{}},
	}},

	{"-1", Number{Value: -1}},

	{"1+-x", Add{
		Number{Value: 1},
		Multiply{Number{Value: -1}, Variable{}},
	}},
}

var TestCasesPolynomial = []*TestCase{
	{"x^3+x^2+x^1+x^0", Add{
		Add{
			Power{Variable{}, Number{Value: 3}},
			Power{Variable{}, Number{Value: 2}},
		},
		Add{
			Power{Variable{}, Number{Value: 1}},
			Power{Variable{}, Number{Value: 0}},
		},
	}},
	{"(2 / 7)*x^(4 * 3 / 0.8) + 3*x^2 + 4*x^1 + 5*x^0", Add{
		Add{
			Multiply{Number{(2.0 / 7.0)}, Power{Variable{},
				Divide{
					Multiply{
						Number{4},
						Number{3},
					},
					Number{0.8},
				}}},
			Multiply{Number{3}, Power{Variable{}, Number{Value: 2}}},
		},
		Add{
			Multiply{Number{4}, Power{Variable{}, Number{Value: 1}}},
			Multiply{Number{5}, Power{Variable{}, Number{Value: 0}}},
		},
	}},
}
