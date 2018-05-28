package zrange

import "github.com/mmcloughlin/geohash"

// RadialRange uses a radius in kilometers, a latitude, and a longitude to return
// a slice of one or more ranges of keys that can be used to efficiently perform
// Geohash-based spatial queries.
//
// This method uses an algorithm that was derived from the "Search" section of this page:
// https://web.archive.org/web/20180526044934/https://github.com/yinqiwen/ardb/wiki/Spatial-Index#search
//
// RadialRange expands upon the ideas referenced above, by:
//
// • Sorting key ranges
//
// • Combining overlapping key ranges
//
// • Handling overflows resulting from bitshifting, such as when querying for: (-90, -180)
//
func RadialRange(params RadialRangeParams) HashRanges {
	return params.
		SetDefaults().
		FindNeighborsWithRadius().
		SortMinAsc().
		CombineRanges()
}

// RadialRangeParams defaults to expecting 64-bit Geohash-encoded keys.
type RadialRangeParams struct {
	BitsOfPrecision uint
	Radius,
	Latitude,
	Longitude float64
}

// SetDefaults sets the default values for the RadialRangeParams type.
func (params RadialRangeParams) SetDefaults() RadialRangeParams {
	if params.BitsOfPrecision == 0 {
		params.BitsOfPrecision = 64
	}
	return params
}

func (params RadialRangeParams) radiusToBits() uint {
	const initialSignificantBits = 2

	for i := len(radiusToBits) - 1; i > 0; i-- {
		if params.Radius < radiusToBits[i] {
			return uint(i*2 + initialSignificantBits)
		}
	}

	return uint(initialSignificantBits)
}

// FindNeighborsWithRadius uses the radius and coordinates to find neighboring
// hash ranges.
func (params RadialRangeParams) FindNeighborsWithRadius() HashRanges {
	rangeBits := params.radiusToBits()

	queryPoint := geohash.EncodeIntWithPrecision(
		params.Latitude,
		params.Longitude,
		rangeBits,
	)

	neighborList := neighbors(geohash.NeighborsIntWithPrecision(queryPoint, rangeBits))
	neighborList = append(neighborList, queryPoint)

	rangeBitsDiff := params.BitsOfPrecision - rangeBits
	return neighborList.shiftIntoRanges(rangeBitsDiff)
}