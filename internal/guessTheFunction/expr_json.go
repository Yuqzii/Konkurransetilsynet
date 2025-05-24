package guessTheFunction

import (
	"encoding/json"
	"fmt"
)

type exprWrapper struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value,omitempty"`
	Left  *exprWrapper    `json:"left,omitempty"`
	Right *exprWrapper    `json:"right,omitempty"`
}

func marshalExpr(expr expr) (*exprWrapper, error) {
	// Marshal based on type of expr
	switch v := expr.(type) {
	case number:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Number",
			Value: data,
		}, nil
	case variable:
		return &exprWrapper{
			Type: "Variable",
		}, nil
	case add:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Add",
			Left:  left,
			Right: right,
		}, nil
	case subtract:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Subtract",
			Left:  left,
			Right: right,
		}, nil
	case multiply:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Multiply",
			Left:  left,
			Right: right,
		}, nil
	case divide:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Divide",
			Left:  left,
			Right: right,
		}, nil
	case power:
		left, err := marshalExpr(v.Left)
		if err != nil {
			return nil, err
		}
		right, err := marshalExpr(v.Right)
		if err != nil {
			return nil, err
		}
		return &exprWrapper{
			Type:  "Power",
			Left:  left,
			Right: right,
		}, nil
	default:
		return nil, fmt.Errorf("unknown expr type")
	}
}

func MarshalExpr(e expr) ([]byte, error) {
	wrapper, err := marshalExpr(e)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wrapper)
}

func unmarshalExprWrapper(wrapper *exprWrapper) (expr, error) {
	// Unmarshal based on type of expr
	switch wrapper.Type {
	case "Number":
		var num number
		err := json.Unmarshal(wrapper.Value, &num)
		if err != nil {
			return nil, err
		}
		return num, nil
	case "Variable":
		var variable variable
		return variable, nil
	case "Add":
		left, err := unmarshalExprWrapper(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExprWrapper(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return add{Left: left, Right: right}, nil
	case "Subtract":
		left, err := unmarshalExprWrapper(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExprWrapper(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return subtract{Left: left, Right: right}, nil
	case "Multiply":
		left, err := unmarshalExprWrapper(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExprWrapper(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return multiply{Left: left, Right: right}, nil
	case "Divide":
		left, err := unmarshalExprWrapper(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExprWrapper(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return divide{Left: left, Right: right}, nil
	case "Power":
		left, err := unmarshalExprWrapper(wrapper.Left)
		if err != nil {
			return nil, err
		}
		right, err := unmarshalExprWrapper(wrapper.Right)
		if err != nil {
			return nil, err
		}
		return power{Left: left, Right: right}, nil
	default:
		return nil, fmt.Errorf("unknown type %q", wrapper.Type)
	}
}

func unmarshalExpr(data []byte) (expr, error) {
	var wrapper exprWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return unmarshalExprWrapper(&wrapper)
}
