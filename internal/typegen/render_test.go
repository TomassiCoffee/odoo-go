package typegen

import "testing"

func TestToPascalCase(t *testing.T) {
	got := ToPascalCase("pos.order.line")
	if got != "PosOrderLine" {
		t.Fatalf("got %q", got)
	}
}

func TestRender(t *testing.T) {
	_, err := Render(MetadataCache{Models: []NormalizedModel{{
		Name: "pos.order",
		Fields: []NormalizedField{
			{Name: "id", Type: FieldInteger},
			{Name: "pos_reference", Type: FieldChar},
		},
	}}}, RenderOptions{PackageName: "odoomodels", OdooImportPath: "github.com/tomassicoffee/odoo-go"})
	if err != nil {
		t.Fatal(err)
	}
}
