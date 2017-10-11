package config

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform/tfdiags"

	hcl2 "github.com/hashicorp/hcl2/hcl"
	hcl2syntax "github.com/hashicorp/hcl2/hcl/hclsyntax"
	hcl2dec "github.com/hashicorp/hcl2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

func TestDetectVariablesHCL2(t *testing.T) {
	tests := []struct {
		Expr      string
		Want      []InterpolatedVariable
		DiagCount int
	}{
		{
			`true`,
			nil,
			0,
		},

		{
			`count.index`,
			[]InterpolatedVariable{
				&CountVariable{
					Type: CountValueIndex,
					key:  "count.index",
					varRange: makeVarRange(tfdiags.SourceRange{
						Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
						End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
					}),
				},
			},
			0,
		},
		{
			`count.baz`,
			[]InterpolatedVariable{
				&CountVariable{
					Type: CountValueInvalid,
					key:  "count.baz",
					varRange: makeVarRange(tfdiags.SourceRange{
						Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
						End:   tfdiags.SourcePos{Line: 1, Column: 10, Byte: 9},
					}),
				},
			},
			1, // invalid "count" attribute
		},
		{
			`count`,
			nil,
			1, // missing "count" attribute
		},

		{
			`var.foo`,
			[]InterpolatedVariable{
				&UserVariable{
					Name: "foo",
					key:  "var.foo",
					varRange: makeVarRange(tfdiags.SourceRange{
						Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
						End:   tfdiags.SourcePos{Line: 1, Column: 8, Byte: 7},
					}),
				},
			},
			0,
		},
		{
			`var.foo.bar`,
			[]InterpolatedVariable{
				&UserVariable{
					Name: "foo",
					key:  "var.foo",
					varRange: makeVarRange(tfdiags.SourceRange{
						Start: tfdiags.SourcePos{Line: 1, Column: 1, Byte: 0},
						End:   tfdiags.SourcePos{Line: 1, Column: 12, Byte: 11},
					}),
				},
			},
			0,
		},
		{
			`var`,
			nil,
			1, // missing "var" attribute
		},
	}

	for _, test := range tests {
		t.Run(test.Expr, func(t *testing.T) {
			var diags tfdiags.Diagnostics
			expr, parseDiags := hcl2syntax.ParseExpression([]byte(test.Expr), "", hcl2.Pos{Line: 1, Column: 1})
			diags = diags.Append(parseDiags)
			body := hcl2SingleAttrBody{
				Name: "value",
				Expr: expr,
			}
			spec := &hcl2dec.AttrSpec{
				Name: "value",
				Type: cty.DynamicPseudoType,
			}

			got, varDiags := DetectVariablesHCL2(body, spec)
			diags = diags.Append(varDiags)

			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					desc := diag.Description()
					t.Logf("- %s: %s", desc.Summary, desc.Detail)
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
		})
	}
}
