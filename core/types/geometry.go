package types

import (
	"math"

	"github.com/shopspring/decimal"
)

type Span struct {
	Lower, Upper float64
}

func (t Span) Contains(raw float64) bool {
	return raw >= t.Lower && raw <= t.Upper
}

type Square struct {
	X0, X1 float64
	Y0, Y1 float64
}

type Point struct {
	X, Y float64
}

func (t Point) InPolygon(raw Polygon) bool {
	return PointInPolygon(t, raw)
}

func (t Point) Slope(p Point) float64 {
	return ComputePointSlope(t, p)
}

func (t Point) DistanceFrom(p Point) float64 {
	return ComputeDistance(t.X, t.Y, p.X, p.Y)
}

func (t Point) RadDistance(p Point) float64 {
	return SphericalDistance(t.Y, t.X, p.Y, p.X)
}

type Line [2]Point

func (t Line) Slope() float64 {
	return ComputeSlope(t[0].Y, t[0].X, t[1].Y, t[1].X)
}

type Polygon []Point

func (t Polygon) PointIn(p Point) bool {
	return PointInPolygon(p, t)
}

func (t Polygon) GeoPointIn(p GeoPoint) bool {
	return PointInPolygon(p.Point(), t)
}

func (t Polygon) Check(x, y float64) bool {
	return PointInPolygon(Point{X: x, Y: y}, t)
}

func (t Polygon) Nearest(p Point) *Point {
	l := len(t)
	if l == 0 {
		return nil
	}

	n := 0
	minDistance := ComputePointDistance(p, t[n])

	for i := 1; i < l; i++ {
		distance := ComputePointDistance(p, t[i])
		if distance < minDistance {
			minDistance = distance
			n = i
		}
	}

	res := &Point{X: t[n].X, Y: t[n].Y}

	return res
}

type PolygonGroup map[string]Polygon

func (t PolygonGroup) Find(x, y float64) (string, bool) {
	for k, v := range t {
		if PointInPolygon(Point{X: x, Y: y}, v) {
			return k, true
		}
	}

	return "", false
}

func PointInPolygon(p Point, poly []Point) bool {
	n := len(poly)
	inside := false
	for i, j := 0, n-1; i < n; i++ {
		if (poly[i].Y > p.Y) != (poly[j].Y > p.Y) &&
			p.X < (poly[j].X-poly[i].X)*(p.Y-poly[i].Y)/(poly[j].Y-poly[i].Y)+poly[i].X {
			inside = !inside
		}
		j = i
	}

	return inside
}

func ComputeSlope(h1, v1, h2, v2 float64) float64 {
	if h1 == h2 {
		return 0
	}

	return (v1 - v2) / (h1 - h2)
}

func ComputePointSlope(p1, p2 Point) float64 {
	return ComputeSlope(p1.Y, p1.X, p2.Y, p2.X)
}

func ComputePointDistance(p1, p2 Point) float64 {
	return ComputeDistance(p1.X, p1.Y, p2.X, p2.Y)
}

func ComputeDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2))
}

func ComputeTimeDataSlope(t1 int64, v1 float64, t2 int64, v2 float64) float64 {
	if t1 == t2 {
		return 0
	}

	return (v1 - v2) / float64(t1-t2)
}

func FittingTwoFloat(f1, f2, all, now float64) float64 {
	if all == 0 {
		return f1
	}

	slope := (f2 - f1) / all
	return f1 + slope*now
}

func InvalidLatitude(raw float64) bool {
	return raw >= -180 && raw <= 180
}

func InvalidLongitude(raw float64) bool {
	return raw >= -90 && raw <= 90
}

func SphericalDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 将经纬度转换为弧度
	lat1 = Degree2Radian(lat1)
	lon1 = Degree2Radian(lon1)
	lat2 = Degree2Radian(lat2)
	lon2 = Degree2Radian(lon2)

	// 使用球面三角法计算距离
	deltaLat := lat2 - lat1
	deltaLon := lon2 - lon1

	d := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(d), math.Sqrt(1-d))

	distance := EarthRadius * c

	return distance
}

