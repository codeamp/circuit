package promql

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/semantic"
)

type ArgKind int

const (
	IdentifierKind ArgKind = iota
	DurationKind
	ExprKind
	NumberKind
	StringKind
	SelectorKind
)

type QueryBuilder interface {
	QuerySpec() (*flux.Spec, error)
}

type Arg interface {
	Type() ArgKind
	Value() interface{}
}

type Identifier struct {
	Name string `json:"name,omitempty"`
}

func (id *Identifier) Type() ArgKind {
	return IdentifierKind
}

func (id *Identifier) Value() interface{} {
	return id.Name
}

func NewIdentifierList(first *Identifier, rest interface{}) ([]*Identifier, error) {
	ids := []*Identifier{first}
	for _, l := range toIfaceSlice(rest) {
		if id, ok := l.(*Identifier); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

type StringLiteral struct {
	String string `json:"string,omitempty"`
}

func (s *StringLiteral) Type() ArgKind {
	return StringKind
}

func (s *StringLiteral) Value() interface{} {
	return s.String
}

type Duration struct {
	Dur time.Duration `json:"dur,omitempty"`
}

func (d *Duration) Type() ArgKind {
	return DurationKind
}

func (d *Duration) Value() interface{} {
	return d.Dur
}

type Number struct {
	Val float64 `json:"val,omitempty"`
}

func (n *Number) Type() ArgKind {
	return NumberKind
}

func (n *Number) Value() interface{} {
	return n.Val
}

func NewNumber(val string) (*Number, error) {
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil, err
	}
	return &Number{num}, nil
}

// MatchKind is an enum for label matching types.
type MatchKind int

// Possible MatchKinds.
const (
	Equal MatchKind = iota
	NotEqual
	RegexMatch
	RegexNoMatch
)

type LabelMatcher struct {
	Name  string    `json:"name,omitempty"`
	Kind  MatchKind `json:"kind,omitempty"`
	Value Arg       `json:"value,omitempty"`
}

func NewLabelMatcher(ident *Identifier, kind MatchKind, value Arg) (*LabelMatcher, error) {
	return &LabelMatcher{
		Name:  ident.Name,
		Kind:  kind,
		Value: value,
	}, nil
}

func NewLabelMatches(first *LabelMatcher, rest interface{}) ([]*LabelMatcher, error) {
	matches := []*LabelMatcher{first}
	for _, m := range toIfaceSlice(rest) {
		if match, ok := m.(*LabelMatcher); ok {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

type Selector struct {
	Name          string          `json:"name,omitempty"`
	Range         time.Duration   `json:"range,omitempty"`
	Offset        time.Duration   `json:"offset,omitempty"`
	LabelMatchers []*LabelMatcher `json:"label_matchers,omitempty"`
}

func (s *Selector) QuerySpec() (*flux.Spec, error) {
	parent := "from"
	ops := []*flux.Operation{
		{
			ID: "from", // TODO: Change this to a UUID
			Spec: &inputs.FromOpSpec{
				Bucket: "prometheus",
			},
		},
	}
	edges := []flux.Edge{}

	rng, err := NewRangeOp(s.Range, s.Offset)
	if err != nil {
		return nil, err
	}

	if rng != nil {
		ops = append(ops, rng)
		edge := flux.Edge{
			Parent: flux.OperationID(parent),
			Child:  "range",
		}
		parent = "range"
		edges = append(edges, edge)
	}

	where, err := NewWhereOperation(s.Name, s.LabelMatchers)
	if err != nil {
		return nil, err
	}

	ops = append(ops, where)
	edge := flux.Edge{
		Parent: flux.OperationID(parent),
		Child:  "where",
	}
	parent = "where"
	edges = append(edges, edge)

	return &flux.Spec{
		Operations: ops,
		Edges:      edges,
	}, nil
}

func NewRangeOp(rng, offset time.Duration) (*flux.Operation, error) {
	if rng == 0 && offset == 0 {
		return nil, nil
	}
	return &flux.Operation{
		ID: "range", // TODO: Change this to a UUID
		Spec: &transformations.RangeOpSpec{
			Start: flux.Time{
				Relative: -rng - offset,
			},
		},
	}, nil
}

var operatorLookup = map[MatchKind]ast.OperatorKind{
	Equal:        ast.EqualOperator,
	NotEqual:     ast.NotEqualOperator,
	RegexMatch:   ast.EqualOperator,
	RegexNoMatch: ast.NotEqualOperator,
}

func NewWhereOperation(metricName string, labels []*LabelMatcher) (*flux.Operation, error) {
	var node semantic.Expression = &semantic.BinaryExpression{
		Operator: ast.EqualOperator,
		Left: &semantic.MemberExpression{
			Object: &semantic.IdentifierExpression{
				Name: "r",
			},
			Property: "_metric",
		},
		Right: &semantic.StringLiteral{
			Value: metricName,
		},
	}
	for _, label := range labels {
		op, ok := operatorLookup[label.Kind]
		if !ok {
			return nil, fmt.Errorf("unknown label match kind %d", label.Kind)
		}
		ref := &semantic.MemberExpression{
			Object: &semantic.IdentifierExpression{
				Name: "r",
			},
			Property: label.Name,
		}
		var value semantic.Expression
		if label.Value.Type() == StringKind {
			value = &semantic.StringLiteral{
				Value: label.Value.Value().(string),
			}
		} else if label.Value.Type() == NumberKind {
			value = &semantic.FloatLiteral{
				Value: label.Value.Value().(float64),
			}
		}
		node = &semantic.LogicalExpression{
			Operator: ast.AndOperator,
			Left:     node,
			Right: &semantic.BinaryExpression{
				Operator: op,
				Left:     ref,
				Right:    value,
			},
		}
	}

	return &flux.Operation{
		ID: "where", // TODO: Change this to a UUID
		Spec: &transformations.FilterOpSpec{
			Fn: &semantic.FunctionExpression{
				Block: &semantic.FunctionBlock{
					Parameters: &semantic.FunctionParameters{
						List: []*semantic.FunctionParameter{{Key: &semantic.Identifier{Name: "r"}}},
					},
					Body: node,
				},
			},
		},
	}, nil
}

func (s *Selector) Type() ArgKind {
	return SelectorKind
}

func (s *Selector) Value() interface{} {
	return s.Name // TODO: Change to AST
}

func NewSelector(metric *Identifier, block, rng, offset interface{}) (*Selector, error) {
	sel := &Selector{
		Name: metric.Name,
	}

	if block != nil {
		sel.LabelMatchers = block.([]*LabelMatcher)
	}

	if rng != nil {
		sel.Range = rng.(time.Duration)
	}

	if offset != nil {
		sel.Offset = offset.(time.Duration)
	}

	return sel, nil
}

type Aggregate struct {
	Without bool          `json:"without,omitempty"`
	By      bool          `json:"by,omitempty"`
	Labels  []*Identifier `json:"labels,omitempty"`
}

func (a *Aggregate) QuerySpec() (*flux.Operation, error) {
	if a.Without {
		return nil, fmt.Errorf("Unable to merge using `without`")
	}
	keys := make([]string, len(a.Labels))
	for i := range a.Labels {
		keys[i] = a.Labels[i].Name
	}
	return &flux.Operation{
		ID: "merge",
		Spec: &transformations.GroupOpSpec{
			By: keys,
		},
	}, nil
}

type OperatorKind int

const (
	UnknownOpKind OperatorKind = iota
	CountValuesKind
	TopKind
	BottomKind
	QuantileKind
	SumKind
	MinKind
	MaxKind
	AvgKind
	StdevKind
	StdVarKind
	CountKind
)

func ToOperatorKind(op string) OperatorKind {
	op = strings.ToLower(op)
	switch op {
	case "count_values":
		return CountValuesKind
	case "topk":
		return TopKind
	case "bottomk":
		return BottomKind
	case "quantile":
		return QuantileKind
	case "sum":
		return SumKind
	case "min":
		return MinKind
	case "max":
		return MaxKind
	case "avg":
		return AvgKind
	case "stddev":
		return StdevKind
	case "stdvar":
		return StdVarKind
	case "count":
		return CountKind
	default:
		return UnknownOpKind
	}
}

type Operator struct {
	Kind OperatorKind `json:"kind,omitempty"`
	Arg  Arg          `json:"arg,omitempty"`
}

func (o *Operator) QuerySpec() (*flux.Operation, error) {
	switch o.Kind {
	case CountValuesKind, BottomKind, QuantileKind, StdVarKind:
		return nil, fmt.Errorf("Unable to run %d yet", o.Kind)
	case CountKind:
		return &flux.Operation{
			ID:   "count",
			Spec: &transformations.CountOpSpec{},
		}, nil
	//case TopKind:
	//	return &flux.Operation{
	//		ID:   "top",
	//		Spec: &transformations.TopOpSpec{}, // TODO: Top doesn't have arg yet
	//	}, nil
	case SumKind:
		return &flux.Operation{
			ID:   "sum",
			Spec: &transformations.SumOpSpec{},
		}, nil
	//case MinKind:
	//	return &flux.Operation{
	//		ID:   "min",
	//		Spec: &transformations.MinOpSpec{},
	//	}, nil
	//case MaxKind:
	//	return &flux.Operation{
	//		ID:   "max",
	//		Spec: &transformations.MaxOpSpec{},
	//	}, nil
	//case AvgKind:
	//	return &flux.Operation{
	//		ID:   "mean",
	//		Spec: &transformations.MeanOpSpec{},
	//	}, nil
	//case StdevKind:
	//	return &flux.Operation{
	//		ID:   "stddev",
	//		Spec: &transformations.StddevOpSpec{},
	//	}, nil
	default:
		return nil, fmt.Errorf("Unknown Op kind %d", o.Kind)
	}
}

type AggregateExpr struct {
	Op        *Operator  `json:"op,omitempty"`
	Selector  *Selector  `json:"selector,omitempty"`
	Aggregate *Aggregate `json:"aggregate,omitempty"`
}

func (a *AggregateExpr) QuerySpec() (*flux.Spec, error) {
	spec, err := a.Selector.QuerySpec()
	if err != nil {
		return nil, err
	}

	if a.Aggregate != nil {
		agg, err := a.Aggregate.QuerySpec()
		if err != nil {
			return nil, err
		}

		parent := flux.OperationID("from")
		if len(spec.Edges) > 0 {
			tail := spec.Edges[len(spec.Edges)-1]
			parent = tail.Child
		}

		spec.Operations = append(spec.Operations, agg)
		spec.Edges = append(spec.Edges, flux.Edge{
			Parent: parent,
			Child:  agg.ID,
		})
	}

	op, err := a.Op.QuerySpec()
	if err != nil {
		return nil, err
	}

	parent := flux.OperationID("from")
	if len(spec.Edges) > 0 {
		tail := spec.Edges[len(spec.Edges)-1]
		parent = tail.Child
	}
	spec.Operations = append(spec.Operations, op)
	spec.Edges = append(spec.Edges, flux.Edge{
		Parent: parent,
		Child:  op.ID,
	})
	return spec, nil
}

func NewAggregateExpr(op *Operator, selector *Selector, group interface{}) (*AggregateExpr, error) {
	expr := &AggregateExpr{
		Op:       op,
		Selector: selector,
	}
	if group != nil {
		expr.Aggregate = group.(*Aggregate)
	}
	return expr, nil
}

type Comment struct {
	Source string `json:"source,omitempty"`
}

func (c *Comment) QuerySpec() (*flux.Spec, error) {
	return nil, fmt.Errorf("Unable to represent comments in the AST")
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
