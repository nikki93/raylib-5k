package entity

//
// Entities
//

//gx:extern Entity
type Entity struct{}

//gx:extern nullEntity
var NullEntity = Entity{}

//gx:extern createEntity
func CreateEntity(...interface{}) Entity

//gx:extern destroyEntity
func DestroyEntity(ent Entity)

//gx:extern exists
func EntityExists(ent Entity) bool

//gx:extern numEntities
func NumEntities() int

//gx:extern uint32_t
func EntityToInt(ent Entity) int

//gx:extern Entity
func IntToEntity(i int) Entity

//
// Components
//

//gx:extern has
func HasComponent[T any](ent Entity) (result bool) { return }

//gx:extern getPtr
func GetComponent[T any](ent Entity) (result *T) { return }

//gx:extern addPtr
func AddComponent[T any](ent Entity, value T) (result *T) { return }

//gx:extern remove
func RemoveComponent[T any](ent Entity) { return }

//gx:extern each
func Each(...interface{}) { return }

//gx:extern clear
func ClearComponent[T any]() { return }

//gx:extern sortPtr
func SortComponent[T any](compare func(a, b *T) bool) { return }

//
// Meta
//

//gx:extern Behavior
type Behavior struct{}