func Bearing(lat1, lon1, lat2, lon2 float64) float64 {
	// 将经纬度转换为弧度
	lat1 = Degree2Radian(lat1)
	lat2 = Degree2Radian(lat2)

	deltaLon := Degree2Radian(lon2 - lon1)

	numerator := math.Sin(deltaLon) * math.Cos(lat2)
	denominator := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)

	azimuthRad := math.Atan2(numerator, denominator)
	res := Radian2Degree(azimuthRad)

	return NormalizeDegree(res)
}

type FlatPoint struct {
	Longitude float64 `db:"longitude" json:"longitude" validate:"gte=-180,lte=180"`
	Latitude  float64 `db:"latitude" json:"latitude" validate:"gte=-90,lte=90"`
}

type GeoPoint struct {
	Longitude float64 `db:"longitude" json:"longitude" validate:"gte=-180,lte=180"`
	Latitude  float64 `db:"latitude" json:"latitude" validate:"gte=-90,lte=90"`
	Height    float64 `db:"height" json:"height"`
}

func NewGeoPointFromFloat(raw ...float64) GeoPoint {
	if len(raw) < 3 {
		return GeoPoint{}
	}

	res := GeoPoint{
		Longitude: raw[0],
		Latitude:  raw[1],
		Height:    raw[2],
	}

	return res
}

func NewGeoPointFromMap(raw map[string]string) GeoPoint {
	longitude, _ := ParseFloat64FromStr(raw["longitude"])
	latitude, _ := ParseFloat64FromStr(raw["latitude"])
	height, _ := ParseFloat64FromStr(raw["height"])

	res := GeoPoint{
		Longitude: longitude,
		Latitude:  latitude,
		Height:    height,
	}

	return res
}

func (t GeoPoint) Invalid() bool {
	return t.Longitude == 0 && t.Latitude == 0
}

func (t GeoPoint) Values() []interface{} {
	values := []interface{}{
		"longitude", t.Longitude,
		"latitude", t.Latitude,
		"height", t.Height,
	}

	return values
}

func (t GeoPoint) Polarize() GeoPoint {
	if t.Longitude > 180 {
		t.Longitude = 180
	} else if t.Longitude < -180 {
		t.Longitude = -180
	}

	if t.Latitude > 90 {
		t.Latitude = 90
	} else if t.Latitude < -90 {
		t.Latitude = -90
	}

	return t
}

func (t GeoPoint) Normalize() GeoPoint {
	t.Longitude = math.Mod(t.Longitude, 360.0)
	if t.Longitude > 180 {
		t.Longitude -= 360
	}

	t.Latitude = math.Mod(t.Latitude, 180.0)
	if t.Latitude > 90 {
		t.Latitude -= 180
	}

	return t
}

func (t GeoPoint) Point() Point {
	return Point{X: t.Longitude, Y: t.Latitude}
}

func (t GeoPoint) DistanceFrom(p GeoPoint) float64 {
	return SphericalDistance(t.Longitude, t.Latitude, p.Longitude, p.Latitude)
}

func (t GeoPoint) SphericalDistanceFrom(p GeoPoint) float64 {
	return SphericalDistance(t.Latitude, t.Longitude, p.Latitude, p.Longitude)
}

func (t GeoPoint) Bearing(p GeoPoint) float64 {
	return Bearing(t.Latitude, t.Longitude, p.Latitude, p.Longitude)
}

