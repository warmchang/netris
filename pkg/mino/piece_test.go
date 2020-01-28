package mino

type PieceTestData struct {
	R0 string
	RR string
	R2 string
	RL string
}

var pieceTestData = []*PieceTestData{
	{Monomino, Monomino, Monomino, Monomino},
	{Domino, "(0,-1),(0,0)", Domino, "(0,-1),(0,0)"},
}

// TODO Resolve CCW rotation resulting in different coords than CW rotation before completing test
/*
func TestPiece(t *testing.T) {
	t.Parallel()

	return

	for i, d := range pieceTestData {
		m := NewMino(d.R0)
		if m == nil || m.String() != d.R0 {
			t.Errorf("failed to create mino %d %s: got %s", i, d.R0, m)
		}

		p := NewPiece(m, Point{0, 0})
		if p == nil || p.Mino.String() != d.R0 {
			t.Errorf("failed to create piece %d %s: got %+v", i, d.R0, p)
		}

		for direction := 0; direction <= 1; direction++ {
			for rotations := 0; rotations < 8; rotations++ {
				p.Mino = p.Rotate(1, direction)
				expected := ""
				switch rotations % 4 {
				case 0:
					if direction == 0 {
						expected = d.RR
					} else {
						expected = d.RL
					}
				case 1:
					expected = d.R2
				case 2:
					if direction == 0 {
						expected = d.RL
					} else {
						expected = d.RR
					}
				case 3:
					expected = d.R0
				default:
					t.Errorf("unexpected rotation count %d", rotations)
				}
				if p.Mino.String() != expected {
					t.Errorf("failed to rotate piece %d - direction %d - rotation %d - expected %s: got %s", i, direction, rotations, expected, p.Mino)
				}
			}
		}

		log.Println(p)
	}
}
*/
