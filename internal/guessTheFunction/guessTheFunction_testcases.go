package guessTheFunction

var TestCases_FunctionParsing = []TestCase{
	{"x+1", add{
		variable{}, number{1}}},

	{"x+1.02020201111", add{
		variable{}, number{1.02020201111}}},

	{"x-3.1415", subtract{
		variable{}, number{3.1415}}},

	{"x+(-1.3)", subtract{
		variable{}, number{1.3}}},

	{"x+13-3+(-1.3+(-1.3)) + x", add{
		multiply{number{2}, variable{}},
		number{7.4}}},

	{"12+x-2", add{
		number{12}, subtract{variable{}, number{2}},
	}},

	{"x-210+x+1", add{
		subtract{
			variable{},
			number{210},
		},
		add{
			variable{},
			number{1},
		}},
	},

	{"3*x+2", add{
		multiply{number{3}, variable{}},
		number{2},
	}},

	{"10*x-9", subtract{
		multiply{number{10}, variable{}},
		number{9},
	}},

	{"-10*x+3", add{
		multiply{number{-10}, variable{}},
		number{3},
	}},

	{"x*-5+10", add{
		multiply{variable{}, number{-5}},
		number{10},
	}},

	{"x*(-5+(1-(-(2))+3)+1-(11*10-(12))+10)+12+-3.13*(3.1222)", add{
		multiply{number{-86}, variable{}},
		number{12 - 3.13*3.1222},
	}},

	{"-1+7*x", add{
		number{-1},
		multiply{number{7}, variable{}},
	}},

	{"-1", number{-1}},

	{"1+-x", add{
		number{1},
		multiply{number{-1}, variable{}},
	}},

	{"x^3+x^2+x^1+x^0", add{
		add{
			power{variable{}, number{3}},
			power{variable{}, number{2}},
		},
		add{
			power{variable{}, number{1}},
			power{variable{}, number{0}},
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
			multiply{number{3}, power{variable{}, number{2}}},
		},
		add{
			multiply{number{4}, power{variable{}, number{1}}},
			multiply{number{5}, power{variable{}, number{0}}},
		},
	}},

	// shortens to -x + 13
	{"(((((x+(2-(3-(4+(5-2*x)))))*(1+1))-((7*(2+1))-5))+((10-(3+2))*4))/2)+((6-(2+1))*(x-x+(1*(2-1))))", add{
		multiply{number{-1}, variable{}},
		number{13},
	}},

	// shortens to 3x^2 + 5x - 6
	{"((x*(2+1))*(x+0))+((10-5)*x)+((3-10)+((2*1)-1))", add{
		add{
			multiply{number{3}, power{Base: variable{}, Exponent: number{2}}},
			multiply{number{5}, variable{}},
		},
		number{-6},
	}},

	// shortens to (68 - 59x - 19x^2 + 9x^3 + x^4)/(3 - 6x + 5x^2)
	{"(x-1)*((((x+2)*(x-3)+(4*x-(2-x)))*(x+1))+((3*(x-5)*(x+4))-(2*x*(1-x))))/(((x-1)^2+(x^(2)-2*x+1))+(((x+2)*(x-2))-((x*x)-4))+(3*x*x-2*x+1))", divide{
		add{
			add{
				add{
					add{
						number{68},
						multiply{number{-59}, variable{}},
					},
					multiply{
						number{-19},
						power{Base: variable{}, Exponent: number{2}},
					},
				},
				multiply{
					number{9},
					power{Base: variable{}, Exponent: number{3}},
				},
			},
			power{Base: variable{}, Exponent: number{4}},
		},
		add{
			add{
				number{3},
				multiply{number{-6}, variable{}},
			},
			multiply{
				number{5},
				power{Base: variable{}, Exponent: number{2}},
			},
		},
	}},
}

var TestCases_SavingLoading = [...]expr{
	add{
		variable{}, number{Value: 10}},

	add{
		variable{}, number{Value: 1}},

	subtract{
		variable{}, number{Value: 31}},

	subtract{
		variable{}, number{Value: 9}},

	add{
		variable{}, add{number{Value: 13}, number{Value: 3}}},

	add{
		number{Value: 12}, subtract{variable{}, number{Value: 2}},
	},

	add{
		subtract{
			variable{},
			number{Value: 210},
		},
		add{
			variable{},
			number{Value: 1},
		}},

	add{
		multiply{number{Value: 3}, variable{}},
		number{Value: 2},
	},

	subtract{
		multiply{number{Value: 10}, variable{}},
		number{Value: 9},
	},

	add{
		multiply{number{Value: -10}, variable{}},
		number{Value: 3},
	},

	add{
		multiply{variable{}, number{Value: -5}},
		number{Value: 10},
	},

	add{
		number{Value: -1},
		multiply{number{Value: 7}, variable{}},
	},

	number{Value: -1},

	add{
		number{Value: 1},
		multiply{number{Value: -1}, variable{}},
	},
}