func (t GeoPoint) Radius(raw float64) (e, s, w, n GeoPoint) {
	r := raw / EarthRadius

	c := t.Normalize()
	deltaLong := math.Tan(r) / math.Cos(Degree2Radian(c.Latitude))
	deltaLong = NormalizeRadian2Degree(deltaLong)

	deltaLag := Radian2Degree(r)

	//east
	e = GeoPoint{
		Longitude: c.Longitude + deltaLong,
		Latitude:  c.Latitude,
	}
	e = e.Normalize()

	//south
	s = GeoPoint{
		Longitude: c.Longitude,
		Latitude:  c.Latitude - deltaLag,
	}
	s = s.Normalize()

	//west
	w = GeoPoint{
		Longitude: c.Longitude - deltaLong,
		Latitude:  c.Latitude,
	}
	w = w.Normalize()

	//north
	n = GeoPoint{
		Longitude: c.Longitude,
		Latitude:  c.Latitude + deltaLag,
	}
	n = n.Normalize()

	return
}

func (t GeoPoint) Cartesian() CartesianPoint {
	lonRad := Degree2Radian(t.Longitude)
	latRad := Degree2Radian(t.Latitude)

	cosLat := math.Cos(latRad)
	sinLat := math.Sin(latRad)
	cosLon := math.Cos(lonRad)
	sinLon := math.Sin(lonRad)

	n := EarthEquatorialRadius / math.Sqrt(1-e2*sinLat*sinLat)

	h := n + t.Height

	res := CartesianPoint{
		X: h * cosLat * cosLon,
		Y: h * cosLat * sinLon,
		Z: ((1-e2)*n + t.Height) * sinLat,
	}.Truncate()

	return res
}

func (t GeoPoint) PanCartesian() CartesianPoint {
	lonRad := Degree2Radian(t.Longitude)
	latRad := Degree2Radian(t.Latitude)

	cosLat := math.Cos(latRad)
	sinLat := math.Sin(latRad)
	cosLon := math.Cos(lonRad)
	sinLon := math.Sin(lonRad)

	res := CartesianPoint{
		X: cosLat * cosLon * EarthRadius,
		Y: cosLat * sinLon * EarthRadius,
		Z: sinLat * EarthRadius,
	}.Truncate()

	return res
}

func Degree2Radian(degree float64) (radian float64) {
	return degree * deg2rad
}

func Radian2Degree(radian float64) (degree float64) {
	return radian / deg2rad
}

func Degree2RadianD(degree float64) (radian decimal.Decimal) {
	return NewDecimal(degree).Mul(deg2radD)
}

func Radian2DegreeD(radian float64) (degree decimal.Decimal) {
	return NewDecimal(radian).Div(deg2radD)
}

func NormalizeRadians(radian float64) float64 { return math.Mod(radian, 2*math.Pi) }

func NormalizeDegree(degree float64) float64 {
	return math.Mod(degree+360.0, 360.0)
}

func NormalizeRadian2Degree(radian float64) float64 {
	return NormalizeDegree(Radian2Degree(radian))
}

type CartesianPoint struct {
	X, Y, Z float64
}

func (t CartesianPoint) IsZero() bool {
	return IsZeroFloat(t.X) && IsZeroFloat(t.Y) && IsZeroFloat(t.Z)
}

func (t CartesianPoint) Truncate() CartesianPoint {
	Truncate(
		&t.X,
		&t.Y,
		&t.Z,
	)

	return t
}

func (t CartesianPoint) Add(p CartesianPoint) CartesianPoint {
	return CartesianPoint{X: t.X + p.X, Y: t.Y + p.Y, Z: t.Z + p.Z}
}

func (t CartesianPoint) Sub(p CartesianPoint) CartesianPoint {
	return CartesianPoint{X: t.X - p.X, Y: t.Y - p.Y, Z: t.Z - p.Z}
}

func (t CartesianPoint) Scale(scalar float64) CartesianPoint {
	return CartesianPoint{X: t.X * scalar, Y: t.Y * scalar, Z: t.Z * scalar}
}

func (t CartesianPoint) DotProduct(p CartesianPoint) float64 {
	return t.X*p.X + t.Y*p.Y + t.Z*p.Z
}

