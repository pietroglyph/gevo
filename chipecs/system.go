package chipecs

import (
	"log"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

// Remove is called when an entity is removed from the world
// so that this system knows that it's gone
func (ps *PhysicsSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range ps.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			ps.Space.RemoveBody(e.Shape.Body)
			break
		}
	}
	if delete >= 0 {
		ps.entities = append(ps.entities[:delete], ps.entities[delete+1:]...)
	}
}

// New is called when the PhysicsSystem is added to a scene
func (ps *PhysicsSystem) New(*ecs.World) {
	log.Println("PhyiscsSystem was added to the scene.")
	// Set up the spacce
	ps.Space = chipmunk.NewSpace()
	ps.Space.Gravity = vect.Vect{X: 0, Y: 0}
}

// Update is called once every frame
func (ps *PhysicsSystem) Update(dt float32) {
	for _, e := range ps.entities {
		pos := e.PhysicsComponent.Shape.Body.Position()
		e.Position = engo.Point{X: float32(pos.X), Y: float32(pos.Y)}
		e.Rotation = 0 // Engo rotates around the origin, not the center, so we don't want to set this
	}
	ps.Space.Step(vect.Float(dt))
}

// Add adds a basic entity that has a physics and space component to the physics system
func (ps *PhysicsSystem) Add(basic *ecs.BasicEntity, physics *PhysicsComponent, space *common.SpaceComponent) {
	ps.entities = append(ps.entities, physicsEntity{basic, physics, space})
	ps.Space.AddBody(physics.Shape.Body)
}
