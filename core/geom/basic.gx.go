package geom

//
// float64
//

//gx:extern INFINITY
const Infinity = 0

//gx:extern PI
const Pi = 0

//gx:extern std::abs
func Abs(f float64) float64

func Sign(f float64) float64 {
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

//gx:extern std::sinf
func Sin(f float64) float64

//gx:extern std::cosf
func Cos(f float64) float64

//gx:extern std::atan2
func Atan2(y, x float64) float64

//gx:extern std::sqrtf
func Sqrt(f float64) float64

//gx:extern std::powf
func Pow(base, exp float64) float64

type Numeric interface {
	int | float32 | float64
}

func Min[T Numeric](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max[T Numeric](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

//gx:extern std::floorf
func floor(f float64) float64

func Floor(f float64) int {
	return int(floor(f))
}

//
// Vec2
//

//gx:extern rl::Vector2
type Vec2 struct {
	X float64
	Y float64
}

// Add two vectors (v1 + v)
//gx:extern rl::Vector2Add
func (v Vec2) Add(u Vec2) Vec2

// Add vector and float value
//gx:extern rl::Vector2AddValue
func (v Vec2) AddValue(f float64) Vec2

// Subtract two vectors (v1 - v)
//gx:extern rl::Vector2Subtract
func (v Vec2) Subtract(u Vec2) Vec2

// Subtract vector by float value
//gx:extern rl::Vector2SubtractValue
func (v Vec2) SubtractValue(f float64) Vec2

// Calculate vector length
//gx:extern rl::Vector2Length
func (v Vec2) Length() float64

// Calculate vector square length
//gx:extern rl::Vector2LengthSqr
func (v Vec2) LengthSqr() float64

// Calculate two vectors dot product
//gx:extern rl::Vector2DotProduct
func (v Vec2) DotProduct(u Vec2) float64

// Calculate distance between two vectors
//gx:extern rl::Vector2Distance
func (v Vec2) Distance(u Vec2) float64

// Calculate angle from two vectors in X-axis
// NOTE: Skipping because it uses degrees...
//gx:extern rl::Vector2Angle
//func (v Vec2) Angle(u Vec2) float64

// Scale vector (multiply by value)
//gx:extern rl::Vector2Scale
func (v Vec2) Scale(f float64) Vec2

// Multiply vector by vector
//gx:extern rl::Vector2Multiply
func (v Vec2) Multiply(u Vec2) Vec2

// Negate vector
//gx:extern rl::Vector2Negate
func (v Vec2) Negate() Vec2

// Divide vector by vector
//gx:extern rl::Vector2Divide
func (v Vec2) Divide(u Vec2) Vec2

// Normalize provided vector
//gx:extern rl::Vector2Normalize
func (v Vec2) Normalize() Vec2

// Calculate linear interpolation between two vectors
//gx:extern rl::Vector2Lerp
func (v Vec2) Lerp(u Vec2, amount float64)

// Calculate reflected vector to normal
//gx:extern rl::Vector2Reflect
func (v Vec2) Reflect(normal Vec2) Vec2

// Rotate Vector by float in Degrees.
// NOTE: Skipping because it uses degrees...
//gx:extern rl::Vector2Rotate
//func (v Vec2) Rotate(degs float64)

// Move Vector towards target
// NOTE: Skipping because not sure how useful...
//gx:extern rl::Vector2MoveTowards
//func (v Vec2) MoveTowards(target Vec2, maxDistance float64)

//
// Vec3
//

//gx:extern rl::Vector3
type Vec3 struct {
	X float64
	Y float64
	Z float64
}

//
// Vec4
//

//gx:extern rl::Vector4
type Vec4 struct {
	X float64
	Y float64
	Z float64
	W float64
}

//
// Quaternion
//

//gx:extern rl::Quaternion
type Quaternion Vec4

//
// Matrix
//

//gx:extern rl::Matrix
type Matrix struct {
	M0, M4, M8, M12  float64 // Matrix first row (4 components)
	M1, M5, M9, M13  float64 // Matrix second row (4 components)
	M2, M6, M10, M14 float64 // Matrix third row (4 components)
	M3, M7, M11, M15 float64 // Matrix fourth row (4 components)
}
