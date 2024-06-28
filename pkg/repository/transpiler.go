package repository

import (
	"fmt"
	// "time"

	"github.com/iancoleman/strcase"
	"go.einride.tech/aip/filtering"
	"gorm.io/gorm/clause"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Transpiler data
type Transpiler struct {
	filter filtering.Filter
}

// Transpile executes the transpilation on the filter
func (t *Transpiler) Transpile() (string, error) {
	if t.filter.CheckedExpr == nil {
		return "", nil
	}
	expr, _, err := t.transpileExpr(t.filter.CheckedExpr.Expr)
	if err != nil {
		return "", err
	}
	return expr, nil
}

func (t *Transpiler) transpileExpr(e *expr.Expr) (string, string, error) {
	switch e.ExprKind.(type) {
	case *expr.Expr_CallExpr:
		ex, err := t.transpileCallExpr(e)
		return ex, "", err
	case *expr.Expr_IdentExpr:
		return t.transpileIdentExpr(e)
	case *expr.Expr_ConstExpr:
		ex, err := t.transpileConstExpr(e)
		return ex, "", err
	case *expr.Expr_SelectExpr:
		ex, err := t.transpileSelectExpr(e)
		return ex, "", err
	default:
		return "", "", fmt.Errorf("unsupported expr: %v", e)
	}
}

func (t *Transpiler) transpileConstExpr(e *expr.Expr) (string, error) {
	switch kind := e.GetConstExpr().ConstantKind.(type) {
	case *expr.Constant_BoolValue:
		return fmt.Sprintf("%v", kind.BoolValue), nil
	case *expr.Constant_DoubleValue:
		return fmt.Sprintf("%v", kind.DoubleValue), nil
	case *expr.Constant_Int64Value:
		return fmt.Sprintf("%v", kind.Int64Value), nil
	case *expr.Constant_StringValue:
		return kind.StringValue, nil
	case *expr.Constant_Uint64Value:
		return fmt.Sprintf("%v", kind.Uint64Value), nil

	default:
		return "", fmt.Errorf("unsupported const expr: %v", kind)
	}
}

func (t *Transpiler) transpileCallExpr(e *expr.Expr) (string, error) {
	switch e.GetCallExpr().Function {
	case filtering.FunctionHas:
		return t.transpileHasCallExpr(e)
	case filtering.FunctionEquals:
		return t.transpileComparisonCallExpr(e, clause.Eq{})
	case filtering.FunctionNotEquals:
		return t.transpileComparisonCallExpr(e, clause.Neq{})
	case filtering.FunctionLessThan:
		return t.transpileComparisonCallExpr(e, clause.Lt{})
	case filtering.FunctionLessEquals:
		return t.transpileComparisonCallExpr(e, clause.Lte{})
	case filtering.FunctionGreaterThan:
		return t.transpileComparisonCallExpr(e, clause.Gt{})
	case filtering.FunctionGreaterEquals:
		return t.transpileComparisonCallExpr(e, clause.Gte{})
	case filtering.FunctionAnd:
		return t.transpileBinaryLogicalCallExpr(e, clause.AndConditions{})
	case filtering.FunctionOr:
		return t.transpileBinaryLogicalCallExpr(e, clause.OrConditions{})
	case filtering.FunctionNot:
		return t.transpileNotCallExpr(e)
	case filtering.FunctionTimestamp:
		return t.transpileTimestampCallExpr(e)
	default:
		return "", fmt.Errorf("unsupported function call: %s", e.GetCallExpr().Function)
	}
}

func (t *Transpiler) transpileIdentExpr(e *expr.Expr) (string, string, error) {

	identExpr := e.GetIdentExpr()
	identExprName := strcase.ToSnake(identExpr.Name)
	identType, ok := t.filter.CheckedExpr.TypeMap[e.Id]
	if !ok {
		return "", "", fmt.Errorf("unknown type of ident expr %d", e.Id)
	}
	if messageType := identType.GetMessageType(); messageType != "" {
		if enumType, err := protoregistry.GlobalTypes.FindEnumByName(protoreflect.FullName(messageType)); err == nil {
			if enumValue := enumType.Descriptor().Values().ByName(protoreflect.Name(identExprName)); enumValue != nil {
				// TODO: Configurable support for string literals.
				return string(enumValue.Name()), messageType, nil
			}
		}
	}
	if wellKnown := identType.GetWellKnown(); wellKnown == expr.Type_TIMESTAMP {
		return string(identExpr.Name), expr.Type_TIMESTAMP.String(), nil
	}

	return string(identExpr.Name), identType.GetMessageType(), nil
}

func (t *Transpiler) transpileSelectExpr(e *expr.Expr) (string, error) {
	return "", fmt.Errorf("does not support SELECT expression for now")
}

func (t *Transpiler) transpileNotCallExpr(e *expr.Expr) (string, error) {
	callExpr := e.GetCallExpr()
	if len(callExpr.Args) != 1 {
		return "", fmt.Errorf(
			"unexpected number of arguments to `%s` expression: %d",
			filtering.FunctionNot,
			len(callExpr.Args),
		)
	}

	return "", fmt.Errorf("does not support NOT expression for now")
}

func (t *Transpiler) transpileComparisonCallExpr(e *expr.Expr, op interface{}) (string, error) {
	callExpr := e.GetCallExpr()
	if len(callExpr.Args) != 2 {
		return "", fmt.Errorf(
			"unexpected number of arguments to `%s`: %d",
			callExpr.GetFunction(),
			len(callExpr.Args),
		)
	}

	ident, idenType, err := t.transpileExpr(callExpr.Args[0])
	if err != nil {
		return "", err
	}
	ident = strcase.ToSnake(ident)

	con, _, err := t.transpileExpr(callExpr.Args[1])
	if err != nil {
		return "", err
	}

	if idenType == expr.Type_TIMESTAMP.String() {
		return fmt.Sprintf("%v@%v", ident, con), nil
	}

	var s string
	switch op.(type) {
	case clause.Eq:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] == \"%s\")", ident, con)
	case clause.Neq:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] != \"%s\")", ident, con)
	case clause.Lt:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] < \"%s\")", ident, con)
	case clause.Lte:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] <= \"%s\")", ident, con)
	case clause.Gt:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] > \"%s\")", ident, con)
	case clause.Gte:
		s = fmt.Sprintf("|> filter(fn: (r) => r[\"%s\"] >= \"%s\")", ident, con)
	}

	return s, nil
}

