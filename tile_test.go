package gotile

import (
	m "github.com/murphy214/mercantile"
	"testing"

)

// test for Make_Coords_Float
func Test_Make_Line_Float(t *testing.T) {
	tv := new(vector_tile.Tile_Value)
	t := "shit"
	tv.StringValue = &t

	testcases := []struct {
			V interface{}
			Expected *vector_tile.Tile_Value
	}{	
		{
			v:"shit"
			Expected:tv,
		},
	}

	for _, tcase := range testcases {
		coords := Reflect_Value(tcase.V)
		if valmine != valtcase {
			t.Errorf("Make_Line_Geom Error, Expected %s Got %s",valtcase,valmine)
		}
	}
}
