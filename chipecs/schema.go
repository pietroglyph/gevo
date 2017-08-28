package chipecs

import (
	"engo.io/ecs"
	"engo.io/engo/common"
	"github.com/vova616/chipmunk"
)

// PhysicsSystem implements a basic system to manage the Chimunk physics engine
// This type implements the engo.System interface
type PhysicsSystem struct {
	Space *chipmunk.Space

	entities []physicsEntity
	world    *ecs.World
}

// PhysicsComponent holds physics data
type PhysicsComponent struct {
	Body *chipmunk.Body
}

type physicsEntity struct {
	*ecs.BasicEntity
	*PhysicsComponent
	*common.SpaceComponent
}