func (t *Transpiler) transpileBinaryLogicalCallExpr(e *expr.Expr, op clause.Expression) (string, error) {
	callExpr := e.GetCallExpr()
	if len(callExpr.Args) != 2 {
		return "", fmt.Errorf(
			"unexpected number of arguments to `%s`: %d",
			callExpr.GetFunction(),
			len(callExpr.Args),
		)
	}
	lhsExpr, _, err := t.transpileExpr(callExpr.Args[0])
	if err != nil {
		return "", err
	}
	rhsExpr, _, err := t.transpileExpr(callExpr.Args[1])
	if err != nil {
		return "", err
	}

	var s string
	switch op.(type) {
	case clause.AndConditions:
		s = fmt.Sprintf("%s&&%s", lhsExpr, rhsExpr)
	case clause.OrConditions:
		return "", fmt.Errorf("does not support OR logical op at the moment")
	}

	return s, nil
}

func (t *Transpiler) transpileHasCallExpr(e *expr.Expr) (string, error) {
	callExpr := e.GetCallExpr()
	if len(callExpr.Args) != 2 {
		return "", fmt.Errorf("unexpected number of arguments to `in` expression: %d", len(callExpr.Args))
	}

	if callExpr.Args[1].GetConstExpr() == nil {
		return "", fmt.Errorf("TODO: add support for transpiling `:` where RHS is other than Const")
	}

	switch callExpr.Args[0].ExprKind.(type) {
	case *expr.Expr_IdentExpr:
		identExpr := callExpr.Args[0]
		constExpr := callExpr.Args[1]
		identType, ok := t.filter.CheckedExpr.TypeMap[callExpr.Args[0].Id]
		if !ok {
			return "", fmt.Errorf("unknown type of ident expr %d", e.Id)
		}
		switch {
		// Repeated primitives:
		// > Repeated fields query to see if the repeated structure contains a matching element.
		case identType.GetListType().GetElemType().GetPrimitive() != expr.Type_PRIMITIVE_TYPE_UNSPECIFIED:
			iden, _, err := t.transpileIdentExpr(identExpr)
			if err != nil {
				return "", err
			}
			con, err := t.transpileConstExpr(constExpr)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("|> filter(fn: (r) => contains(value: \"%v\", set: \"%v\"))", iden, con), nil
		default:
			return "", fmt.Errorf("TODO: add support for transpiling `:` on other types than repeated primitives")
		}
	case *expr.Expr_SelectExpr:
		operand := callExpr.Args[0].GetSelectExpr().Operand

		switch operand.ExprKind.(type) {
		case *expr.Expr_IdentExpr:
		case *expr.Expr_SelectExpr:
		default:
			return "", fmt.Errorf("TODO: add support for more complicated transpiling")
		}

	default:
		return "", fmt.Errorf("TODO: add support for transpiling `:` where LHS is other than Ident and Select")
	}
	return "", fmt.Errorf("TODO: add support for more transpiling")
}

