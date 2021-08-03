package crawlers

import (
	"strconv"
)

type PropertyInfo struct {
	Address, Link, Condition, Parking, BuiltIn, NumOfFloors, Heating, AirConditioning, ToiletAndBathroom string
	HouseArea, LotArea, NumOfRooms                                                                       int
	Price, PricePerSqrMeter                                                                              float64
}

func (pi PropertyInfo) GetHeaders() []string {
	return []string{"Cím", "URL", "Állapot", "Parkolás", "Építés éve", "Emeletek száma", "Fűtés", "Légkondicionálás", "WC/Fürdő", "Alapterület", "Telekterület", "Szobák száma", "Ár", "Négyzetméter Ár"}
}

func (pi PropertyInfo) ToSlice() []string {
	return []string{pi.Address, pi.Link, pi.Condition, pi.Parking, pi.BuiltIn,
		pi.NumOfFloors, pi.Heating, pi.AirConditioning, pi.ToiletAndBathroom,
		strconv.Itoa(pi.HouseArea), strconv.Itoa(pi.LotArea), strconv.Itoa(pi.NumOfRooms),
		strconv.FormatFloat(pi.Price, 'f', 2, 64), strconv.FormatFloat(pi.PricePerSqrMeter, 'f', 2, 64)}
}

// is it too much memory to copy the list? probably not
func IsPropPresentInList(l []PropertyInfo, p PropertyInfo) bool {
	for _, prop := range l {
		if prop.HouseArea == p.HouseArea &&
			prop.LotArea == p.LotArea &&
			prop.Price == p.Price {
			return true
		}
	}
	return false
}
