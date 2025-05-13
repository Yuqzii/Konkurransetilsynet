package guessTheFunction

import (
	"encoding/json"
	"fmt"
)

type ExprWrapper struct {
	Type     string          `json:"type"`
	Value    json.RawMessage `json:"value,omitempty"`
	Left     *ExprWrapper    `json:"left,omitempty"`
	Right    *ExprWrapper    `json:"right,omitempty"`
	Base     *ExprWrapper    `json:"base,omitempty"`
	Exponent *ExprWrapper    `json:"exponent,omitempty"`
}

func marshalExpr(expr Expr) (*ExprWrapper, error) {
	// Marshal based on type of expr
	switch v := expr.(type) {
	case Number:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:  "Number",
			Value: data,
		}, nil
	case Variable:
		return &ExprWrapper{
			Type: "Variable",
		}, nil
	case Add:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:  "Add",
			Left:  left,
			Right: right,
		}, nil
	case Subtract:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:  "Subtract",
			Left:  left,
			Right: right,
		}, nil
	case Multiply:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:  "Multiply",
			Left:  left,
			Right: right,
		}, nil
	case Divide:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:  "Divide",
			Left:  left,
			Right: right,
		}, nil
	case Power:
		base, err := marshalExpr(v.Base)
		if err != nil {
			return nil, err
		}
		exponent, err := marshalExpr(v.Exponent)
		if err != nil {
			return nil, err
		}
		return &ExprWrapper{
			Type:     "Power",
			Base:     base,
			Exponent: exponent,
		}, nil
	default:
		return nil, fmt.Errorf("unknown expr type")
	}
}

func MarshalExpr(e Expr) ([]byte, error) {
	wrapper, err := marshalExpr(e)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wrapper)
}

func unmarshalExpr(wrapper *ExprWrapper) (Expr, error) {
	// Marshal based on type of expr
	switch wrapper.Type {
	case "Number":
		var num Number
		err := json.Unmarshal(wrapper.Value, &num)
		if err != nil {
			return nil, err
		}
		return num, nil
	case "Variable":
		var variable Variable
		return variable, nil
	case "Add":
		left, err := unmarshalExpr(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExpr(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return Add{Left: left, Right: right}, nil
	case "Subtract":
		left, err := unmarshalExpr(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExpr(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return Subtract{Left: left, Right: right}, nil
	case "Multiply":
		left, err := unmarshalExpr(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExpr(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return Multiply{Left: left, Right: right}, nil
	case "Divide":
		left, err := unmarshalExpr(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExpr(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return Divide{Left: left, Right: right}, nil
	case "Power":
		base, err := unmarshalExpr(wrapper.Base)
		if err != nil {
			return nil, err
		}
		exponent, err := unmarshalExpr(wrapper.Exponent)
		if err != nil {
			return nil, err
		}
		return Power{Base: base, Exponent: exponent}, nil
	default:
		return nil, fmt.Errorf("unknown type %q", wrapper.Type)
	}
}

func UnmarshalExpr(data []byte) (Expr, error) {
	var wrapper ExprWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return unmarshalExpr(&wrapper)
}