func (t *Transpiler) transpileTimestampCallExpr(e *expr.Expr) (string, error) {

	callExpr := e.GetCallExpr()
	if len(callExpr.Args) != 1 {
		return "", fmt.Errorf(
			"unexpected number of arguments to `%s`: %d", callExpr.Function, len(callExpr.Args),
		)
	}
	constArg, ok := callExpr.Args[0].ExprKind.(*expr.Expr_ConstExpr)
	if !ok {
		return "", fmt.Errorf("expected constant string arg to %s", callExpr.Function)
	}
	stringArg, ok := constArg.ConstExpr.ConstantKind.(*expr.Constant_StringValue)
	if !ok {
		return "", fmt.Errorf("expected constant string arg to %s", callExpr.Function)
	}

	return stringArg.StringValue, nil
}

// TODO: temporary solution to recusrively find target filter expr name to replace
func ExtractConstExpr(e *expr.Expr, targetName string, found bool) (string, bool) {
	identExprName := strcase.ToSnake(e.GetIdentExpr().GetName())
	if len(e.GetCallExpr().GetArgs()) == 0 && identExprName == targetName {
		return "", true
	}
	if found {
		return e.GetConstExpr().GetStringValue(), true
	}

	var strValue string
	for _, e := range e.GetCallExpr().GetArgs() {
		strValue, found = ExtractConstExpr(e, targetName, found)
		if strValue != "" && found {
			return strValue, true
		}
	}

	return "", false
}

// TODO: temporary solution to hijack and replace the `pipeline_id` filter on the fly to swap to `pipeline_uid` for query
func HijackConstExpr(e *expr.Expr, beforeExprName string, replaceExprName string, replaceExprValue string, found bool) (string, bool) {
	identExprName := strcase.ToSnake(e.GetIdentExpr().GetName())
	if len(e.GetCallExpr().GetArgs()) == 0 && identExprName == beforeExprName {
		e.GetIdentExpr().Name = replaceExprName
		return "", true
	}
	if found {
		*e = expr.Expr{
			Id: e.GetId(),
			ExprKind: &expr.Expr_ConstExpr{
				ConstExpr: &expr.Constant{
					ConstantKind: &expr.Constant_StringValue{
						StringValue: replaceExprValue,
					},
				},
			},
		}
		return e.GetConstExpr().GetStringValue(), true
	}

	var strValue string
	for _, e := range e.GetCallExpr().GetArgs() {
		strValue, found = HijackConstExpr(e, beforeExprName, replaceExprName, replaceExprValue, found)
		if strValue != "" && found {
			return strValue, true
		}
	}

	return "", false
}