func (t CartesianPoint) CrossProduct(p CartesianPoint) CartesianPoint {
	return CartesianPoint{X: t.Y*p.Z - t.Z*p.Y, Y: t.Z*p.X - t.X*p.Z, Z: t.X*p.Y - t.Y*p.X}
}

func (t CartesianPoint) CrossProductNum(raw CartesianPoint) float64 {
	res := math.Sqrt(math.Pow(t.Y*raw.Z-t.Z*raw.Y, 2) +
		math.Pow(t.Z*raw.X-t.X*raw.Z, 2) +
		math.Pow(t.X*raw.Y-t.Y*raw.X, 2))

	return res
}

func (t CartesianPoint) GeoPoint() GeoPoint {
	return CartesianToGeoPoint(t)
}

func (t CartesianPoint) ParallelTo(p CartesianPoint) bool {
	return t.CrossProduct(p).IsZero()
}

func (t CartesianPoint) ParallelToX() bool {
	return !IsZeroFloat(t.X) && IsZeroFloat(t.Y) && IsZeroFloat(t.Z)
}

func (t CartesianPoint) ParallelToY() bool {
	return !IsZeroFloat(t.Y) && IsZeroFloat(t.X) && IsZeroFloat(t.Z)
}

func (t CartesianPoint) ParallelToZ() bool {
	return !IsZeroFloat(t.Z) && IsZeroFloat(t.X) && IsZeroFloat(t.Y)
}

func CartesianToGeoPoint(cp CartesianPoint) GeoPoint {
	r := math.Sqrt(cp.DotProduct(cp))

	res := GeoPoint{
		Longitude: Radian2Degree(math.Atan2(cp.Y, cp.X)),
		Latitude:  Radian2Degree(math.Asin(cp.Z / r)),
		Height:    r - EarthRadius,
	}.Normalize()

	return res
}

type Inclination struct {
	PanAngle  float64 `db:"pan_angle" json:"pan_angle"`   //水平角度
	TiltAngle float64 `db:"tilt_angle" json:"tilt_angle"` //垂直角度
}

func (t Inclination) PanSpan(delta float64) Span {
	return Span{Lower: t.PanAngle - delta, Upper: t.PanAngle + delta}
}

func (t Inclination) TiltSpan(delta float64) Span {
	return Span{Lower: t.TiltAngle - delta, Upper: t.TiltAngle + delta}
}

func (t Inclination) Cartesian() CartesianPoint {
	panRad := Degree2Radian(t.PanAngle)
	tiltRad := Degree2Radian(t.TiltAngle)

	cosPan := math.Cos(panRad)
	sinPan := math.Sin(panRad)
	cosTilt := math.Cos(tiltRad)
	sinTilt := math.Sin(tiltRad)

	x := cosPan * cosTilt
	y := sinPan * cosTilt
	z := sinTilt

	return CartesianPoint{X: x, Y: y, Z: z}.Truncate()
}

type ObservationStation struct {
	GeoPoint
	Inclination
}

func (t ObservationStation) Direction() CartesianPoint {
	lat := Degree2Radian(t.Latitude)
	lon := Degree2Radian(t.Longitude)

	res := CartesianPoint{
		X: math.Cos(lat) * math.Cos(lon),
		Y: math.Cos(lat) * math.Sin(lon),
		Z: math.Sin(lat),
	}.Truncate()

	return res
}

func (t ObservationStation) Ray() ObserverRay {
	res := ObserverRay{
		GeoPoint:       t.GeoPoint,
		CartesianPoint: t.GeoPoint.Cartesian(),
		Direction:      t.Inclination.Cartesian(),
	}

	return res
}

func (t ObservationStation) FlatRay() ObserverRay {
	pan := Degree2Radian(t.PanAngle)

	res := ObserverRay{
		GeoPoint:       t.GeoPoint,
		CartesianPoint: t.GeoPoint.PanCartesian(),
		Direction: CartesianPoint{
			X: math.Sin(pan),
			Y: math.Cos(pan),
		},
	}

	return res
}

