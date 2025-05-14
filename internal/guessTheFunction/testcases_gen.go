//go:generate go run ../../tools/gen_guessTheFunction_testcases

package guessTheFunction

var TestCasesLinear = []*TestCase{
	{"x+10", add{
		variable{}, number{Value: 10}}},

	{"x+1", add{
		variable{}, number{Value: 1}}},

	{"x-31", subtract{
		variable{}, number{Value: 31}}},

	{"x-9", subtract{
		variable{}, number{Value: 9}}},

	{"x+13+3", add{
		variable{}, add{number{Value: 13}, number{Value: 3}}}},

	{"12+x-2", add{
		number{Value: 12}, subtract{variable{}, number{Value: 2}},
	}},

	{"x-210+x+1", add{
		subtract{
			variable{},
			number{Value: 210},
		},
		add{
			variable{},
			number{Value: 1},
		}},
	},

	{"3*x+2", add{
		multiply{number{Value: 3}, variable{}},
		number{Value: 2},
	}},

	{"10*x-9", subtract{
		multiply{number{Value: 10}, variable{}},
		number{Value: 9},
	}},

	{"-10*x+3", add{
		multiply{number{Value: -10}, variable{}},
		number{Value: 3},
	}},

	{"x*-5+10", add{
		multiply{variable{}, number{Value: -5}},
		number{Value: 10},
	}},

	{"-1+7*x", add{
		number{Value: -1},
		multiply{number{Value: 7}, variable{}},
	}},

	{"-1", number{Value: -1}},

	{"1+-x", add{
		number{Value: 1},
		multiply{number{Value: -1}, variable{}},
	}},
}

var TestCasesPolynomial = []*TestCase{
	{"x^3+x^2+x^1+x^0", add{
		add{
			power{variable{}, number{Value: 3}},
			power{variable{}, number{Value: 2}},
		},
		add{
			power{variable{}, number{Value: 1}},
			power{variable{}, number{Value: 0}},
		},
	}},
	{"(2 / 7)*x^(4 * 3 / 0.8) + 3*x^2 + 4*x^1 + 5*x^0", add{
		add{
			multiply{number{(2.0 / 7.0)}, power{variable{},
				divide{
					multiply{
						number{4},
						number{3},
					},
					number{0.8},
				}}},
			multiply{number{3}, power{variable{}, number{Value: 2}}},
		},
		add{
			multiply{number{4}, power{variable{}, number{Value: 1}}},
			multiply{number{5}, power{variable{}, number{Value: 0}}},
		},
	}},
}
