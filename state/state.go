package state

type state struct {
	mesh *Mesh
}

func newState() state {
	return state{
		mesh: NewMesh(),
	}
}