func (t ObservationStation) FlatCrossPoint(raw ObservationStation) (p GeoPoint, ok bool) {
	return FlatCrossPoint(t, raw)
}

func (t ObservationStation) FitHeight(p *GeoPoint) {
	if t.TiltAngle == 0 {
		return
	}

	dis := t.DistanceFrom(*p)
	p.Height = t.Height - dis*math.Tan(Degree2Radian(t.TiltAngle))
}

type ObserverRay struct {
	GeoPoint
	CartesianPoint
	Direction CartesianPoint
}

func (t ObserverRay) Diff(raw ObserverRay) CartesianPoint {
	return t.Direction.Sub(raw.Direction)
}

func (t ObserverRay) Fit(dis float64) CartesianPoint {
	res := CartesianPoint{
		X: t.X + t.Direction.X*dis,
		Y: t.Y + t.Direction.Y*dis,
		Z: t.Z + t.Direction.Z*dis,
	}.Truncate()

	return res
}

func (t ObserverRay) FitFlat(dis float64) CartesianPoint {
	res := CartesianPoint{
		X: t.X + t.Direction.X*dis,
		Y: t.Y + t.Direction.Y*dis,
	}.Truncate()

	return res
}

func (t ObserverRay) CrossProduct(raw ObserverRay) CartesianPoint {
	return t.Direction.CrossProduct(raw.Direction)
}

func (t ObserverRay) CrossProductNum(raw ObserverRay) float64 {
	return t.Direction.CrossProductNum(raw.Direction)
}

func (t ObserverRay) IntersectFlat(raw ObserverRay) (p GeoPoint, ok bool) {
	det := t.Direction.X*raw.Direction.Y - raw.Direction.X*t.Direction.Y
	if IsZeroFloat(det) {
		return
	}

	c := (raw.Direction.X*(raw.CartesianPoint.Y-t.CartesianPoint.Y) -
		raw.Direction.Y*(raw.CartesianPoint.X-t.CartesianPoint.X)) / det
	s := (t.Direction.X*(raw.CartesianPoint.Y-t.CartesianPoint.Y) -
		t.Direction.Y*(raw.CartesianPoint.X-t.CartesianPoint.X)) / det

	ok = c >= 0 && s >= 0
	if !ok {
		return
	}

	p = t.FitFlat(c).GeoPoint()

	return
}

func (t ObserverRay) ParallelTo(raw ObserverRay) bool {
	return t.Direction.ParallelTo(raw.Direction)
}

func FlatCrossPoint(stationA, stationB ObservationStation) (p GeoPoint, ok bool) {
	rayA := stationA.FlatRay()
	rayB := stationB.FlatRay()

	crossProduct := rayA.CrossProduct(rayB)
	crossProductMag := crossProduct.DotProduct(crossProduct)
	crossProductMag = math.Sqrt(crossProductMag)

	//正交
	if IsZeroFloat(crossProductMag - 1) {
		p, ok = rayA.IntersectFlat(rayB)
		if ok {
			finalPanRayCheck(&p, rayA, rayB)
			return
		}

		return
	}

	//平行
	if rayA.ParallelTo(rayB) {
		return
	}

	t := (rayA.Direction.X*crossProduct.Y - rayA.Direction.Y*crossProduct.X -
		rayB.Direction.X*crossProduct.Y + rayB.Direction.Y*crossProduct.X) /
		(crossProductMag * crossProductMag)

	p = rayA.FitFlat(t).GeoPoint()

	stationA.FitHeight(&p)

	finalPanRayCheck(&p, rayA, rayB)

	return p, true
}

func finalPanRayCheck(p *GeoPoint, ray ...ObserverRay) {
	for _, v := range ray {
		if v.Direction.ParallelToX() {
			p.Longitude = v.Longitude
		}

		if v.Direction.ParallelToY() {
			p.Latitude = v.Latitude
		}
	}
}
